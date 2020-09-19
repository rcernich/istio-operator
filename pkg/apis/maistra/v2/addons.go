package v2

// AddonsConfig configures additional features for use with the mesh
type AddonsConfig struct {
	// Metrics configures metrics storage solutions for the mesh.
	// +optional
	Metrics *MetricsAddonsConfig `json:"metrics,omitempty"`
	// Tracing configures tracing solutions used with the mesh.
	// +optional
	Tracing *TracingConfig `json:"tracing,omitempty"`
	// Visualization configures visualization solutions used with the mesh
	// +optional
	Visualization *VisualizationAddonsConfig `json:"visualization,omitempty"`
	// Misc configures miscellaneous solutions used with the mesh
	// +optional
	Misc *MiscAddonsConfig `json:"misc,omitempty"`
}

// MetricsAddonsConfig configures metrics storage for the mesh.
type MetricsAddonsConfig struct {
	// Prometheus configures prometheus solution for metrics storage
	// .Values.prometheus.enabled, true if not null
	// implies other settings related to prometheus, e.g. .Values.telemetry.v2.prometheus.enabled,
	// .Values.kiali.prometheusAddr, etc.
	// +optional
	Prometheus *PrometheusAddonConfig `json:"prometheus,omitempty"`
}

// TracingConfig configures tracing solutions for the mesh.
// .Values.global.enableTracing
type TracingConfig struct {
	// Type represents the type of tracer to be installed.
	Type TracerType `json:"type,omitempty"`
	// Sampling sets the mesh-wide trace sampling percentage. Should be between
	// 0.0 - 100.0. Precision to 0.01, scaled as 0 to 10000, e.g.: 100% = 10000,
	// 1% = 100
	// .Values.pilot.traceSampling
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10000
	// +optional
	Sampling *int32 `json:"sampling,omitempty"`
	// Jaeger configures Jaeger as the tracer used with the mesh.
	// .Values.tracing.jaeger.enabled, true if not null
	// implies other settings related to tracing, e.g. .Values.global.tracer.zipkin.address,
	// .Values.kiali.dashboard.jaegerURL, etc.
	// +optional
	Jaeger *JaegerTracerConfig `json:"jaeger,omitempty"`
	//Zipkin      *ZipkinTracerConfig
	//Lightstep   *LightstepTracerConfig
	//Datadog     *DatadogTracerConfig
	//Stackdriver *StackdriverTracerConfig
}

// TracerType represents the tracer type to use
type TracerType string

const (
	// TracerTypeNone is used to represent no tracer
	TracerTypeNone TracerType = "None"
	// TracerTypeJaeger is used to represent Jaeger as the tracer
	TracerTypeJaeger TracerType = "Jaeger"
	// TracerTypeZipkin      TracerType = "Zipkin"
	// TracerTypeLightstep   TracerType = "Lightstep"
	// TracerTypeDatadog     TracerType = "Datadog"
	// TracerTypeStackdriver TracerType = "Stackdriver"
)

// VisualizationAddonsConfig configures visualization addons used with the mesh.
// More than one may be specified.
type VisualizationAddonsConfig struct {
	// Grafana configures a grafana instance to use with the mesh
	// .Values.grafana.enabled, true if not null
	// +optional
	Grafana *GrafanaAddonConfig `json:"grafana,omitempty"`
	// Kiali configures a kiali instance to use with the mesh
	// .Values.kiali.enabled, true if not null
	// +optional
	Kiali *KialiAddonConfig `json:"kiali,omitempty"`
}

// MiscAddonsConfig configures miscellaneous addons
type MiscAddonsConfig struct {
	// +optional
	ThreeScale *ThreeScaleConfig `json:"3scale,omitempty"`
}
