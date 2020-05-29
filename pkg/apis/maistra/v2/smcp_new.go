package v2

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type SMCP struct {
	// Should this be separate from Proxy.Logging?
	Logging   *LoggingConfig
	Policy    *PolicyConfig
	Telemetry *TelemetryConfig
	Proxy     *ProxyConfig
}

type LogLevel string

const (
	LogLevelTrace    LogLevel = "trace"
	LogLevelDebug    LogLevel = "debug"
	LogLevelInfo     LogLevel = "info"
	LogLevelWarning  LogLevel = "warning"
	LogLevelError    LogLevel = "error"
	LogLevelCritical LogLevel = "critical"
	LogLevelOff      LogLevel = "off"
)

type EnvoyComponent string

type LoggingConfig struct {
	// .Values.global.proxy.logLevel, overridden by sidecar.istio.io/logLevel
	Level LogLevel
	// .Values.global.proxy.componentLogLevel, overridden by sidecar.istio.io/componentLogLevel
	// map of <component>:<level>
	ComponentLevel map[EnvoyComponent]LogLevel
	// .Values.global.logAsJson
	LogAsJSON bool
}

// .Values.mixer.policy.enabled
type MixerPolicyConfig struct {
	// .Values.global.disablePolicyChecks | default "true" (false, inverted logic)
	// Set the following variable to false to disable policy checks by the Mixer.
	// Note that metrics will still be reported to the Mixer.
	EnableChecks bool
	// .Values.global.policyCheckFailOpen, maps to MeshConfig.policyCheckFailOpen
	// policyCheckFailOpen allows traffic in cases when the mixer policy service cannot be reached.
	// Default is false which means the traffic is denied when the client is unable to connect to Mixer.
	FailOpen bool
}

type RemotePolicyConfig struct {
	// .Values.global.remotePolicyAddress, maps to MeshConfig.mixerCheckServer
	Address string
	// .Values.global.createRemoteSvcEndpoints
	CreateServices bool
	// .Values.global.disablePolicyChecks | default "true" (false, inverted logic)
	// Set the following variable to false to disable policy checks by the Mixer.
	// Note that metrics will still be reported to the Mixer.
	EnableChecks bool
	// .Values.global.policyCheckFailOpen, maps to MeshConfig.policyCheckFailOpen
	// policyCheckFailOpen allows traffic in cases when the mixer policy service cannot be reached.
	// Default is false which means the traffic is denied when the client is unable to connect to Mixer.
	FailOpen bool
}

type IstiodPolicyConfig struct{}

type PolicyType string

const (
	PolicyTypeMixer  PolicyType = "Mixer"
	PolicyTypeRemote PolicyType = "Remote"
	PolicyTypeIstiod PolicyType = "Istiod"
)

type PolicyConfig struct {
	Type   PolicyType
	Mixer  *MixerPolicyConfig
	Remote *RemotePolicyConfig
	Istiod *IstiodPolicyConfig
}

type TelemetryBatchingConfig struct {
	// .Values.mixer.telemetry.reportBatchMaxEntries, maps to MeshConfig.reportBatchMaxEntries
	// Set reportBatchMaxEntries to 0 to use the default batching behavior (i.e., every 100 requests).
	// A positive value indicates the number of requests that are batched before telemetry data
	// is sent to the mixer server
	MaxEntries int32
	// .Values.mixer.telemetry.reportBatchMaxTime, maps to MeshConfig.reportBatchMaxTime
	// Set reportBatchMaxTime to 0 to use the default batching behavior (i.e., every 1 second).
	// A positive time value indicates the maximum wait time since the last request will telemetry data
	// be batched before being sent to the mixer server
	MaxTime string
}

// .Values.telemetry.v1.enabled
type MixerTelemetryConfig struct {
	// .Values.mixer.telemetry.sessionAffinityEnabled, maps to MeshConfig.sidecarToTelemetrySessionAffinity
	SessionAffinity bool
	Batching        TelemetryBatchingConfig
}

