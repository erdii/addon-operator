package addon_metadata_version

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

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

type amvValidator interface {
	Validate(amv *addonsv1alpha1.AddonMetadataVersion) error
}

type multiError struct {
	Errs []error
}

func (e multiError) Error() string {
	buf := &strings.Builder{}
	fmt.Fprintf(buf, "multiple errors:\n")
	for _, e := range e.Errs {
		buf.WriteString("  ")
		buf.WriteString(e.Error())
		buf.WriteString("\n")
	}

	return buf.String()
}

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

	addonName, addonVersion, err := addonsv1alpha1helpers.SplitAddonMetadataVersionName(amv.Name)
	if err != nil {
		return ctrl.Result{}, err
	}

	validators := []amvValidator{
		&addonsv1alpha1helpers.CVSValidator{},
	}

	var validationErrors []error
	for _, validator := range validators {
		if err := validator.Validate(amv); err != nil {
			validationErrors = append(validationErrors, err)
		}
	}
	if len(validationErrors) != 0 {
		return ctrl.Result{}, multiError{
			Errs: validationErrors,
		}
	}

	// TODO: validation

	sss := &hivev1.SelectorSyncSet{
		ObjectMeta: v1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", addonName),
			Namespace:    "TODO",
			Labels: map[string]string{
				"todo/addon-name":    addonName,
				"todo/addon-version": addonVersion,
			},
		},
		Spec: hivev1.SelectorSyncSetSpec{
			SyncSetCommonSpec: hivev1.SyncSetCommonSpec{
				Resources: []runtime.RawExtension{
					{Object: &addonsv1alpha1.Addon{
						ObjectMeta: v1.ObjectMeta{
							Name: addonName,
						},
						Spec: amv.Spec.Template.Spec,
					}},
				},
			},
		},
	}

	bytes, err := yaml.Marshal(sss)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("could not marshall sss. %w", err)
	}

	log.Info("output", "sss", sss)
	fmt.Println(string(bytes))

	switch amv.Spec.DeletionStrategy.Type {
	case addonsv1alpha1.AddonDeletionStrategyWonky:
		// TODO: create another sss with deletion flow stuff
	default:
		return ctrl.Result{}, fmt.Errorf("unknown DeletionStrategyType: %s", amv.Spec.DeletionStrategy.Type)
	}

	// How to hand the generated SSSs over to hive?
	// Options:
	// a) hand over to another controller (and selectively bubble status up into this resource)
	// b) generate clients for all hive shards and apply them here
	//
	// Separation of concerns vs simplicity?

	return ctrl.Result{}, nil
}
