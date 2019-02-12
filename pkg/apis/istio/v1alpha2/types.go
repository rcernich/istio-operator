package v1alpha2

import (
	operatorsv1alpha1api "github.com/openshift/api/operator/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true

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
	General GeneralConfig `json:"general"`
	// Monitoring specific configuration details, e.g. logging, tracing, etc.
	Monitoring MonitoringConfig `json:"monitoring"`
	// Security specific configuration details, including Citadel
	Security SecurityConfig `json:"security"`
	// Galley specific configuration
	Galley GalleyConfig `json:"galley"`
	// Pilot specific configuration
	Pilot PilotConfig `json:"pilot"`
	// Proxy configuration
	Proxy ProxyConfig `json:"proxy"`
	// Side-car configuration
	// TODO: Move into proxy?
	SidecarInjector SidecarInjectorConfig `json:"sidecarInjector"`
	// Ingress and Egress gateway services configuration
	Gateways *GatewaysConfig `json:"gateways"`
	// Mixer configuration
	Mixer MixerConfig `json:"mixer"`
	// Prometheus configuration
	// NOT IMPLEMENTED
	PrometheusConfig *PrometheusConfig `json:"prometheus"`
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
	// PullPolicy for the image
	PullPolicy corev1.PullPolicy
	// PodSchedulingPrefs are preferences for pod scheduling
	PodSchedulingPrefs ArchSchedulingPrefs
	// PriorityClassName for the pods
	PriorityClassName string
	// ConfigValidation specifies whether or not the validating webhook should
	// be installed.  Defaults to true.
	ConfigValidation bool
	Debug *DebugConfig
}

// DebugConfig defines configuration specific to debugging
type DebugConfig struct {
	// GoDebug sets GODEBUG environment variable on pods
	GoDebug string
	// EnableCorDump enables collection of core files if a container crashes.
	// Requires ability to configure privileged on the collector container
	// (the enable-core-dump init container)
	EnableCoreDump bool
}

// DeploymentConfig specifies deployment settings specific to a component, most
// importantly, Image.Name, but allows for customization of a limited number of
// Deployment fields.
type DeploymentConfig struct {
	Image ImageConfig
	Autoscaler AutoscalerConfig
	// ReplicaCount is the number of replicas for the deployment.  Defaults to 1.
	ReplicaCount *int32
	// Resources represent deployment specific resource requirements,
	// Defaults to requests: {cpu: 10m}
	Resources *corev1.ResourceRequirements
}

// ImageConfig identifies the details of an image to use.  Defaults specified
// in GeneralConfig will be applied.
type ImageConfig struct {
	// Name of the image
	Name string
	// Tag of the image
	Tag string
	// Registry housing the image
	Registry string
}

// AutoscalerConfig represents a configuration for a horizontal auto-scaler
type AutoscalerConfig struct {
	MinReplicas                    *int32
	MaxReplicas                    *int32
	TargetCPUUtilizationPercentage *int32
}

// ArchSchedulingPrefs are used to define pod scheduling preferences based on
// node architecture.
type ArchSchedulingPrefs struct {
	// AMD64 represents amd64 architecture.  defaults to 2, NoPreference
	AMD64 *int32
	// S390X represents s390x architecture.  defaults to 2, NoPreference
	S390X *int32
	// PPC64LE represents ppc64le.  defaults to 2, NoPreference
	PPC64LE *int32
}

// MonitoringConfig represents configuration specific to monitoring
type MonitoringConfig struct {
	// EnableTracing enables tracing within the control plane.  Defaults to true.
	EnableTracing *bool
	// Tracer is the configuration for the tracer being used.  Defaults to
	// Zipkin.
	Tracer TracerConfig
	// Port is the port on which monitoring details are provided.
	// Defaults to 9093
	Port *int32
	// AccessLogFile is the name of the file to which logging should be directed.
	// Defaults to /dev/stdout
	AccessLogFile string
	// AccessLogEncoding is the type of encoding that should be used for the
	// logs.  May be TEXT or JSON.  Defaults to TEXT.
	AccessLogEncoding string
}

// TracerType is a custom type for specifying the type of tracer configured.
type TracerType string

const (
	// ZipkinTracerType for a Zipkin tracer
	ZipkinTracerType TracerType = "zipkin"
	// LightStepTracerType for a LightStep tracer
	LightStepTracerType TracerType = "lightstep"
)

// TracerConfig represents the configuration of a tracer
type TracerConfig struct {
	// Type of tracer
	Type TracerType
	// LightStep configuration
	LightStep *LightStepConfig
	// Zipkin configuration
	Zipkin *ZipkinConfig
}

// LightStepConfig is the configuration for a LightStep tracer
type LightStepConfig struct {
	// Address of the LightStep server
	Address string
	// AccessToken used to access the server
	AccessToken string
	// Secure configure secure access.  Defaults to true.
	Secure *bool
	// CACertPath is the path to the CA certs.
	CACertPath string
}

