package remote_object

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/yaml"

	addonsv1alpha1 "github.com/openshift/addon-operator/apis/addons/v1alpha1"
	"github.com/openshift/addon-operator/internal/cache"
)

const (
	defaultObjectResyncInterval = 1 * time.Minute
)

type RemoteObjectReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	Recorder    record.EventRecorder
	ClientCache cache.ClientCache
}

func (r *RemoteObjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(
			&addonsv1alpha1.RemoteObject{},
			builder.WithPredicates(predicate.GenerationChangedPredicate{}),
		).
		Complete(r)
}

// RemoteObjectReconciler/Controller entrypoint
func (r *RemoteObjectReconciler) Reconcile(
	ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	log := r.Log.WithValues("remote-object", req.NamespacedName.String())

	ro := &addonsv1alpha1.RemoteObject{}
	if err := r.Get(ctx, req.NamespacedName, ro); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	resyncInterval := defaultObjectResyncInterval

	if ro.Spec.PriorityClassName != "" {
		// PriorityClass Lookup
		pc := &addonsv1alpha1.RemoteObjectPriorityClass{}
		if err := r.Get(ctx, client.ObjectKey{
			Name: ro.Spec.PriorityClassName,
		}, pc); err != nil && !errors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("lookup RemoteObjectPriorityClass: %w", err)
		} else if errors.IsNotFound(err) {
			r.Recorder.Eventf(ro, corev1.EventTypeWarning, "MissingPriorityClass",
				"RemoteObjectPriorityClass %q not found", ro.Spec.PriorityClassName)
			return ctrl.Result{
				RequeueAfter: defaultObjectResyncInterval,
			}, nil
		}

		resyncInterval = pc.ResyncInterval.Duration
	}

	_, remoteClient, _, ok := r.ClientCache.Get(ro.Namespace)
	if !ok {
		r.Recorder.Event(ro, corev1.EventTypeNormal, "WaitingForClient",
			"Waiting for RemoteCluster clients to become available")
		return ctrl.Result{
			RequeueAfter: resyncInterval,
		}, nil
	}

	if err := r.syncObject(ctx, log, ro, remoteClient); err != nil {
		return ctrl.Result{}, fmt.Errorf(
			"syncing RemoteObject %s: %w", client.ObjectKeyFromObject(ro), err)
	}

	return ctrl.Result{
		RequeueAfter: resyncInterval,
	}, nil
}

func (r *RemoteObjectReconciler) syncObject(
	ctx context.Context,
	log logr.Logger,
	remoteObject *addonsv1alpha1.RemoteObject,
	remoteClient client.Client,
) error {
	obj, err := unstructuredFromRaw(remoteObject.Spec.Object)
	if err != nil {
		return fmt.Errorf("parsing object: %w", err)
	}

	// handle deletion
	if !remoteObject.DeletionTimestamp.IsZero() {
		if err := remoteClient.Delete(ctx, obj); err != nil &&
			!errors.IsNotFound(err) {
			return fmt.Errorf("deleting RemoteObject: %w", err)
		}
		if controllerutil.ContainsFinalizer(
			remoteObject, addonsv1alpha1.RemoteClusterFinalizer) {
			controllerutil.RemoveFinalizer(
				remoteObject, addonsv1alpha1.RemoteClusterFinalizer)
			if err := r.Update(ctx, remoteObject); err != nil {
				return fmt.Errorf("removing finalizer: %w", err)
			}
		}
		return nil
	}

	// ensure finalizer
	if !controllerutil.ContainsFinalizer(
		remoteObject, addonsv1alpha1.RemoteClusterFinalizer) {
		controllerutil.AddFinalizer(remoteObject, addonsv1alpha1.RemoteClusterFinalizer)
		if err := r.Update(ctx, remoteObject); err != nil {
			return fmt.Errorf("adding finalizer: %w", err)
		}
	}

	if err := r.reconcileObject(ctx, obj, remoteClient, log); err != nil {
		// TODO: Update RemoteObject status on non-transient errors, preventing sync.
		return fmt.Errorf("reconciling object: %w", err)
	}

	syncStatus(obj, remoteObject)
	setAvailableCondition(remoteObject)

	meta.SetStatusCondition(&remoteObject.Status.Conditions, metav1.Condition{
		Type:    addonsv1alpha1.RemoteObjectSynced,
		Status:  metav1.ConditionTrue,
		Reason:  "ObjectSynced",
		Message: "Object was synced with the RemoteCluster.",
	})
	remoteObject.Status.UpdatePhase()
	remoteObject.Status.LastHeartbeatTime = metav1.Now()
	if err := r.Status().Update(ctx, remoteObject); err != nil {
		return fmt.Errorf("updating RemoteObject Status: %w", err)
	}

	return nil
}

