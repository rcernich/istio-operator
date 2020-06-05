package v2

type PolicyConfig struct {
	Type   PolicyType
	Mixer  *MixerPolicyConfig
	Remote *RemotePolicyConfig
	Istiod *IstiodPolicyConfig
}

type PolicyType string

const (
	PolicyTypeMixer  PolicyType = "Mixer"
	PolicyTypeRemote PolicyType = "Remote"
	PolicyTypeIstiod PolicyType = "Istiod"
)

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

