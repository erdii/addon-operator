package addon_metadata_version

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	addonsv1alpha1 "github.com/openshift/addon-operator/apis/addons/v1alpha1"
	addonsv1alpha1helpers "github.com/openshift/addon-operator/internal/apihelpers/addons/v1alpha1"
)

type AddonMetadataVersionReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *AddonMetadataVersionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&addonsv1alpha1.AddonMetadataVersion{}).
		Complete(r)
}

type syncValidator func(ctx context.Context, conditionType string, amv *addonsv1alpha1.AddonMetadataVersion) metav1.Condition

// AddonMetadataVersionReconciler/Controller entrypoint
func (r *AddonMetadataVersionReconciler) Reconcile(
	ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("addonmetadataversion", req.NamespacedName.String())

	amv := &addonsv1alpha1.AddonMetadataVersion{}
	err := r.Get(ctx, req.NamespacedName, amv)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("reconcile", "amv", amv)

	syncValidators := map[string]syncValidator{
		"IconValid":       addonsv1alpha1helpers.ValidateIcon,
		"IndexImageValid": addonsv1alpha1helpers.ValidateIndexImage,
		"NamespacesValid": addonsv1alpha1helpers.ValidateNamespaces,
	}

	newConditions := false
	for conditionType := range syncValidators {
		condition := meta.FindStatusCondition(amv.Status.Conditions, conditionType)
		if condition == nil {
			meta.SetStatusCondition(&amv.Status.Conditions, metav1.Condition{
				Type:    conditionType,
				Status:  metav1.ConditionUnknown,
				Reason:  "ValidationPending",
				Message: "Validators have yet to run.",
			})
			newConditions = true
		}
	}
	if newConditions {
		return ctrl.Result{
			Requeue: true,
		}, r.Status().Update(ctx, amv)
	}

	for conditionType, validator := range syncValidators {
		// No need to check for nil values because the previous step ensures that all conditions are already present
		condition := meta.FindStatusCondition(amv.Status.Conditions, conditionType)
		if condition.ObservedGeneration == amv.Generation {
			continue
		}
		updatedCondition := validator(ctx, conditionType, amv)
		meta.SetStatusCondition(&amv.Status.Conditions, updatedCondition)
	}

	return ctrl.Result{}, r.Status().Update(ctx, amv)
}