// ZipkinConfig is the configuration for a Zipkin tracer
type ZipkinConfig struct {
	// Address of the Zipkin server.  Defaults to zipkin:9411.
	Address string
}

// SecurityConfig represents security configuration for the control plane.
type SecurityConfig struct {
	// ControlPlaneSecurityEnabled enables TLS for control plane communications.
	// Defaults to false.
	ControlPlaneSecurityEnabled bool
	// DisablePolicyChecks disables policy checks against remote policy server.
	// Defaults to false.
	DisablePolicyChecks bool
	// PolicyCheckFailOpen allows traffic when remote policy server cannot be
	// reached.  Defaults to false.
	PolicyCheckFailOpen bool
	// MTLSEnabled enables use of TLS for service mesh communications.
	// Defaults to false.
	MTLSEnabled bool
	// The trust root of the system.  Defaults to cluster.local
	TrustDomain string
	// Citadel configuration.
	Citadel CitadelConfig `json:"citadel"`
}

// CitadelConfig represents the configuration for the Citadel component
type CitadelConfig struct {
	// DeploymentConfig for the Citadel component.  Default Image.Name is citadel.
	DeploymentConfig
	// SelfSigned if using self-signed certificates.  Defaults to true.
	SelfSigned *bool
}

// GalleyConfig represents the configuration for the Galley component
type GalleyConfig struct {
	// DeploymentConfig for the Galley component.  Default Image.Name is galley
	DeploymentConfig
}

// PilotConfig represents the configuration for the Pilot component
type PilotConfig struct {
	// DeploymentConfig for the Pilot component.  Default Image.Name is pilot,
	// auto-scaler is { min: 1, max: 5, targetAverageUtilization: 80 },
	// requests { cpu: 500m, memory: 2048Mi },
	DeploymentConfig
	// WatchedNamespaces are the namespaces to be watched/managed by this Istio
	// control plane.  Defaults to "*"
	WatchedNamespaces []string
	// RandomTraceSampling ...  Defaults to 100.0
	RandomTraceSampling *float64
	// Sidecar configures a sidecar (isto-proxy) container on Pilot.  Defaults to true.
	Sidecar *bool
	// PushThrottleCount corresponds to the environment variable
	// PILOT_PUSH_THROTTLE_COUNT. Defaults to 100
	PushThrottleCount *int32
}

// ProxyConfig represents the configuration used when configuring proxies
// (e.g. sidecars).
type ProxyConfig struct {
	// DeploymentConfig for the Proxy image.  Defaults Image.Name to proxyv2, requests: { cpu: 10m }
	DeploymentConfig `json:"image"`
	// InitContainer is the init image used with sidecars.  Also used when enabling
	// core dump.  Defaults Image.Name to proxy_init
	InitContainer DeploymentConfig
	// AutoInject configures the policy in the sidecar.  Defaults to enabled
	AutoInject string
	// ProxyDomain is the DNS domain suffix for pilot proxy agent.
	// XXX: for multi-tenant, this might need to be configured to the control
	// plane's namespace, so pods in other namespaces can access pilot
	ProxyDomain string
	// DiscoveryDomain is the DNS domain suffix for pilot proxy discovery.
	DiscoveryDomain string
	// EgressWhiteList represents the IP ranges to include/exclude for egress
	// Defaults to IncludeIPRanges = [ "*" ]
	EgressWhiteList *EgressWhiteList
	// IngressWhiteList represents the ports to include/exclude for ingress
	// Defaults to IncludePorts = [ "*" ]
	IngressWhiteList *IngressWhiteList
	// Concurrency represents the number of worker threads used by Proxy.
	// Default is 0, one thread per cpu.
	Concurrency *int32
	// Status configures readiness for the sidecars
	Status StatusConfig
	// Privileged if the sidecar should be run with elevated privileges.
	// Defaults to false
	Privileged bool
}

// EgressWhiteList represents the IP ranges to include/exclude for egress
type EgressWhiteList struct {
	// IncludeIPRanges are the IP ranges to which egress is allowed.
	IncludeIPRanges []string
	// ExcludeIPRanges are the IP ranges to which egress is rejected.
	ExcludeIPRanges []string
}

// IngressWhiteList represents the ports to include/exclude for ingress
type IngressWhiteList struct {
	// IncludePorts are the ports for which ingress is allowed.
	IncludePorts []string
	// ExcludePorts are the ports for which ingress is rejected.
	ExcludePorts []string
}

