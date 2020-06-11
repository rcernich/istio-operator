package v2

type ControlPlaneClusterConfig struct {
	// .Values.global.multiCluster.clusterName, defaults to Kubernetes
	Name string
	// .Values.global.network
	// XXX: not sure what the difference is between this and cluster name
	Network string
	// .Values.global.multiCluster.enabled, if not null
	MultiCluster *MultiClusterConfig
	// .Values.global.meshExpansion.enabled, if not null
	// XXX: it's not clear whether or not there is any overlap with MultiCluster,
	// i.e. does MultiCluster require mesh expansion ports to be configured on
	// the ingress gateway?
	MeshExpansion *MeshExpansionConfig
}

// implies the following:
// adds external to RequestedNetworkView (ISTIO_META_REQUESTED_NETWORK_VIEW) for egress gateway
// adds "global" and "{{ valueOrDefault .DeploymentMeta.Namespace \"default\" }}.global" to pod dns search suffixes
type MultiClusterConfig struct {
	// .Values.global.k8sIngress.enabled
	// implies the following:
	// .Values.global.k8sIngress.gatewayName will match the ingress gateway
	// .Values.global.k8sIngress.enableHttps will be true if gateway service exposes port 443
	// XXX: not sure whether or not this is specific to multicluster, mesh expansion, or both
	Ingress bool
	// .Values.global.meshNetworks
	// XXX: if non-empty, local cluster network should be configured as:
	//  <spec.cluster.network>:
	//      endpoints:
	//      - fromRegistry: <spec.cluster.name>
	//      gateways:
	//      - service: <ingress-gateway-service-name>
	//        port: 443 # mtls port
	MeshNetworks map[string][]MeshNetworkConfig
}

type MeshExpansionConfig struct {
	// .Values.global.meshExpansion.useILB, true if not null, otherwise uses ingress gateway
	ILBGateway *ILBGatewayConfig
}

type ILBGatewayConfig struct {
	// ports for ILB gateway are hard coded
	// service type is hard-coded to LoadBalancer
	Service GatewayServiceConfig
	Volumes []VolumeConfig
	Runtime ComponentRuntimeConfig
}

type MeshNetworkConfig struct {
	Endpoints []MeshEndpointConfig
	Gateways  []MeshGatewayConfig
}

type MeshEndpointConfig struct {
	FromRegistry string
	FromCIDR     string
}

type MeshGatewayConfig struct {
	Service string
	Address string
	Port    int32
}
