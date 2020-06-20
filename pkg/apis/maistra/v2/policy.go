package v2

type PolicyConfig struct {
	// Required, the policy implementation
	Type   PolicyType `json:"type,omitempty"`
	// Mixer configuration (legacy, v1)
	// .Values.mixer.policy.enabled
	Mixer  *MixerPolicyConfig `json:"mixer,omitempty"`
	// Remote mixer configuration (legacy, v1)
	// .Values.mixer.policy.remotePolicyAddress
	Remote *RemotePolicyConfig `json:"remote,omitempty"`
	// Istiod policy implementation (v2)
	// XXX: is this the default policy config, i.e. what's used if mixer is not
	// being used?  Does this need to be explicit?
	Istiod *IstiodPolicyConfig `json:"istiod,omitempty"`
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
	EnableChecks bool `json:"enableChecks,omitempty"`
	// .Values.global.policyCheckFailOpen, maps to MeshConfig.policyCheckFailOpen
	// policyCheckFailOpen allows traffic in cases when the mixer policy service cannot be reached.
	// Default is false which means the traffic is denied when the client is unable to connect to Mixer.
	FailOpen bool `json:"failOpen,omitempty"`
	Runtime  *DeploymentRuntimeConfig `json:"runtime,omitempty"`
	Adapters *MixerPolicyAdaptersConfig `json:"adapters,omitempty"`
}

type MixerPolicyAdaptersConfig struct {
	// .Values.mixer.policy.adapters.useAdapterCRDs, removed in istio 1.4, defaults to false
	UseAdapterCRDs bool `json:"useAdapterCRDs,omitempty"`
	// .Values.mixer.policy.adapters.kubernetesenv.enabled, defaults to true
	KubernetesEnv bool `json:"kubernetesenv,omitempty"`
}

type RemotePolicyConfig struct {
	// .Values.global.remotePolicyAddress, maps to MeshConfig.mixerCheckServer
	Address string `json:"address,omitempty"`
	// .Values.global.createRemoteSvcEndpoints
	CreateServices bool `json:"createServices,omitempty"`
	// .Values.global.disablePolicyChecks | default "true" (false, inverted logic)
	// Set the following variable to false to disable policy checks by the Mixer.
	// Note that metrics will still be reported to the Mixer.
	EnableChecks bool `json:"enableChecks,omitempty"`
	// .Values.global.policyCheckFailOpen, maps to MeshConfig.policyCheckFailOpen
	// policyCheckFailOpen allows traffic in cases when the mixer policy service cannot be reached.
	// Default is false which means the traffic is denied when the client is unable to connect to Mixer.
	FailOpen bool `json:"failOpen,omitempty"`
}

type IstiodPolicyConfig struct{}
