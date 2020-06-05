package v2

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// TODO: figure out multicluster options
// The following structure could be considered for "master" control planes,
// while one which only allows specification of remote addresses for pilot,
// policy, and telemetry should be consider for "slaves."  Ideally, there would
// be tooling which would help generate an SMCPSlave resource that could be
// installed directly.
type SMCP struct {
	// .Values.global.multiCluster.clusterName, defaults to Kubernetes
	Cluster string
	// .Values.global.network
	// XXX: not sure what the difference is between this and cluster name
	Network    string
	// Should this be separate from Proxy.Logging?
	Logging   *LoggingConfig
	Policy    *PolicyConfig
	Proxy     *ProxyConfig
	Security  *SecurityConfig
	Telemetry *TelemetryConfig
	Tracing   *TracingConfig
}

type LoggingConfig struct {
	// .Values.global.proxy.logLevel, overridden by sidecar.istio.io/logLevel
	Level LogLevel
	// .Values.global.proxy.componentLogLevel, overridden by sidecar.istio.io/componentLogLevel
	// map of <component>:<level>
	ComponentLevel map[EnvoyComponent]LogLevel
	// .Values.global.logAsJson
	LogAsJSON bool
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

// XXX: original mappings from mesh-config and injector config

type AuthenticationPolicyType string

const (
	AuthenticationPolicyTypeNone      AuthenticationPolicyType = "NONE"
	AuthenticationPolicyTypeMutualTLS AuthenticationPolicyType = "MUTUAL_TLS"
	AuthenticationPolicyTypeInherit   AuthenticationPolicyType = "INHERIT"
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
