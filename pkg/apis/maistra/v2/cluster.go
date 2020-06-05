package v2

type ClusterConfig struct {
	Clustering ControlPlaneClusteringConfig
}

type ControlPlaneClusteringConfig struct {
	Type   ControlPlaneType
	Master *ControlPlaneMasterConfig
	Slave  *ControlPlaneSlaveConfig
}

type ControlPlaneType string

const (
	ControlPlaneTypeMaster ControlPlaneType = "Master"
	ControlPlaneTypeSlave  ControlPlaneType = "Slave"
)

type ControlPlaneMasterConfig struct {
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

type ControlPlaneSlaveConfig struct {
	// .Values.global.remotePilotAddress
	// if specified, cannot specify MeshNetworks
    Pilot string
    Policy *RemotePolicyConfig
    // implies v1
    Telemetry *RemoteTelemetryConfig
}
