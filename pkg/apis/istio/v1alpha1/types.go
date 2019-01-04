package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	operatorsv1alpha1api "github.com/openshift/api/operator/v1alpha1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type IstioOperatorConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   IstioOperatorConfigSpec   `json:"spec"`
	Status IstioOperatorConfigStatus `json:"status"`
}

type IstioOperatorConfigSpec struct {
	operatorsv1alpha1api.OperatorSpec `json:",inline"`

	GeneralConfig GeneralConfig `json:"generalConfig"`
	SidecarInjectorConfig SidecarInjectorConfig `json:"sidecarInjectorConfig"`
	SecurityConfig SecurityConfig `json:"securityConfig"`
	GatewaysConfig GatewaysConfig `json:"gatewaysConfig"`
	MixerConfig MixerConfig `json:"mixerConfig"`
	PilotConfig PilotConfig `json:"pilotConfig"`
	PrometheusConfig PrometheusConfig `json:"prometheusConfig"`
	GalleyConfig GalleyConfig `json:"galleyConfig"`
}

type GeneralConfig struct {
	metav1.TypeMeta `json:",inline"`

	ImageRegistry *string
	ImageTag *string
	MonitoringPort *int // default to 9093
	ProxyConfig *ProxyConfig
	ImagePullPolicy string
	ControlPlaneSecurityEnabled bool
	DisablePolicyChecks bool
	PolicyCheckFailOpen bool
	EnableTracing bool // true
	MTlsEnabled bool
	ImagePullSecrets *string // private registry key
	PodSchedulingPrefs *PodSchedulingPrefs
	OneNamespace bool
	ConfigValidation bool //true
	MeshExpansionConfig *MeshExpansionConfig
	MultiClusterEnabled *bool
	DefaultResources corev1.ResourceList
	PriorityClassName *string
	Crds bool //true?
	UseMcp bool //true
	TrustDomain *string
	SdsConfig *SdsConfig
}

type ProxyConfig struct {
	Image string `json:"image"` // default to proxyv2
	InitImage string // proxy_init
	ProxyDomain *string
	DiscoveryDomain *string
	Resources corev1.ResourceList // cpu = 10m
	Concurrency int
	AccessLogFile *string // default to /dev/stdout
	AccessLogEncoding *string // default to TEXT
	Privileged bool
	EnableCoreDump bool
	StatusPort int // 15020
	ReadinessProbeSettings *ProbeSettings // 1s initial delay, 2s period, 30 failure threshold
	IncludeIpRanges *string  // *
	ExcludeIpRanges *string
	IncludeInboundPorts *string // *
	ExcludeInboundPorts *string
	AutoInjectEnabled bool // true
	EnvoyStatsdConfig *EnvoyStatsdConfig
	TracerConfig *TracerConfig
}

type ProbeSettings struct {
	InitialDelaySeconds *int32 `json:"initialDelaySeconds,omitempty" protobuf:"varint,2,opt,name=initialDelaySeconds"`
	TimeoutSeconds *int32 `json:"timeoutSeconds,omitempty" protobuf:"varint,3,opt,name=timeoutSeconds"`
	PeriodSeconds *int32 `json:"periodSeconds,omitempty" protobuf:"varint,4,opt,name=periodSeconds"`
	SuccessThreshold *int32 `json:"successThreshold,omitempty" protobuf:"varint,5,opt,name=successThreshold"`
	FailureThreshold *int32 `json:"failureThreshold,omitempty" protobuf:"varint,6,opt,name=failureThreshold"`
}

type EnvoyStatsdConfig struct {
	Enabled bool
	Host *string
	Port *int
}

type TracerType string
const (
	ZipkinTracerType TracerType = "zipkin"
	LightStepTracerType TracerType = "lightstep"
)

type TracerConfig struct {
	Type TracerType
	LightStepConfig *LightStepConfig
	ZipkinConfig *ZipkinConfig
}

type LightStepConfig struct {
	Address *string
	AccessToken *string
	Secure bool // true
	CaCertPath *string
}

type ZipkinConfig struct {
	Address *string
}

type PodSchedulingPrefs struct {
	AMD64 int // 2
	S390X int // 2
	PPC64LE int // 2
}

type MeshExpansionConfig struct {
	Enabled bool
	UseIlb bool
}

type SdsConfig struct {
	Enabled *bool
	UdsPath *string
	EnableTokenMount bool
}

type SidecarInjectorConfig struct {
	Enabled bool //true
	ReplicaCount int //1
	Image string // sidecar_injector
	EnableNamespacesByDefault bool //false
}

