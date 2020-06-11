package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	maistrav1 "github.com/maistra/istio-operator/pkg/apis/maistra/v1"
)

func init() {
	SchemeBuilder.Register(&ServiceMeshControlPlane{}, &ServiceMeshControlPlaneList{})
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceMeshControlPlane is the Schema for the controlplanes API
// +k8s:openapi-gen=true
type ServiceMeshControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ControlPlaneSpec   `json:"spec,omitempty"`
	Status ControlPlaneStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceMeshControlPlaneList contains a list of ServiceMeshControlPlane
type ServiceMeshControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceMeshControlPlane `json:"items"`
}

// ControlPlaneStatus defines the observed state of ServiceMeshControlPlane
type ControlPlaneStatus struct {
	maistrav1.ControlPlaneStatus `json:",inline"`
}

// ControlPlaneSpec represents the configuration for installing a control plane
type ControlPlaneSpec struct {
	// XXX: the resource name is intended to be used as the revision name, which
	// is used by istio.io/rev labels/annotations to specify which control plane
	// workloads should be connecting with.

	// Version specifies what Maistra version of the control plane to install.
	// When creating a new ServiceMeshControlPlane with an empty version, the
	// admission webhook sets the version to the current version.
	// Existing ServiceMeshControlPlanes with an empty version are treated as
	// having the version set to "v1.0"
	Version string `json:"version,omitempty"`
	Cluster *ControlPlaneClusterConfig
	// Should this be separate from Proxy.Logging?
	Logging   *LoggingConfig
	Policy    *PolicyConfig
	Proxy     *ProxyConfig
	Security  *SecurityConfig
	Telemetry *TelemetryConfig
	Tracing   *TracingConfig
	Gateways  *GatewaysConfig
	// Runtime configuration for pilot (and galley, pre 1.2)
	Runtime *ControlPlaneRuntimeConfig
	Addons  *AddonsConfig
}
