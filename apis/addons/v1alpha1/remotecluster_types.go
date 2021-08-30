package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RemoteCluster is the Schema for the RemoteClusters API
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Namespace",type="string",JSONPath=".status.localNamespace"
// +kubebuilder:printcolumn:name="API Server",type="string",JSONPath=".status.remoteClusterState.apiServer"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.remoteClusterState.version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type RemoteCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RemoteClusterSpec `json:"spec,omitempty"`
	// +kubebuilder:default={phase:Pending}
	Status RemoteClusterStatus `json:"status,omitempty"`
}

// RemoteClusterList contains a list of RemoteCluster
// +kubebuilder:object:root=true
type RemoteClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RemoteCluster `json:"items"`
}

// RemoteClusterSpec defines the desired state of RemoteCluster.
type RemoteClusterSpec struct {
	// Secret pointing to a kubeconfig to connect to a remote cluster.
	KubeconfigSecret corev1.SecretReference `json:"kubeconfigSecret"`
	// Remote cluster resync interval.
	// +kubebuilder:default="5m"
	ResyncInterval metav1.Duration `json:"resyncInterval"`
}

type RemoteClusterStatus struct {
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions is a list of status conditions ths object is in.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// DEPRECATED: This field is not part of any API contract
	// it will go away as soon as kubectl can print conditions!
	// Human readable status - please use .Conditions from code
	Phase RemoteClusterPhase `json:"phase,omitempty"`
	// RemoteCluster state information.
	Remote RemoteClusterState `json:"remoteClusterState,omitempty"`
	// LocalNamespace containing remote objects for this Kubernetes Cluster.
	LocalNamespace string `json:"localNamespace,omitempty"`
	// Last time the controller operated on this cluster.
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime,omitempty"`
}

// UpdatePhase sets the Phase according to reported conditions.
func (s *RemoteClusterStatus) UpdatePhase() {
	reachableCond := meta.FindStatusCondition(s.Conditions, RemoteClusterReachable)
	if reachableCond == nil {
		s.Phase = RemoteClusterPhasePending
		return
	}

	if reachableCond.Status == metav1.ConditionFalse {
		s.Phase = RemoteClusterPhaseUnreachable
		return
	}

	if reachableCond.Status == metav1.ConditionTrue {
		s.Phase = RemoteClusterPhaseReachable
		return
	}

	s.Phase = RemoteClusterPhaseUnknown
}

type RemoteClusterState struct {
	// Kubernetes API Server address used to talk with the remote cluster.
	// Will be automatically discovered from the referenced Kubeconfig secret.
	APIServer string `json:"apiServer"`
	// Version of the API server.
	Version string `json:"version,omitempty"`
}

type RemoteClusterPhase string

// Well-known RemoteCluster Phases for printing a Status in kubectl,
// see deprecation notice in RemoteClusterStatus for details.
const (
	RemoteClusterPhasePending     RemoteClusterPhase = "Pending"
	RemoteClusterPhaseReachable   RemoteClusterPhase = "Reachable"
	RemoteClusterPhaseUnreachable RemoteClusterPhase = "Unreachable"
	RemoteClusterPhaseUnknown     RemoteClusterPhase = "Unknown"
)

const (
	// Reachable condition is True as long as the remote cluster can be contacted.
	RemoteClusterReachable = "Reachable"
)

const RemoteClusterFinalizer = "addons.managed.openshift.io/cleanup"

const (
	// SecretTypeKubeconfig contains a Kubeconfig to connect to another cluster.
	//
	// Required fields:
	// - Secret.Data["kubeconfig"] - Kubeconfig yaml.
	SecretTypeKubeconfig corev1.SecretType = "addons.managed.openshift.io/kubeconfig"

	// Key of the Kubeconfig for a kubeconfig secret.
	SecretKubeconfigKey = "kubeconfig"
)

func init() {
	SchemeBuilder.Register(&RemoteCluster{}, &RemoteClusterList{})
}