func (r *RemoteObjectReconciler) reconcileObject(
	ctx context.Context,
	obj *unstructured.Unstructured,
	remoteClient client.Client,
	log logr.Logger,
) error {
	currentObj := obj.DeepCopy()
	err := remoteClient.Get(ctx, client.ObjectKeyFromObject(obj), currentObj)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("getting: %w", err)
	}

	if errors.IsNotFound(err) {
		err := remoteClient.Create(ctx, obj)
		if err != nil {
			return fmt.Errorf("creating: %w", err)
		}
	}

	// Update
	if !equality.Semantic.DeepDerivative(
		obj.Object, currentObj.Object) {
		log.Info("patching spec", "obj", client.ObjectKeyFromObject(obj))
		// this is only updating "known" fields,
		// so annotations/labels and other properties will be preserved.
		err := remoteClient.Patch(
			ctx, obj, client.MergeFrom(&unstructured.Unstructured{}))

		// Alternative to override the object completely:
		// err := r.Update(ctx, obj)
		if err != nil {
			return fmt.Errorf("patching spec: %w", err)
		}
	} else {
		*obj = *currentObj
	}
	return nil
}

func unstructuredFromRaw(raw *runtime.RawExtension) (*unstructured.Unstructured, error) {
	obj := &unstructured.Unstructured{}
	if err := yaml.Unmarshal(raw.Raw, obj); err != nil {
		return nil, fmt.Errorf("converting RawExtension into unstructured: %w", err)
	}
	return obj, nil
}

func conditionsFromUnstructured(obj *unstructured.Unstructured) []metav1.Condition {
	var foundConditions []metav1.Condition

	conditions, exist, err := unstructured.
		NestedSlice(obj.Object, "status", "conditions")
	if err != nil || !exist {
		// no status conditions
		return nil
	}

	for _, condI := range conditions {
		cond, ok := condI.(map[string]interface{})
		if !ok {
			// no idea what that is supposed to be
			continue
		}

		// Check conditions observed generation, if set
		observedGeneration, _, _ := unstructured.NestedInt64(
			cond, "observedGeneration",
		)

		condType, _ := cond["type"].(string)
		condStatus, _ := cond["status"].(string)
		condReason, _ := cond["reason"].(string)
		condMessage, _ := cond["message"].(string)

		foundConditions = append(foundConditions, metav1.Condition{
			Type:               condType,
			Status:             metav1.ConditionStatus(condStatus),
			Reason:             condReason,
			Message:            condMessage,
			ObservedGeneration: observedGeneration,
		})
	}

	return foundConditions
}

func setAvailableCondition(remoteObject *addonsv1alpha1.RemoteObject) {
	switch remoteObject.Spec.AvailabilityProbe.Type {
	case addonsv1alpha1.RemoteObjectProbeCondition:
		if remoteObject.Spec.AvailabilityProbe.Condition == nil {
			return
		}

		cond := meta.FindStatusCondition(
			remoteObject.Status.Conditions,
			remoteObject.Spec.AvailabilityProbe.Condition.Type,
		)

		if cond == nil {
			meta.SetStatusCondition(&remoteObject.Status.Conditions, metav1.Condition{
				Type:   addonsv1alpha1.RemoteObjectAvailable,
				Status: metav1.ConditionUnknown,
				Reason: "MissingCondition",
				Message: fmt.Sprintf("Missing %s condition.",
					remoteObject.Spec.AvailabilityProbe.Condition.Type),
			})
		} else if cond.Status == metav1.ConditionTrue {
			meta.SetStatusCondition(&remoteObject.Status.Conditions, metav1.Condition{
				Type:   addonsv1alpha1.RemoteObjectAvailable,
				Status: metav1.ConditionTrue,
				Reason: "ProbeSuccess",
				Message: fmt.Sprintf("Probed condition %s is True.",
					remoteObject.Spec.AvailabilityProbe.Condition.Type),
			})
		} else if cond.Status == metav1.ConditionFalse {
			meta.SetStatusCondition(&remoteObject.Status.Conditions, metav1.Condition{
				Type:   addonsv1alpha1.RemoteObjectAvailable,
				Status: metav1.ConditionFalse,
				Reason: "ProbeFailure",
				Message: fmt.Sprintf("Probed condition %s is False.",
					remoteObject.Spec.AvailabilityProbe.Condition.Type),
			})
		}
	}
}

// Sync Status from obj to remoteObject checking observedGeneration.
func syncStatus(
	obj *unstructured.Unstructured,
	remoteObject *addonsv1alpha1.RemoteObject,
) {
	for _, cond := range conditionsFromUnstructured(obj) {
		// Update Condition ObservedGeneration to relate to the current Generation of the RemoteObject in the Fleet Cluster,
		// if the Condition of the object in the RemoteCluster is not stale and such a property exists.
		if cond.ObservedGeneration != 0 &&
			cond.ObservedGeneration == obj.GetGeneration() {
			cond.ObservedGeneration = remoteObject.Generation
		}

		meta.SetStatusCondition(
			&remoteObject.Status.Conditions, cond)
	}

	// Update general ObservedGeneration to relate to the current Generation of the RemoteObject,
	// if the Status of the object in the RemoteCluster is not stale and such a property exists.
	observedGeneration, ok, err := unstructured.NestedInt64(obj.Object, "status", "observedGeneration")
	if !ok || err != nil {
		// TODO: enhanced error reporting to find miss-typed fields?
		// observedGeneration field not present or of invalid type -> nothing to do
		return
	}
	if observedGeneration != 0 &&
		observedGeneration == obj.GetGeneration() {
		remoteObject.Status.ObservedGeneration = remoteObject.Generation
	}
}