type RemoteTelemetryConfig struct {
	// .Values.global.remoteTelemetryAddress, maps to MeshConfig.mixerReportServer
	Address string
	// .Values.global.createRemoteSvcEndpoints
	CreateServices bool
	Batching       TelemetryBatchingConfig
}

type MetadataExchangeConfig struct {
	// .Values.telemetry.v2.metadataExchange.wasmEnabled
	// Indicates whether to enable WebAssembly runtime for metadata exchange filter.
	WASMEnabled bool
}

// previously enablePrometheusMerge
// annotates injected pods with prometheus.io annotations (scrape, path, port)
// overridden through prometheus.istio.io/merge-metrics
type PrometheusFilterConfig struct {
	// defaults to true
	Scrape bool
	// Indicates whether to enable WebAssembly runtime for stats filter.
	WASMEnabled bool
}

type StackDriverFilterConfig struct {
	// all default to false
	Logging         bool
	Monitoring      bool
	Topology        bool
	DisableOutbound bool
	ConfigOverride  map[string]string
}

type AccessLogTelemetryFilterConfig struct {
	// defaults to 43200s
	// To reduce the number of successful logs, default log window duration is
	// set to 12 hours.
	LogWindoDuration string
}

// .Values.telemetry.v2.enabled
type IstiodTelemetryConfig struct {
	// always enabled
	MetadataExchange *MetadataExchangeConfig
	// .Values.telemetry.v2.prometheus.enabled
	PrometheusFilter *PrometheusFilterConfig
	// .Values.telemetry.v2.stackdriver.enabled
	StackDriverFilter *StackDriverFilterConfig
	// .Values.telemetry.v2.accessLogPolicy.enabled
	AccessLogTelemetryFilter *AccessLogTelemetryFilterConfig
}

type TelemetryType string

const (
	TelemetryTypeMixer  TelemetryType = "Mixer"
	TelemetryTypeRemote TelemetryType = "Remote"
	TelemetryTypeIstiod TelemetryType = "Istiod"
)

type TelemetryConfig struct {
	Type   TelemetryType
	Mixer  *MixerTelemetryConfig
	Remote *RemoteTelemetryConfig
	Istiod *IstiodTelemetryConfig
}

type ProxyConfig struct {
	// XXX: should this be independent of global logging?
	Logging    LoggingConfig
	Networking ProxyNetworkingConfig
	Readiness  ProxyReadinessConfig
	Tracing    ProxyTracingConfig
	// maps to defaultConfig.proxyAdminPort, defaults to 15000
	AdminPort int32
	// .Values.global.proxy.concurrency, maps to defaultConfig.concurrency
	// XXX: removed in 1.7
	// XXX: this is defaulted to 2 in our values.yaml, but should probably be 0
	Concurrency int32
}

type ProxyTracingConfig struct {
	Jaeger      *JaegerTracerConfig
	Zipkin      *ZipkinTracerConfig
	Lightstep   *LightstepTracerConfig
	Datadog     *DatadogTracerConfig
	Stackdriver *StackdriverTracerConfig
}

type ProxyReadinessConfig struct {
	// .Values.sidecarInjectorWebhook.rewriteAppHTTPProbe, defaults to false
	// rewrite probes for application pods to route through sidecar
	RewriteApplicationProbes bool
	// .Values.global.proxy.statusPort, overridden by status.sidecar.istio.io/port, defaults to 15020
	// Default port for Pilot agent health checks. A value of 0 will disable health checking.
	// XXX: this has no affect on which port is actually used for status.
	StatusPort int32
	// .Values.global.proxy.readinessInitialDelaySeconds, overridden by readiness.status.sidecar.istio.io/initialDelaySeconds, defaults to 1
	InitialDelaySeconds int32
	// .Values.global.proxy.readinessPeriodSeconds, overridden by readiness.status.sidecar.istio.io/periodSeconds, defaults to 2
	PeriodSeconds int32
	// .Values.global.proxy.readinessFailureThreshold, overridden by readiness.status.sidecar.istio.io/failureThreshold, defaults to 30
	FailureThreshold int32
}

