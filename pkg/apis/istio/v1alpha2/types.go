package v1alpha2

import (
	operatorsv1alpha1api "github.com/openshift/api/operator/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IstioControlPlane represents the configuration and state of an Istio control plane
type IstioControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   IstioControlPlaneSpec   `json:"spec"`
	Status IstioControlPlaneStatus `json:"status"`
}

// IstioControlPlaneSpec is the spec for an Istio control plane
type IstioControlPlaneSpec struct {
	operatorsv1alpha1api.OperatorSpec `json:",inline"`

	// General represents general configuration for the control plane, e.g.
	// defaults, shared settings, etc.
	General          GeneralConfig         `json:"general"`
	// Monitoring specific configuration details, e.g. logging, tracing, etc.
	Monitoring       MonitoringConfig      `json:"monitoring"`
	// Security specific configuration details, including Citadel
	Security         SecurityConfig        `json:"security"`
	// Galley specific configuration
	Galley           GalleyConfig          `json:"galley"`
	// Pilot specific configuration
	Pilot            PilotConfig           `json:"pilot"`
	// Proxy configuration
	Proxy            ProxyConfig           `json:"proxy"`
	// Side-car configuration
	// TODO: Move into proxy?
	SidecarInjector  SidecarInjectorConfig `json:"sidecarInjector"`
	// Ingress and Egress gateway services configuration
	Gateways         GatewaysConfig        `json:"gateways"`
	// Mixer configuration
	Mixer            MixerConfig           `json:"mixer"`
	// Prometheus configuration
	// NOT IMPLEMENTED
	PrometheusConfig *PrometheusConfig     `json:"prometheus"`
}

// IstioControlPlaneStatus is the status of an Istio control plane
type IstioControlPlaneStatus struct {
	operatorsv1alpha1api.OperatorStatus `json:",inline"`
}

// GeneralConfig is the general configuration for an Istio control plane
type GeneralConfig struct {
	// DeploymentDefaults are defaults to be used with component deployments,
	// e.g. image registry and tag, resources, autoscaling, etc.
	DeploymentDefaults DeploymentConfig
	// PodSchedulingPrefs are preferences for pod scheduling
	PodSchedulingPrefs PodSchedulingPrefs
	// PriorityClassName for the pods
	PriorityClassName  string
	// WatchedNamespaces are the namespaces to be watched/managed by this Istio control plane
	WatchedNamespaces       []string // false
	// ConfigValidation specifies whether or not the validating webhook should be installed
	ConfigValidation   bool // true
}
// DeploymentConfig specifies deployment settings specific to a component, most
// importantly, Image.Name, but allows for customization of a limited number of
// Deployment fields.
type DeploymentConfig struct {
	ImageConfig
	AutoscalerConfig
	// ReplicaCount is the number of replicas for the deployment
	ReplicaCount *int32                      // 1
	// Resources represent deployment specific resource requirements
	Resources    corev1.ResourceRequirements // requests: {cpu: 10m}
	// Env are any additional environment variables that should be applied
	Env          []corev1.EnvVar
}
 // ImageConfig identifies the details of an image to use.  Defaults specified
 // in GeneralConfig will be applied.
type ImageConfig struct {
	// Name of the image
	Name       string
	// Tag of the image
	Tag        string
	// Registry housing the image
	Registry   string
	// PullPolicy for the image
	PullPolicy string
}

// AutoscalerConfig represents a configuration for a horizontal auto-scaler
type AutoscalerConfig struct {
	MinReplicas                    *int32
	MaxReplicas                    *int32
	TargetCPUUtilizationPercentage *int32
}

// PodSchedulingPrefs ...
type PodSchedulingPrefs struct {
	AMD64   *int32 // 2
	S390X   *int32 // 2
	PPC64LE *int32 // 2
}

// MonitoringConfig represents configuration specific to monitoring
type MonitoringConfig struct {
	// EnableTracing enables tracing within the control plane
	EnableTracing     bool // true
	// TracerConfig is the configuration for the tracer being used
	TracerConfig      TracerConfig
	// MonitoringPort is the port on which monitoring details are provided
	MonitoringPort    *int32 // default to 9093
	// AccessLogFile is the name of the file to which logging should be directed
	AccessLogFile     string // default to /dev/stdout
	// AccessLogEncoding is the type of encoding that should be used for the
	// logs.  May be TEXT or JSON
	AccessLogEncoding string // default to TEXT
}

// TracerType is a custom type for specifying the type of tracer configured.
type TracerType string

const (
	// ZipkinTracerType for a Zipkin tracer
	ZipkinTracerType    TracerType = "zipkin"
	// LightStepTracerType for a LightStep tracer
	LightStepTracerType TracerType = "lightstep"
)

