package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// RemoteObject is the Schema for the RemoteObjects API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Remote Name",type="string",JSONPath=".spec.object.metadata.name"
// +kubebuilder:printcolumn:name="Remote Namespace",type="string",JSONPath=".spec.object.metadata.namespace"
// +kubebuilder:printcolumn:name="Kind",type="string",JSONPath=".spec.object.kind"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type RemoteObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RemoteObjectSpec `json:"spec,omitempty"`
	// +kubebuilder:default={phase:Pending}
	Status RemoteObjectStatus `json:"status,omitempty"`
}

// RemoteObjectList contains a list of RemoteObject
// +kubebuilder:object:root=true
type RemoteObjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RemoteObject `json:"items"`
}

// RemoteObjectSpec defines the desired state of RemoteObject.
type RemoteObjectSpec struct {
	// +kubebuilder:validation:EmbeddedResource
	// +kubebuilder:pruning:PreserveUnknownFields
	Object *runtime.RawExtension `json:"object"`
	// Configures how availability of the RemoteObject is determined.
	AvailabilityProbe RemoteObjectProbe `json:"availabilityProbe"`
	// Priority Class name.
	// RemoteObjectPriorityClasses need to be defined beforehand.
	PriorityClassName string `json:"priorityClassName,omitempty"`
}

type RemoteObjectProbe struct {
	// Type of the probe.
	// +kubebuilder:validation:Enum=Condition
	Type RemoteObjectProbeType `json:"type"`
	// Condition specific configuration parameters.
	// Only present if Type = Condition.
	Condition *RemoteObjectProbeConditionSpec `json:"condition"`
}

type RemoteObjectProbeConditionSpec struct {
	// Condition Type to check.
	Type string `json:"type"`
}

type RemoteObjectProbeType string

const (
	RemoteObjectProbeCondition RemoteObjectProbeType = "Condition"
)

type RemoteObjectStatus struct {
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions is a list of status conditions ths object is in.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// DEPRECATED: This field is not part of any API contract
	// it will go away as soon as kubectl can print conditions!
	// Human readable status - please use .Conditions from code
	Phase RemoteObjectPhase `json:"phase,omitempty"`
	// Last time the controller operated on this object.
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime,omitempty"`
}

// UpdatePhase sets the Phase according to reported conditions.
func (s *RemoteObjectStatus) UpdatePhase() {
	if availableCond := meta.FindStatusCondition(s.Conditions, RemoteObjectAvailable); availableCond != nil {
		switch availableCond.Status {
		case metav1.ConditionTrue:
			s.Phase = RemoteObjectPhaseAvailable
			return
		case metav1.ConditionFalse:
			s.Phase = RemoteObjectPhaseUnavailable
			return
		}
	}

	s.Phase = RemoteObjectPhasePending
}

const (
	// Object was successfully synced to the remote cluster.
	// True when the object was successfully synced to the remote cluster.
	// Might transition to False when subsequent updates fail to update the remote object or pull new status.
	RemoteObjectSynced = "addons.managed.openshift.io/Synced"

	// Object is considered available, based on the configured Availabilty Probe.
	RemoteObjectAvailable = "addons.managed.openshift.io/Available"
)

type RemoteObjectPhase string

// Well-known RemoteObject Phases for printing a Status in kubectl,
// see deprecation notice in RemoteObjectStatus for details.
const (
	// A RemoteObject is considered Pending until it was pushed to a RemoteCluster and it's availabilty was determined.
	RemoteObjectPhasePending RemoteObjectPhase = "Pending"
	// RemoteObjects are Available when passing their availabilityProbe.
	RemoteObjectPhaseAvailable RemoteObjectPhase = "Available"
	// RemoteObjects are Unavailable when failing their availabilityProbe.
	RemoteObjectPhaseUnavailable RemoteObjectPhase = "Unavailable"
)

func init() {
	SchemeBuilder.Register(&RemoteObject{}, &RemoteObjectList{})
}
