package v1alpha1

import (
	"errors"
	"fmt"

	addonsv1alpha1 "github.com/openshift/addon-operator/apis/addons/v1alpha1"
)

var (
	ErrInstallTypeOwnNamespaceNamespaceRequired           = errors.New(".spec.install.ownNamespace.namespace is required when .spec.install.type = OwnNamespace")
	ErrInstallTypeOwnNamespaceCatalogSourceImageRequired  = errors.New(".spec.install.ownNamespacee.catalogSourceImage is required when .spec.install.type = OwnNamespace")
	ErrInstallTypeAllNamespacesNamespaceRequired          = errors.New(".spec.install.allNamespaces.namespace is required when .spec.install.type = AllNamespaces")
	ErrInstallTypeAllNamespacesCatalogSourceImageRequired = errors.New(".spec.install.allNamespaces.catalogSourceImage is required when .spec.install.type = AllNamespaces")
	ErrInstallTypeUnknown                                 = errors.New(".spec.install.type is unknown")
)

// Validate addon.Spec.Install then extract
// targetNamespace and catalogSourceImage from it
func parseAddonOLMInstallConfig(addonSpec *addonsv1alpha1.AddonSpec) (
	targetNamespace, catalogSourceImage string, err error,
) {
	switch addonSpec.Install.Type {
	case addonsv1alpha1.OLMOwnNamespace:
		if addonSpec.Install.OLMOwnNamespace == nil ||
			len(addonSpec.Install.OLMOwnNamespace.Namespace) == 0 {
			// invalid/missing configuration
			// TODO: Move error reporting into webhook and reduce this code to a sanity check.
			return "", "", fmt.Errorf("%w", ErrInstallTypeOwnNamespaceNamespaceRequired)
		}
		targetNamespace = addonSpec.Install.OLMOwnNamespace.Namespace
		if len(addonSpec.Install.OLMOwnNamespace.CatalogSourceImage) == 0 {
			// invalid/missing configuration
			// TODO: Move error reporting into webhook and reduce this code to a sanity check.
			return "", "", fmt.Errorf("%w", ErrInstallTypeOwnNamespaceCatalogSourceImageRequired)
		}
		catalogSourceImage = addonSpec.Install.OLMOwnNamespace.CatalogSourceImage

	case addonsv1alpha1.OLMAllNamespaces:
		if addonSpec.Install.OLMAllNamespaces == nil ||
			len(addonSpec.Install.OLMAllNamespaces.Namespace) == 0 {
			// invalid/missing configuration
			// TODO: Move error reporting into webhook and reduce this code to a sanity check.
			return "", "", fmt.Errorf("%w", ErrInstallTypeAllNamespacesNamespaceRequired)
		}
		targetNamespace = addonSpec.Install.OLMAllNamespaces.Namespace
		if len(addonSpec.Install.OLMAllNamespaces.CatalogSourceImage) == 0 {
			// invalid/missing configuration
			// TODO: Move error reporting into webhook and reduce this code to a sanity check.
			return "", "", fmt.Errorf("%w", ErrInstallTypeAllNamespacesCatalogSourceImageRequired)
		}
		catalogSourceImage = addonSpec.Install.OLMAllNamespaces.CatalogSourceImage

	default:
		// Unsupported Install Type
		// This should never happen, unless the schema validation is wrong.
		// The .install.type property is set to only allow known enum values.
		return "", "", fmt.Errorf("%s: %w", addonSpec.Install.Type, ErrInstallTypeUnknown)
	}

	return targetNamespace, catalogSourceImage, nil
}
