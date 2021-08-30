package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HiveClusterSpec specifies how to connect to a hive shard.
type HiveClusterSpec struct {
	APIUrl              string                 `json:"apiURL"`
	KubeconfigSecretRef corev1.ObjectReference `json:"kubeconfigSecretRef"`
}

// HiveClusterStatus defines the observed state of a hive shard.
type HiveClusterStatus struct {
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions is a list of status conditions ths object is in.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// HiveCluster is the Schema for the HiveClusters API
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type HiveCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HiveClusterSpec   `json:"spec,omitempty"`
	Status HiveClusterStatus `json:"status,omitempty"`
}

// HiveClusterList contains a list of HiveClusters
// +kubebuilder:object:root=true
type HiveClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HiveCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HiveCluster{}, &HiveClusterList{})
}