// StatusConfig represents the configuration for readiness/status of the sidecar
type StatusConfig struct {
	// Port on which status is available.  Defaults to 15020.
	Port *int32 `json:"[prt],omitempty" protobuf:"varint,1,opt,name=port"`
	// InitialDelaySeconds used to set Probe.InitialDelaySeconds.  Defaults to 1
	InitialDelaySeconds *int32 `json:"initialDelaySeconds,omitempty" protobuf:"varint,2,opt,name=initialDelaySeconds"`
	// TimeoutSeconds used to set Probe.TimeoutSeconds
	TimeoutSeconds *int32 `json:"timeoutSeconds,omitempty" protobuf:"varint,3,opt,name=timeoutSeconds"`
	// PeriodSeconds used to set Probe.PeriodSeconds.  Defaults to 2.
	PeriodSeconds *int32 `json:"periodSeconds,omitempty" protobuf:"varint,4,opt,name=periodSeconds"`
	// SuccessThreshold used to set Probe.SuccessThreshold
	SuccessThreshold *int32 `json:"successThreshold,omitempty" protobuf:"varint,5,opt,name=successThreshold"`
	// FailureThreshold used to set Probe.FailureThreshold.  Defaults to 30.
	FailureThreshold *int32 `json:"failureThreshold,omitempty" protobuf:"varint,6,opt,name=failureThreshold"`
}

// SidecarInjectorConfig represents the configuration for the sidecar injector
type SidecarInjectorConfig struct {
	// DeploymentConfig for the sidecar injector.  Default Image.Name is sidecar_injector
	DeploymentConfig
	// EnableNamespacesByDefault configures the webhook to inject sidecars for
	// all namespaces.  Default is false.
	EnableNamespacesByDefault bool
}

// GatewaysConfig represents the configuration for mesh ingress and egress.
type GatewaysConfig struct {
	// IngressGateway configuration.  Default name is istio-ingressgateway.
	IngressGateway *CommonGatewayConfig
	// EgressGateway configuration.  Default name is istio-egressgateway.
	EgressGateway *CommonGatewayConfig
}

// CommonGatewayConfig is configuration common to all gateways
type CommonGatewayConfig struct {
	// Type is the service type for the gateway.  Defaults to ClusterIP.
	Type corev1.ServiceType
	// Ports are the ports exposed by the gateway.
	Ports []corev1.ServicePort
	// Autoscaler configurations
	Autoscaler AutoscalerConfig
	// ReplicaCount is the number of replicas for the deployment.  Defaults to 1.
	ReplicaCount *int32
	// Resources are the resource requirements for the gateway pods
	Resources *corev1.ResourceRequirements
}

// MixerConfig represents the configuration for the Mixer component
type MixerConfig struct {
	// Policy is the configuration specific to the Policy deployment.  Defaults
	// Image.Name to mixer, auto-scaler to { min: 1, max: 5, targetAverageUtiliation: 80 },
	// env to [ GODEBUG: gctrace=2 ]
	Policy DeploymentConfig
	// Telemetry is the configuration specific to the Telemetry deployment.
	// Defaults Image.Name to mixer, auto-scaler to { min: 1, max: 5, targetAverageUtiliation: 80 },
	// env to [ GODEBUG: gctrace=2 ]
	Telemetry DeploymentConfig
	// Adapters to configure. kubernetes, prometheus and stdio are configured by
	// default.
	Adapters []MixerAdapterConfig
}

// MixerAdapterType defines a type of mixer adapter
type MixerAdapterType string

const (
	// MixerAdapterTypeKubernetesEnv represents the kubernetes env adapter
	MixerAdapterTypeKubernetesEnv MixerAdapterType = "kubernetes"
	// MixerAdapterTypePrometheus represents the prometheus adapter
	MixerAdapterTypePrometheus MixerAdapterType = "prometheus"
	// MixerAdapterTypeStdio represents the stdio adapter
	MixerAdapterTypeStdio MixerAdapterType = "stdio"
)

// MixerAdapterConfig represents the configuration of a Mixer adapter
type MixerAdapterConfig struct {
	// Type of adapter
	Type MixerAdapterType
	// Enabled ...
	Enabled bool
	// KubernetesEnv adapter config
	KubernetesEnv *KubernetesEnvMixerAdapterConfig
	// Stdio adapter config
	Stdio *StdioMixerAdapterConfig
	// Prometheus adapter config
	Prometheus *PrometheusMixerAdapterConfig
}

// KubernetesEnvMixerAdapterConfig is the configuration for the kubernetes env mixer adapter
type KubernetesEnvMixerAdapterConfig struct {
}

// StdioMixerAdapterConfig is the configuration for the stdio mixer adapter
type StdioMixerAdapterConfig struct {
	// OutputAsJSON ...  Defaults to true.
	OutputAsJSON bool
}

// PrometheusMixerAdapterConfig is the configuration for the prometheus mixer adapter
type PrometheusMixerAdapterConfig struct {
	// MetricExpiryDuration ... Defaults to 10m
	MetricExpiryDuration string
}

// PrometheusConfig represents the configuration for a prometheus component
// NOT IMPLEMENTED
type PrometheusConfig struct {
	DeploymentConfig        // docker.io/prom/prometheus:v2.3.1
	Retention        string // 6h
	ScrapeInterval   string // 15s
	NodePort         *int32 // 32090
}