type SecurityConfig struct {
	Enabled bool //true
	ReplicaCount int //1
	Image string // citadel
	SelfSigned bool //true
	TrustDomain *string //cluster.local
}

type GatewaysConfig struct {
	IngressGateways []IngressGatewayConfig
	EgressGateways []EgressGatewayConfig
	IlbGateways []IlbGatewayConfig
	K8sIngressConfig *K8sIngressConfig
}

type CommonGatewayConfig struct {
	Name string
	Type corev1.ServiceType
	Ports []corev1.ServicePort
	Resources corev1.ResourceList
}

type IngressGatewayConfig struct {
	CommonGatewayConfig
}

type EgressGatewayConfig struct {
	CommonGatewayConfig
}

type IlbGatewayConfig struct {
	CommonGatewayConfig
}

type K8sIngressConfig struct {
	Enabled *bool
	EnableHttps *bool
	GatewayName *string // default to ingress
}

type MixerConfig struct {
	PolicyConfig *CommonMixerConfig
	TelemetryConfig *CommonMixerConfig
	AdaptersConfig []MixerAdapterConfig
}

type CommonMixerConfig struct {
	Enabled bool
	ReplicaCount int
	AutoscalerConfig *AutoscalerConfig
}

type AutoscalerConfig struct {
	MinReplicas *int32
	MaxReplicas *int32
	TargetCPUUtilizationPercentage *int32
}

type MixerAdapterConfig struct {
	Type string
	Enabled bool
	KubernetesEnvConfig *KubernetesEnvMixerAdapterConfig
	StdIoConfig *StdIoMixerAdapterConfig
	PrometheusConfig *PrometheusMixerAdapterConfig
}

type KubernetesEnvMixerAdapterConfig struct {
}

type StdIoMixerAdapterConfig struct {
	OutputAsJson bool
}

type PrometheusMixerAdapterConfig struct {
	MetricExpiryDuration string
}

type PilotConfig struct {
	Image string
	ReplicaCount *int
	Resources corev1.ResourceList
	AutoscalerConfig *AutoscalerConfig
}

type PrometheusConfig struct {
	Image string
	Retention string
	ScrapeInterval string
	NodePort *int32
}

type IstioOperatorConfigStatus struct {
	operatorsv1alpha1api.OperatorStatus `json:",inline"`
}

type GalleyConfig struct {
	Image string
	ReplicaCount *int
	Resources corev1.ResourceList
	AutoscalerConfig *AutoscalerConfig
}


type Installation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              *InstallationSpec   `json:"spec,omitempty"`
	Status            *InstallationStatus `json:"status,omitempty"`
}

type InstallationSpec struct {
	DeploymentType *string `json:"deployment_type,omitempty"`    // "origin"
	Istio    *IstioSpec    `json:"istio,omitempty"`
	Jaeger   *JaegerSpec   `json:"jaeger,omitempty"`
	//Kiali    *KialiSpec    `json:"kiali,omitempty"`
	Launcher *LauncherSpec `json:"launcher,omitempty"`
}

type IstioSpec struct {
	Authentication *bool   `json:"authentication,omitempty"`
	Community      *bool   `json:"community,omitempty"`
	Prefix         *string `json:"prefix,omitempty"`             // "maistra/"
	Version        *string `json:"version,omitempty"`            // "0.1.0"
}

type JaegerSpec struct {
	Prefix              *string `json:"prefix,omitempty"`
	Version             *string `json:"version,omitempty"`
	ElasticsearchMemory *string `json:"elasticsearch_memory,omitempty"`  // 1Gi
}

//type KialiSpec struct {
//	Username *string `json:"username,omitempty"`
//	Password *string `json:"password,omitempty"`
//	Prefix   *string `json:"prefix,omitempty"`    // "kiali/"
//	Version  *string `json:"version,omitempty"`   // "v0.5.0"
//}

type LauncherSpec struct {
	OpenShift *OpenShiftSpec `json:"openshift,omitempty"`
	GitHub    *GitHubSpec    `json:"github,omitempty"`
	Catalog   *CatalogSpec   `json:"catalog,imitempty"`
}

type OpenShiftSpec struct {
	User     *string `json:"user,omitempty"`
	Password *string `json:"password,omitempty"`
}

type GitHubSpec struct {
	Username *string `json:"username,omitempty"`
	Token    *string `json:"token,omitempty"`
}

type CatalogSpec struct {
	Filter *string `json:"filter,omitempty"`
	Branch *string `json:"branch,omitempty"`
	Repo   *string `json:"repo,omitempty"`
}

type InstallationStatus struct {
	State *string `json:"state,omitempty"`
	Spec              *InstallationSpec   `json:"spec,omitempty"`
}