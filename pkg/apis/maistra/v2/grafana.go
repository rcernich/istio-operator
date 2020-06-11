package v2

type GrafanaAddonConfig struct {
	// only one of install or address may be specified
	// install grafana and manage with control plane
	Install *GrafanaInstallConfig
	// use existing grafana installation at address
	Address *string
}

type GrafanaInstallConfig struct {
	// true if the entire install should be managed by Maistra, false if using grafana CR (not supported)
	SelfManaged bool
	Config      GrafanaConfig
	Service     ComponentServiceConfig
	// .Values.grafana.persist, true if not null
	Persistence *ComponentPersistenceConfig
	Runtime     *ComponentRuntimeConfig
    // .Values.grafana.security.enabled, true if not null
    // XXX: unused for maistra, as we use oauth-proxy
    Security *GrafanaSecurityConfig
}

type GrafanaConfig struct {
	// This is pretty cheesy...
	// .Values.grafana.env
	Env map[string]string
	// .Values.grafana.envSecrets
	EnvSecrets map[string]string
}

type GrafanaSecurityConfig struct {
    SecretName string
    UsernameKey string
    PassphraseKey string
}
