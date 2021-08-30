package addon_metadata

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	addonsv1alpha1 "github.com/openshift/addon-operator/apis/addons/v1alpha1"
)

type HiveClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *HiveClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&addonsv1alpha1.HiveCluster{}).
		Complete(r)
}

// HiveClusterReconciler/Controller entrypoint
func (r *HiveClusterReconciler) Reconcile(
	ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("hivecluster", req.NamespacedName.String())

	hiveCluster := &addonsv1alpha1.HiveCluster{}
	err := r.Get(ctx, req.NamespacedName, hiveCluster)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("reconcile", "hiveCluster", hiveCluster)

	return ctrl.Result{}, nil
}