type ProxyNetworkingConfig struct {
	// maps to defaultConfig.connectionTimeout, defaults to 10s
	ConnectionTimeout string
	Initialization    ProxyNetworkInitConfig
	TrafficControl    ProxyTrafficControlConfig
	Protocol          ProxyNetworkProtocolConfig
	DNS               ProxyDNSConfig
}

type ProxyNetworkProtocolConfig struct {
	// .Values.global.proxy.protocolDetectionTimeout, maps to protocolDetectionTimeout
	DetectionTimeout string
	Debug            ProxyNetworkProtocolDebugConfig
}

type ProxyNetworkProtocolDebugConfig struct {
	EnableInboundSniffing  bool
	EnableOutboundSniffing bool
}

type ProxyDNSConfig struct {
	// .Values.global.podDNSSearchNamespaces
	// Custom DNS config for the pod to resolve names of services in other
	// clusters. Use this to add additional search domains, and other settings.
	// see
	// https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#dns-config
	// This does not apply to gateway pods as they typically need a different
	// set of DNS settings than the normal application pods (e.g., in
	// multicluster scenarios).
	// NOTE: If using templates, follow the pattern in the commented example below.
	//    podDNSSearchNamespaces:
	//    - global
	//    - "{{ valueOrDefault .DeploymentMeta.Namespace \"default\" }}.global"
	SearchSuffixes []string
}
type ProxyNetworkInterceptionMode string

const (
	ProxyNetworkInterceptionModeRedirect ProxyNetworkInterceptionMode = "REDIRECT"
	ProxyNetworkInterceptionModeTProxy   ProxyNetworkInterceptionMode = "TPROXY"
)

type ProxyNetworkInitType string

const (
	ProxyNetworkInitTypeCNI           ProxyNetworkInitType = "CNI"
	ProxyNetworkInitTypeInitContainer ProxyNetworkInitType = "InitContainer"
)

type ProxyNetworkInitConfig struct {
	Type ProxyNetworkInitType
	// istio_cni.enabled, use cni or iptables
	CNI           *ProxyCNIConfig
	InitContainer *ProxyInitContainerConfig
}

type ProxyCNIConfig struct {
	// TODO: add runtime configuration
}

type ProxyInitContainerConfig struct {
	// TODO: add runtime configuration
}

type ProxyTrafficControlConfig struct {
	InterceptionMode ProxyNetworkInterceptionMode
	// traffic.sidecar.istio.io/includeInboundPorts defaults to * (all ports)
	Inbound  ProxyInboundTrafficControlConfig
	Outbound ProxyOutboundTrafficControlConfig
}

type ProxyInboundTrafficControlConfig struct {
	// traffic.sidecar.istio.io/includeInboundPorts defaults to * (all ports)
	// * or comma separated list of integers
	IncludedPorts []string
}

type ProxyOutboundTrafficControlConfig struct {
	// .Values.global.proxy.includeIPRanges, overridden by traffic.sidecar.istio.io/includeOutboundIPRanges
	// * or comma separated list of CIDR
	IncludedIPRanges []string
	// .Values.global.proxy.excludeIPRanges, overridden by traffic.sidecar.istio.io/excludeOutboundIPRanges
	// * or comma separated list of CIDR
	ExcludedIPRanges []string
	// .Values.global.proxy.excludeOutboundPorts, overridden by traffic.sidecar.istio.io/excludeOutboundPorts
	// comma separated list of integers
	ExcludedPorts []int32
	// .Values.global.outboundTrafficPolicy.mode
	Policy ProxyOutboundTrafficPolicy
}

type ProxyOutboundTrafficPolicy string

const (
	ProxyOutboundTrafficPolicyAllowAny     ProxyOutboundTrafficPolicy = "ALLOW_ANY"
	ProxyOutboundTrafficPolicyRegistryOnly ProxyOutboundTrafficPolicy = "REGISTRY_ONLY"
)

