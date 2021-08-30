package remote_cluster

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	addonsv1alpha1 "github.com/openshift/addon-operator/apis/addons/v1alpha1"
	"github.com/openshift/addon-operator/internal/cache"
	"github.com/openshift/addon-operator/internal/reconcile"
)

const (
	remoteClusterFinalizer = "fleet.ensure-stack.org/cleanup"
)

type RemoteClusterReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	Recorder    record.EventRecorder
	ClientCache cache.ClientCache
}

func (r *RemoteClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(
			&addonsv1alpha1.RemoteCluster{},
			// TODO: Why isn't that the default for all reconcilers?
			builder.WithPredicates(predicate.GenerationChangedPredicate{}),
		).
		Owns(&corev1.Namespace{}).
		Complete(r)
}

// RemoteClusterReconciler/Controller entrypoint
func (r *RemoteClusterReconciler) Reconcile(
	ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	log := r.Log.WithValues("remote-cluster", req.NamespacedName.String())

	rc := &addonsv1alpha1.RemoteCluster{}
	if err := r.Get(ctx, req.NamespacedName, rc); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle Deletion
	if !rc.DeletionTimestamp.IsZero() {
		r.ClientCache.Free(rc.Status.LocalNamespace)

		controllerutil.RemoveFinalizer(rc, remoteClusterFinalizer)
		if err := r.Update(ctx, rc); err != nil {
			return ctrl.Result{}, fmt.Errorf("removing finalizer to RemoteCluster: %w", err)
		}
		return ctrl.Result{}, nil
	}

	// Add Finalizer to cleanup cached clients
	if !controllerutil.ContainsFinalizer(rc, remoteClusterFinalizer) {
		controllerutil.AddFinalizer(rc, remoteClusterFinalizer)
		if err := r.Update(ctx, rc); err != nil {
			return ctrl.Result{}, fmt.Errorf("adding finalizer to RemoteCluster: %w", err)
		}
	}

	// Reconcile a Namespace for the Cluster to contain Remote Objects for it.
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-" + rc.Name,
		},
	}
	if err := controllerutil.SetControllerReference(rc, ns, r.Scheme); err != nil {
		return ctrl.Result{}, fmt.Errorf("setting controller reference on namespace: %w", err)
	}
	if _, err := reconcile.Namespace(ctx, log, r.Client, ns); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconciling Cluster Namespace: %w", err)
	}
	rc.Status.LocalNamespace = ns.Name

	// Get or Create clients if needed.
	host, remoteClient, remoteDiscoveryClient, ok := r.ClientCache.Get( //nolint:ineffassign,staticcheck
		rc.Status.LocalNamespace)
	if !ok {
		// Create Clients
		var err error
		host, remoteClient, remoteDiscoveryClient, err = r.createRemoteClients(ctx, rc, log)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("creating remote clients: %w", err)
		}
		if remoteClient == nil || remoteDiscoveryClient == nil {
			// Could not create client
			return ctrl.Result{
				// Make sure we re-check periodically
				RequeueAfter: rc.Spec.ResyncInterval.Duration,
			}, nil
		}

		r.ClientCache.Set(
			rc.Status.LocalNamespace, host, remoteClient, remoteDiscoveryClient)
	}
	rc.Status.Remote.APIServer = host

	// Check connection.
	if err := r.checkRemoteVersion(ctx, rc, remoteDiscoveryClient); err != nil {
		return ctrl.Result{}, fmt.Errorf("could not check RemoteCluster version: %w", err)
	}

	return ctrl.Result{
		// Make sure we re-check periodically
		RequeueAfter: rc.Spec.ResyncInterval.Duration,
	}, nil
}

func (r *RemoteClusterReconciler) checkRemoteVersion(
	ctx context.Context, rc *addonsv1alpha1.RemoteCluster, remoteDiscoveryClient discovery.DiscoveryInterface,
) error {
	rc.Status.LastHeartbeatTime = metav1.Now()

	version, err := remoteDiscoveryClient.ServerVersion()
	if err != nil {
		meta.SetStatusCondition(&rc.Status.Conditions, metav1.Condition{
			Type:               addonsv1alpha1.RemoteClusterReachable,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: rc.Generation,
			Reason:             "APIError",
			Message:            "Remote cluster API is not responding.",
		})
		rc.Status.UpdatePhase()
		return r.Status().Update(ctx, rc)
	}

	rc.Status.Remote.Version = version.String()
	meta.SetStatusCondition(&rc.Status.Conditions, metav1.Condition{
		Type:               addonsv1alpha1.RemoteClusterReachable,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: rc.Generation,
		Reason:             "APIResponding",
		Message:            "Remote cluster API is responding.",
	})
	rc.Status.UpdatePhase()
	return r.Status().Update(ctx, rc)
}

func (r *RemoteClusterReconciler) createRemoteClients(
	ctx context.Context, rc *addonsv1alpha1.RemoteCluster, log logr.Logger,
) (
	host string,
	remoteClient client.Client,
	remoteDiscoveryClient discovery.DiscoveryInterface,
	err error,
) {
	// Lookup Kubeconfig Secret
	secretKey := client.ObjectKey{
		Name:      rc.Spec.KubeconfigSecret.Name,
		Namespace: rc.Spec.KubeconfigSecret.Namespace,
	}
	kubeconfigSecret := &corev1.Secret{}
	if err := r.Get(ctx, secretKey, kubeconfigSecret); errors.IsNotFound(err) {
		log.Info("missing kubeconfig secret", "secret", secretKey)
		r.Recorder.Eventf(rc, corev1.EventTypeWarning, "InvalidConfig", "missing kubeconfig secret %q", secretKey)
		return "", nil, nil, nil
	} else if err != nil {
		return "", nil, nil, fmt.Errorf("getting Kubeconfig secret: %w", err)
	}

	// Kubeconfig sanity check
	if kubeconfigSecret.Type != addonsv1alpha1.SecretTypeKubeconfig {
		log.Info("invalid secret type", "secret", secretKey)
		r.Recorder.Eventf(rc, corev1.EventTypeWarning, "InvalidConfig",
			"invalid secret type %q, want %q", kubeconfigSecret.Type, addonsv1alpha1.SecretTypeKubeconfig)
		return "", nil, nil, nil
	}
	kubeconfig, ok := kubeconfigSecret.Data[addonsv1alpha1.SecretKubeconfigKey]
	if !ok {
		log.Info(fmt.Sprintf("missing %q key in kubeconfig secret", addonsv1alpha1.SecretKubeconfigKey), "secret", secretKey)
		r.Recorder.Eventf(rc, corev1.EventTypeWarning, "InvalidConfig",
			"missing %q key in kubeconfig secret", addonsv1alpha1.SecretKubeconfigKey)
		return "", nil, nil, nil
	}

	// Create Clients
	remoteRESTConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		log.Error(err, "invalid kubeconfig", "secret", secretKey)
		r.Recorder.Eventf(rc, corev1.EventTypeWarning, "InvalidConfig",
			"invalid kubeconfig: %w", err)
		return "", nil, nil, nil
	}
	remoteDiscoveryClient, err = discovery.NewDiscoveryClientForConfig(remoteRESTConfig)
	if err != nil {
		return "", nil, nil, fmt.Errorf("creating RemoteCluster discovery client: %w", err)
	}
	remoteClient, err = client.New(remoteRESTConfig, client.Options{})
	if err != nil {
		return "", nil, nil, fmt.Errorf("creating RemoteCluster client: %w", err)
	}

	return remoteRESTConfig.Host, remoteClient, remoteDiscoveryClient, nil
}
