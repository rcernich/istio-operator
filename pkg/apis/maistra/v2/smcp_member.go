package v2

// XXX: resource type for external meshes belonging to a multi-cluster mesh
// TODO: are there other fields that are specific to remote k8s clusters?
type ControlPlaneClusterMemberConfig struct {
	// .Values.global.multiCluster.clusterName, defaults to Kubernetes
	Cluster string
	// .Values.global.network
	// XXX: not sure what the difference is between this and cluster name
	Network    string
	// .Values.global.remotePilotAddress
    Pilot string
    Policy *RemotePolicyConfig
    // implies v1
    Telemetry *RemoteTelemetryConfig
}
