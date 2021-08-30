package addon_metadata

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	addonsv1alpha1 "github.com/openshift/addon-operator/apis/addons/v1alpha1"
)

type AddonMetadataReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *AddonMetadataReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&addonsv1alpha1.AddonMetadata{}).
		Complete(r)
}

// AddonMetadataReconciler/Controller entrypoint
func (r *AddonMetadataReconciler) Reconcile(
	ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("addonmetadata", req.NamespacedName.String())

	addonmeta := &addonsv1alpha1.AddonMetadata{}
	err := r.Get(ctx, req.NamespacedName, addonmeta)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("reconcile", "addonmeta", addonmeta)

	// sss := &hivev1.SelectorSyncSet{
	// 	ObjectMeta: v1.ObjectMeta{
	// 		GenerateName: fmt.Sprintf("%s-", addonName),
	// 		Namespace:    "TODO",
	// 		Labels: map[string]string{
	// 			"todo/addon-name":    addonName,
	// 			"todo/addon-version": addonVersion,
	// 		},
	// 	},
	// 	Spec: hivev1.SelectorSyncSetSpec{
	// 		SyncSetCommonSpec: hivev1.SyncSetCommonSpec{
	// 			Resources: []runtime.RawExtension{
	// 				{Object: &addonsv1alpha1.Addon{
	// 					ObjectMeta: v1.ObjectMeta{
	// 						Name: addonName,
	// 					},
	// 					Spec: amv.Spec.Template.Spec,
	// 				}},
	// 			},
	// 		},
	// 	},
	// }

	// bytes, err := yaml.Marshal(sss)
	// if err != nil {
	// 	return ctrl.Result{}, fmt.Errorf("could not marshall sss. %w", err)
	// }

	// log.Info("output", "sss", sss)
	// fmt.Println(string(bytes))

	return ctrl.Result{}, nil
}
