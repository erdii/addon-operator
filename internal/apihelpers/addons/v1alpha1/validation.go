package v1alpha1

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/containers/image/v5/docker"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	addonsv1alpha1 "github.com/openshift/addon-operator/apis/addons/v1alpha1"
)

type CVSValidator struct{}

func (v *CVSValidator) Validate(ctx context.Context, conditionType string, amv *addonsv1alpha1.AddonMetadataVersion) metav1.Condition {
	return metav1.Condition{
		Type:               conditionType,
		Status:             metav1.ConditionUnknown,
		Reason:             "Unimplemented",
		ObservedGeneration: amv.Generation,
		Message:            "This is umimplemented.",
	}
}

type IconValidator struct{}

func (v *IconValidator) Validate(ctx context.Context, conditionType string, amv *addonsv1alpha1.AddonMetadataVersion) metav1.Condition {
	var status metav1.ConditionStatus
	var reason, message string

	if amv.Spec.Icon == "" {
		status = metav1.ConditionFalse
		reason = "NotSpecified"
		message = ".spec.icon must be a base64 encoded image."
	} else if _, err := base64.StdEncoding.DecodeString(amv.Spec.Icon); err != nil {
		status = metav1.ConditionFalse
		reason = "NotDecodable"
		message = ".spec.icon must be a base64 encoded image."
	} else {
		status = metav1.ConditionTrue
		reason = "Decodable"
		message = ".spec.icon is base64 encoded."
	}

	return metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: amv.Generation,
	}
}

type IndexImageValidator struct{}

func (v *IndexImageValidator) Validate(ctx context.Context, conditionType string, amv *addonsv1alpha1.AddonMetadataVersion) metav1.Condition {
	ref, err := docker.ParseReference(fmt.Sprintf("//%s", amv.Spec.IndexImage))
	if err != nil {
		return metav1.Condition{
			Type:               conditionType,
			Status:             metav1.ConditionFalse,
			Reason:             "ImageRefNotParseable",
			Message:            "Can't parse .spec.indexImage reference.",
			ObservedGeneration: amv.Generation,
		}
	}

	img, err := ref.NewImage(ctx, nil)
	if err != nil {
		return metav1.Condition{
			Type:               conditionType,
			Status:             metav1.ConditionFalse,
			Reason:             "ImageRefNotOpenable",
			Message:            "Can't open .spec.indexImage reference.",
			ObservedGeneration: amv.Generation,
		}
	}
	defer img.Close()

	_, _, err = img.Manifest(ctx)
	if err != nil {
		return metav1.Condition{
			Type:               conditionType,
			Status:             metav1.ConditionFalse,
			Reason:             "ImageManifestNotOpenable",
			Message:            "Can't open .spec.indexImage reference.",
			ObservedGeneration: amv.Generation,
		}
	}

	return metav1.Condition{
		Type:               conditionType,
		Status:             metav1.ConditionTrue,
		Reason:             "AllIsGood",
		Message:            "Index image is 100% perfectly fine",
		ObservedGeneration: amv.Generation,
	}
}