// TracerConfig represents the configuration of a tracer
type TracerConfig struct {
	// Type of tracer
	Type      TracerType // zipkin
	// LightStep configuration
	LightStep *LightStepConfig
	// Zipkin configuration
	Zipkin    *ZipkinConfig
}

type LightStepConfig struct {
	Address     string
	AccessToken string
	Secure      bool // true
	CaCertPath  string
}

type ZipkinConfig struct {
	Address string // zipkin:9411
}

type SecurityConfig struct {
	ControlPlaneSecurityEnabled bool // false
	DisablePolicyChecks         bool // false
	PolicyCheckFailOpen         bool // false
	MTLSEnabled                 bool // false
	TrustDomain                 string
	Citadel                     CitadelConfig `json:"citadel"`
}

type CitadelConfig struct {
	DeploymentConfig        // citadel
	SelfSigned       bool   // true
	TrustDomain      string // cluster.local
}

type GalleyConfig struct {
	DeploymentConfig // galley
}

type PilotConfig struct {
	DeploymentConfig        // pilot, min=1, max=5, targetAverageUtilization=80, requests: {cpu: 500m, memory: 2048Mi}, env: [ PILOT_PUSH_THROTTLE_COUNT: 100, GODEBUG: gctrace=2 ]
	TraceSampling    string // 100.0
	Sidecar          bool   // true
}

type ProxyConfig struct {
	Image            DeploymentConfig `json:"image"` // proxyv2, requests: { cpu: 10m }
	InitImage        DeploymentConfig // proxy_init
	AutoInject       bool             // true
	ProxyDomain      string
	DiscoveryDomain  string
	EgressWhiteList  EgressWhiteList
	IngressWhiteList IngressWhiteList
	Concurrency      *int32       // 0
	Status           StatusConfig // 15020
	Privileged       bool         // false
	EnableCoreDump   bool         // false
}

type EgressWhiteList struct {
	IncludeIPRanges string // *
	ExcludeIPRanges string
}

type IngressWhiteList struct {
	IncludePorts string // *
	ExcludePorts string
}

type StatusConfig struct {
	// 15020
	Port *int32 `json:"[prt],omitempty" protobuf:"varint,1,opt,name=port"`
	// 1
	InitialDelaySeconds *int32 `json:"initialDelaySeconds,omitempty" protobuf:"varint,2,opt,name=initialDelaySeconds"`
	TimeoutSeconds      *int32 `json:"timeoutSeconds,omitempty" protobuf:"varint,3,opt,name=timeoutSeconds"`
	// 2
	PeriodSeconds    *int32 `json:"periodSeconds,omitempty" protobuf:"varint,4,opt,name=periodSeconds"`
	SuccessThreshold *int32 `json:"successThreshold,omitempty" protobuf:"varint,5,opt,name=successThreshold"`
	// 30
	FailureThreshold *int32 `json:"failureThreshold,omitempty" protobuf:"varint,6,opt,name=failureThreshold"`
}

type SidecarInjectorConfig struct {
	DeploymentConfig
	EnableNamespacesByDefault bool // false
}

type GatewaysConfig struct {
	IngressGateway *CommonGatewayConfig
	EgressGateway  *CommonGatewayConfig
}

type CommonGatewayConfig struct {
	Type      corev1.ServiceType
	Ports     []corev1.ServicePort
	Resources corev1.ResourceList
}

type IngressGatewayConfig struct {
	CommonGatewayConfig // ClusterIP, ports....
}

type EgressGatewayConfig struct {
	CommonGatewayConfig // ClusterIP, ports....
}

type MixerConfig struct {
	Policy    DeploymentConfig // mixer, min=1, max=5, targetAverageUtilization=80, env: [ GODEBUG: gctrace=2 ]
	Telemetry DeploymentConfig // mixer, min=1, max=5, targetAverageUtilization=80, env: [ GODEBUG: gctrace=2 ]
	Adapters  []MixerAdapterConfig
}

type MixerAdapterType string

const (
	KubernetesEnvMixerAdapterType MixerAdapterType = "k8s"
	PrometheusMixerAdapterType    MixerAdapterType = "prometheus"
	StdioMixerAdapterType         MixerAdapterType = "stdio"
)

type MixerAdapterConfig struct {
	Type          MixerAdapterType
	Enabled       bool
	KubernetesEnv *KubernetesEnvMixerAdapterConfig // true
	Stdio         *StdioMixerAdapterConfig // true
	Prometheus    *PrometheusMixerAdapterConfig // true
}

type KubernetesEnvMixerAdapterConfig struct {
}

type StdioMixerAdapterConfig struct {
	OutputAsJson bool // true
}

type PrometheusMixerAdapterConfig struct {
	MetricExpiryDuration string // 10m
}

type PrometheusConfig struct {
	DeploymentConfig // docker.io/prom/prometheus:v2.3.1
	Retention      string // 6h
	ScrapeInterval string // 15s
	NodePort       *int32 // 32090
}
