package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RemoteObjectPriorityClass is the Schema for the RemoteObjectPriorityClasss API
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Resync Interval",type="string",JSONPath=".resyncInterval"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type RemoteObjectPriorityClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	ResyncInterval metav1.Duration `json:"resyncInterval"`
}

// RemoteObjectPriorityClassList contains a list of RemoteObjectPriorityClass
// +kubebuilder:object:root=true
type RemoteObjectPriorityClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RemoteObjectPriorityClass `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RemoteObjectPriorityClass{}, &RemoteObjectPriorityClassList{})
}