// istiod/pilot
type TrafficManagementConfig struct {
	// should we allow customization of the image used? deprecate this, at a minimum
	ImagePullPolicy  corev1.PullPolicy // should default to always.  does it make sense to allow this to be configured?
	Plugins          []string          // should these be part of runtime config?
	MaxConnectionAge string            // keepaliveMaxServerConnectionAge
	// additional env vars?  should these be part of runtime or should they be incorporated into plugin configuration?
	Revision string // ??? what to do with this?
}

type MultiCluster struct {
	ClusterName   string // ISTIO_META_CLUSTER_ID
	CentralIstioD string
}

type LoggingFirstCut struct {
	LogLevel  string // based on global.  should this be part of runtime config?
	LogAsJSON bool   // based on global.  should this be part of runtime config?
}

type Debug struct {
	enableProtocolSniffingForOutbound bool
	enableProtocolSniffingForInbound  bool
	enableAnalysis                    bool
}
type Telemetry struct {
	TraceSampling string // should this be part of traffic management or telemetry?
}

// these should be global configuration
type ProxyConfigOld struct {
	ClusterDomain        string
	TrustDomain          string
	IncludeIPRanges      []string
	ExcludeIPRanges      []string
	StatusPort           string
	ExcludeOutboundPorts []string
	ImagePullPolicy      corev1.PullPolicy
	EnableCoreDump       bool // should we even expose this?
	Image                string
	LogLevel             string
	ComponentLogLevel    string
	ServicePort          string
	LogAsJSON            bool
	Lifecycle            corev1.Lifecycle // should this be part of runtime
	Network              string           // ISTIO_META_NETWORK
}

type GlobalConfig struct {
	JWTPolicy          string // global.jwtPolicy
	JWKSResolverRootCA string
	CertProvider       string // global.pilotCertProvider
	CAAddress          string
}

type DeploymentRuntimeConfig struct {
	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	// The deployment strategy to use to replace existing pods with new ones.
	// +optional
	// +patchStrategy=retainKeys
	Strategy appsv1.DeploymentStrategy `json:"strategy,omitempty" patchStrategy:"retainKeys" protobuf:"bytes,4,opt,name=strategy"`

	// The number of old ReplicaSets to retain to allow rollback.
	// This is a pointer to distinguish between explicit zero and not specified.
	// Defaults to 10.
	// +optional
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty" protobuf:"varint,6,opt,name=revisionHistoryLimit"`

	AutoScaling *AutoScalerConfig

	PodDisruption string
}

type AutoScalerConfig struct {
	// lower limit for the number of pods that can be set by the autoscaler, default 1.
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty" protobuf:"varint,2,opt,name=minReplicas"`
	// upper limit for the number of pods that can be set by the autoscaler; cannot be smaller than MinReplicas.
	MaxReplicas int32 `json:"maxReplicas" protobuf:"varint,3,opt,name=maxReplicas"`
	// target average CPU utilization (represented as a percentage of requested CPU) over all the pods;
	// if not specified the default autoscaling policy will be used.
	// +optional
	TargetCPUUtilizationPercentage *int32 `json:"targetCPUUtilizationPercentage,omitempty" protobuf:"varint,4,opt,name=targetCPUUtilizationPercentage"`
}

type PodRuntimeConfig struct {
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,7,rep,name=nodeSelector"`

	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty" protobuf:"bytes,18,opt,name=affinity"`

	// If specified, the pod will be dispatched by specified scheduler.
	// If not specified, the pod will be dispatched by default scheduler.
	// +optional
	SchedulerName string `json:"schedulerName,omitempty" protobuf:"bytes,19,opt,name=schedulerName"`

	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`

	// Compute Resources required by this container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

type SidecarInjector struct {
	IncludeIPRanges []string
	ExcludeIPRanges []string
}
