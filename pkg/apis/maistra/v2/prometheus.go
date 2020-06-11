package v2

type PrometheusAddonConfig struct {
	// true if the entire install should be managed by Maistra, false if using prometheus CR (not supported)
	SelfManaged bool
    Install PrometheusInstallConfig
    Config PrometheusConfig
}

type PrometheusConfig struct {
    Retention string
    ScrapeInterval string
}

type PrometheusInstallConfig struct {
    Service PrometheusServiceConfig
    Runtime *ComponentRuntimeConfig
	// .Values.prometheus.provisionPrometheusCert
	// 1.6+
	ProvisionCert bool
	// this seems to overlap with provision cert, as this manifests something similar to the above
	EnableSecurity bool
}

type PrometheusServiceConfig struct {
	Annotations map[string]string
	// .Values.prometheus.service.nodePort.port, ...enabled is true if not null
	NodePort *int32
	Ingress  *PrometheusIngressConfig
}

type PrometheusIngressConfig struct {
	Hosts       []string
	ContextPath string
	TLS         map[string]string // RawExtension?
	Annotations map[string]string
	Labels      map[string]string
}

