package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AddonMetadataSpec defines the desired state of Addon.
type AddonMetadataSpec struct {
	CurrentVersionRef      corev1.ObjectReference `json:"currentVersionRef"`
	HiveClusterMatchLabels metav1.LabelSelector   `json:"hiveClusterMatchLabels"`
}

// AddonMetadataStatus defines the observed state of Addon
type AddonMetadataStatus struct {
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions is a list of status conditions ths object is in.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// AddonMetadata is the Schema for the AddonMetadatas API
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type AddonMetadata struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AddonMetadataSpec   `json:"spec,omitempty"`
	Status AddonMetadataStatus `json:"status,omitempty"`
}

// AddonMetadataList contains a list of AddonMetadatas
// +kubebuilder:object:root=true
type AddonMetadataList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AddonMetadata `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AddonMetadata{}, &AddonMetadataList{})
}
