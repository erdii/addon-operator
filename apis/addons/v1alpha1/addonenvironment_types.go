package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AddonEnvironmentSpec struct {
	Transport AddonEnvironmentTransport `json:"transport"`
}

type AddonEnvironmentTransport struct {
	Type      AddonEnvironmentTransportType `json:"type"`
	HiveShard *HiveShardTransportSpec       `json:"hiveShard"`
}

type AddonEnvironmentTransportType string

const (
	HiveShard AddonEnvironmentTransportType = "HiveShard"
)

type HiveShardTransportSpec struct {
	ClusterRef corev1.ObjectReference `json:"clusterRef"`
}

// AddonMetadataStatus defines the observed state of Addon
type AddonEnvironmentStatus struct {
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions is a list of status conditions ths object is in.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// AddonEnvironment is the Schema for the AddonEnvironments API
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type AddonEnvironment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AddonEnvironmentSpec   `json:"spec,omitempty"`
	Status AddonEnvironmentStatus `json:"status,omitempty"`
}

// AddonEnvironmentList contains a list of AddonEnvironments
// +kubebuilder:object:root=true
type AddonEnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AddonEnvironment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AddonEnvironment{}, &AddonEnvironmentList{})
}
