package v2

import (
	corev1 "k8s.io/api/core/v1"
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

type SidecarInjector struct {
	IncludeIPRanges []string
	ExcludeIPRanges []string
}
