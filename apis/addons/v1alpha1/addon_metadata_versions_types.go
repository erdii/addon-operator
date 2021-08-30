package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AddonMetadataVersionSpec defines the desired state of Addon.
type AddonMetadataVersionSpec struct {
	DeletionStrategy AddonDeletionStrategy `json:"deletionStrategy"`
	// Template for addon object rendered into the output
	Template AddonTemplateSpec `json:"template"`
}

type AddonDeletionStrategy struct {
	Type AddonDeletionStrategyType `json:"type"`
}

type AddonDeletionStrategyType string

// Deletion flows are hard...
const (
	AddonDeletionStrategyWonky AddonDeletionStrategyType = "Wonky"
)

type AddonTemplateSpec struct {
	// Specification of the desired Addon.
	Spec AddonSpec `json:"spec,omitempty"`
}

// AddonMetadataStatus defines the observed state of Addon
type AddonMetadataVersionStatus struct {
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions is a list of status conditions ths object is in.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// AddonMetadataVersion is the Schema for the AddonMetadataVersions API
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type AddonMetadataVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AddonMetadataVersionSpec   `json:"spec,omitempty"`
	Status AddonMetadataVersionStatus `json:"status,omitempty"`
}

// AddonMetadataVersionList contains a list of AddonMetadataVersions
// +kubebuilder:object:root=true
type AddonMetadataVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AddonMetadataVersion `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AddonMetadataVersion{}, &AddonMetadataVersionList{})
}
