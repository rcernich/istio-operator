package v2

type AddonsConfig struct {
	Metrics       MetricsAddonsConfig
	Tracing       TracingConfig
	Visualization VisualizationAddonsConfig
}

type MetricsAddonsConfig struct {
	Prometheus *PrometheusAddonConfig
}

// .Values.global.enableTracing
type TracingConfig struct {
	Type        TracingType
	Jaeger      *JaegerTracerConfig
	Zipkin      *ZipkinTracerConfig
	Lightstep   *LightstepTracerConfig
	Datadog     *DatadogTracerConfig
	Stackdriver *StackdriverTracerConfig
}

type TracingType string

const (
	TracingTypeJaeger      TracingType = "Jaeger"
	TracingTypeZipkin      TracingType = "Zipkin"
	TracingTypeLightstep   TracingType = "Lightstep"
	TracingTypeDatadog     TracingType = "Datadog"
	TracingTypeStackdriver TracingType = "Stackdriver"
)

type VisualizationAddonsConfig struct {
	Grafana *GrafanaAddonConfig
	Kiali   *KialiAddonConfig
}
