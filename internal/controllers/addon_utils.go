package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	addonsv1alpha1 "github.com/openshift/addon-operator/apis/addons/v1alpha1"
)

// Report Addon status to communicate that everything is alright
func (r *AddonReconciler) reportReadinessStatus(
	ctx context.Context, addon *addonsv1alpha1.Addon) error {
	meta.SetStatusCondition(&addon.Status.Conditions, metav1.Condition{
		Type:               addonsv1alpha1.Available,
		Status:             metav1.ConditionTrue,
		Reason:             "FullyReconciled",
		ObservedGeneration: addon.Generation,
	})
	addon.Status.ObservedGeneration = addon.Generation
	addon.Status.Phase = addonsv1alpha1.PhaseReady
	return r.Status().Update(ctx, addon)
}

// Report Addon status to communicate that the Addon is terminating
func (r *AddonReconciler) reportTerminationStatus(
	ctx context.Context, addon *addonsv1alpha1.Addon) error {
	meta.SetStatusCondition(&addon.Status.Conditions, metav1.Condition{
		Type:               addonsv1alpha1.Available,
		Status:             metav1.ConditionFalse,
		Reason:             "Terminating",
		ObservedGeneration: addon.Generation,
	})
	addon.Status.ObservedGeneration = addon.Generation
	addon.Status.Phase = addonsv1alpha1.PhaseTerminating
	return r.Status().Update(ctx, addon)
}

// Report Addon status to communicate that the resource is misconfigured
func (r *AddonReconciler) reportConfigurationError(ctx context.Context, addon *addonsv1alpha1.Addon, message string) error {
	addon.Status.ObservedGeneration = addon.Generation
	addon.Status.Phase = addonsv1alpha1.PhaseError
	meta.SetStatusCondition(&addon.Status.Conditions, metav1.Condition{
		Type:    addonsv1alpha1.Available,
		Status:  metav1.ConditionFalse,
		Reason:  "ConfigurationError",
		Message: message,
	})
	addon.Status.ObservedGeneration = addon.Generation
	addon.Status.Phase = addonsv1alpha1.PhaseError
	return r.Status().Update(ctx, addon)
}

type addonInstallConfigContext struct {
	targetNamespace, catalogSourceImage, packageName string
	stop                                             bool
}

// Validate addon.Spec.Install then extract
// targetNamespace and catalogSourceImage from it
func (r *AddonReconciler) parseAddonInstallConfig(
	ctx context.Context, log logr.Logger, addon *addonsv1alpha1.Addon) (
	// targetNamespace, catalogSourceImage string, stop bool, err error,
	configCtx addonInstallConfigContext, err error,
) {
	switch addon.Spec.Install.Type {
	case addonsv1alpha1.OwnNamespace:
		if addon.Spec.Install.OwnNamespace == nil ||
			len(addon.Spec.Install.OwnNamespace.Namespace) == 0 {
			// invalid/missing configuration
			// TODO: Move error reporting into webhook and reduce this code to a sanity check.
			configCtx.stop = true
			return configCtx, r.reportConfigurationError(ctx, addon,
				".spec.install.ownNamespace.namespace is required when .spec.install.type = OwnNamespace")
		}
		configCtx.targetNamespace = addon.Spec.Install.OwnNamespace.Namespace
		if len(addon.Spec.Install.OwnNamespace.CatalogSourceImage) == 0 {
			// invalid/missing configuration
			// TODO: Move error reporting into webhook and reduce this code to a sanity check.
			configCtx.stop = true
			return configCtx, r.reportConfigurationError(ctx, addon,
				".spec.install.ownNamespace.catalogSourceImage is required when .spec.install.type = OwnNamespace")
		}
		configCtx.catalogSourceImage = addon.Spec.Install.OwnNamespace.CatalogSourceImage
		if len(addon.Spec.Install.OwnNamespace.PackageName) == 0 {
			// invalid/missing configuration
			// TODO: Move error reporting into webhook and reduce this code to a sanity check.
			configCtx.stop = true
			return configCtx, r.reportConfigurationError(ctx, addon,
				".spec.install.ownNamespace.packageName is required when .spec.install.type = OwnNamespace")
		}
		configCtx.packageName = addon.Spec.Install.OwnNamespace.PackageName

	case addonsv1alpha1.AllNamespaces:
		if addon.Spec.Install.AllNamespaces == nil ||
			len(addon.Spec.Install.AllNamespaces.Namespace) == 0 {
			// invalid/missing configuration
			// TODO: Move error reporting into webhook and reduce this code to a sanity check.
			configCtx.stop = true
			return configCtx, r.reportConfigurationError(ctx, addon,
				".spec.install.allNamespaces.namespace is required when .spec.install.type = AllNamespaces")
		}
		configCtx.targetNamespace = addon.Spec.Install.AllNamespaces.Namespace
		if len(addon.Spec.Install.AllNamespaces.CatalogSourceImage) == 0 {
			// invalid/missing configuration
			// TODO: Move error reporting into webhook and reduce this code to a sanity check.
			configCtx.stop = true
			return configCtx, r.reportConfigurationError(ctx, addon,
				".spec.install.allNamespaces.catalogSourceImage is required when .spec.install.type = AllNamespaces")
		}
		configCtx.catalogSourceImage = addon.Spec.Install.AllNamespaces.CatalogSourceImage
		if len(addon.Spec.Install.AllNamespaces.PackageName) == 0 {
			// invalid/missing configuration
			// TODO: Move error reporting into webhook and reduce this code to a sanity check.
			configCtx.stop = true
			return configCtx, r.reportConfigurationError(ctx, addon,
				".spec.install.allNamespaces.packageName is required when .spec.install.type = AllNamespaces")
		}
		configCtx.packageName = addon.Spec.Install.AllNamespaces.PackageName

	default:
		// Unsupported Install Type
		// This should never happen, unless the schema validation is wrong.
		// The .install.type property is set to only allow known enum values.
		log.Error(fmt.Errorf("invalid Addon install type: %q", addon.Spec.Install.Type), "stopping Addon reconcilation")
		configCtx.stop = true
		return configCtx, nil
	}

	return configCtx, nil
}
