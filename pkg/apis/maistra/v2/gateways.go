package v2

import (
	corev1 "k8s.io/api/core/v1"
)

// ISTIO_META_ROUTER_MODE: "sni-dnat" seems to be used by all gateway types, but does not appear to be read by
type GatewaysConfig struct {
	// works in conjunction with cluster.meshExpansion.ingress configuration
	// (for enabling ILB gateway and mesh expansion ports)
	Ingress            *GatewayConfig
	Egress             *GatewayConfig
	AdditionalGateways map[string]GatewayConfig
}

// XXX: should standard istio secrets be configured automatically, i.e. should
// the user be forced to add these manually?
type GatewayConfig struct {
	Service GatewayServiceConfig
    // sets ISTIO_META_ROUTER_MODE env, defaults to sni-dnat
    RouterMode RouterModeType
    // sets ISTIO_META_REQUESTED_NETWORK_VIEW env, defaults to empty list
    RequestedNetworkView []string
	// .Values.gateways.<gateway-name>.sds.enabled
	EnableSDS bool
	Volumes   []VolumeConfig
	Runtime   *ComponentRuntimeConfig
}

type RouterModeType string
const (
    RouterModeTypeSNI_DNAT RouterModeType = "sni-dnat"
    RouterModeTypeStandard RouterModeType = "standard"
)

type GatewayServiceConfig struct {
    // XXX: selector is ignored
	corev1.ServiceSpec
	Annotations map[string]string
	Labels      map[string]string
}

// XXX: this may be overkill, as only ConfigMap and Secret volume types are
// supported, and then mounts are only created for secret volumes.
type VolumeConfig struct {
	// Volume.Name maps to .Values.gateways.<gateway-name>.<type>.<type-name> (type-name is configMapName or secretName)
	// .configVolumes -> .configMapName = volume.name
	// .secretVolumes -> .secretName = volume.name
	// Only ConfigMap and Secret fields are supported
	Volume corev1.Volume
	// Mount.Name maps to .Values.gateways.<gateway-name>.<type>.name
	// .configVolumes -> .name = mount.name, .mountPath = mount.mountPath
	// .secretVolumes -> .name = mount.name, .mountPath = mount.mountPath
	// Only Name and MountPath fields are supported
	Mount corev1.VolumeMount
}
