package v2

type KialiAddonConfig struct {
    Install *KialiInstallConfig
}

type KialiInstallConfig struct {
    // Name of Kiali CR, Namespace must match control plane namespace
    Name string
    // Create a Kiali resource if not present.  If false, will use an existing
    // Kiali resource.  This allows full customization of the Kiali CR (almost).
    Create bool
    Config      KialiConfig
    // XXX: provided for upstream support, only ingress is used, and then only
    // for enablement and contextPath
	Service     ComponentServiceConfig
    // XXX: largely unused, only image pull policy and image pull secrets are
    // relevant for maistra
	Runtime     *ComponentRuntimeConfig
}

type KialiConfig struct {
    Dashboard KialiDashboardConfig
}

type KialiDashboardConfig struct {
    ViewOnlyMode bool
    // XXX: should the user have a choice here, or should these be configured
    // automatically if they are enabled for the control plane installation?
    // Grafana endpoint will be configured based on Grafana configuration
    EnableGrafana bool
    // Prometheus endpoint will be configured based on Prometheus configuration
    EnablePrometheus bool
    // Tracing endpoint will be configured based on Tracing configuration
    EnableTracing bool
}