package v2

type AddonsConfig struct {
	Metrics       MetricsAddonsConfig
	Visualization VisualizationAddonsConfig
}

type MetricsAddonsConfig struct {
	Prometheus *PrometheusAddonConfig
}

type VisualizationAddonsConfig struct {
	Kiali   *KialiAddonConfig
	Grafana *GrafanaAddonConfig
}
