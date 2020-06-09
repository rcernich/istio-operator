package v2

type ControlPlaneClusterConfig struct {
	// .Values.global.multiCluster.clusterName, defaults to Kubernetes
	Name string
	// .Values.global.network
	// XXX: not sure what the difference is between this and cluster name
	Network    string
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

