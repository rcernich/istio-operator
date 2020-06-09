package v2

import (
	corev1 "k8s.io/api/core/v1"
)

type ProxyConfig struct {
	// XXX: should this be independent of global logging?
	Logging    LoggingConfig
	Networking ProxyNetworkingConfig
	Runtime    ProxyRuntimeConfig
	// maps to defaultConfig.proxyAdminPort, defaults to 15000
	AdminPort int32
	// .Values.global.proxy.concurrency, maps to defaultConfig.concurrency
	// XXX: removed in 1.7
	// XXX: this is defaulted to 2 in our values.yaml, but should probably be 0
	Concurrency int32
}

type ProxyNetworkingConfig struct {
	// maps to defaultConfig.connectionTimeout, defaults to 10s
	ConnectionTimeout string
	Initialization    ProxyNetworkInitConfig
	TrafficControl    ProxyTrafficControlConfig
	Protocol          ProxyNetworkProtocolConfig
	DNS               ProxyDNSConfig
}

type ProxyNetworkInitConfig struct {
	Type ProxyNetworkInitType
	// istio_cni.enabled, use cni or iptables
	CNI           *ProxyCNIConfig
	InitContainer *ProxyInitContainerConfig
}

type ProxyNetworkInitType string

const (
	ProxyNetworkInitTypeCNI           ProxyNetworkInitType = "CNI"
	ProxyNetworkInitTypeInitContainer ProxyNetworkInitType = "InitContainer"
)

type ProxyCNIConfig struct {
    // TODO: add runtime configuration
    Runtime *ProxyCNIRuntimeConfig
}

type ProxyCNIRuntimeConfig struct {
    ContainerConfig
	PriorityClassName string `json:"priorityClassName,omitempty" protobuf:"bytes,24,opt,name=priorityClassName"`
}

type ProxyInitContainerConfig struct {
	// TODO: add runtime configuration
    Runtime *ContainerConfig
}

type ProxyTrafficControlConfig struct {
	InterceptionMode ProxyNetworkInterceptionMode
	// traffic.sidecar.istio.io/includeInboundPorts defaults to * (all ports)
	Inbound  ProxyInboundTrafficControlConfig
	Outbound ProxyOutboundTrafficControlConfig
}

type ProxyNetworkInterceptionMode string

const (
	ProxyNetworkInterceptionModeRedirect ProxyNetworkInterceptionMode = "REDIRECT"
	ProxyNetworkInterceptionModeTProxy   ProxyNetworkInterceptionMode = "TPROXY"
)

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

type ProxyRuntimeConfig struct {
	Readiness ProxyReadinessConfig
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
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
