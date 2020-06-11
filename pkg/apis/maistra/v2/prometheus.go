package v2

type PrometheusAddonConfig struct {
	// only one of install or address may be specified
	// install prometheus and manage with control plane
	Install *PrometheusInstallConfig
	// use existing prometheus installation at address
	Address *string
}

type PrometheusConfig struct {
	Retention      string
	ScrapeInterval string
}

type PrometheusInstallConfig struct {
	// true if the entire install should be managed by Maistra, false if using prometheus CR (not supported)
	SelfManaged bool
	Config      PrometheusConfig
	Service     ComponentServiceConfig
	Runtime     *ComponentRuntimeConfig
	// .Values.prometheus.provisionPrometheusCert
	// 1.6+
	//ProvisionCert bool
	// this seems to overlap with provision cert, as this manifests something similar to the above
	// .Values.prometheus.security.enabled, version < 1.6
	//EnableSecurity bool
	UseTLS bool
}
