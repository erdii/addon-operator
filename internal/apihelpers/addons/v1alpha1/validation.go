package v1alpha1

import addonsv1alpha1 "github.com/openshift/addon-operator/apis/addons/v1alpha1"

type CVSValidator struct{}

func (v *CVSValidator) Validate(amv *addonsv1alpha1.AddonMetadataVersion) error {
	return nil
}
