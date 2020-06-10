package v2

type TelemetryConfig struct {
	Type   TelemetryType
	Mixer  *MixerTelemetryConfig
	Remote *RemoteTelemetryConfig
	Istiod *IstiodTelemetryConfig
}

type TelemetryType string

const (
	TelemetryTypeMixer  TelemetryType = "Mixer"
	TelemetryTypeRemote TelemetryType = "Remote"
	TelemetryTypeIstiod TelemetryType = "Istiod"
)

// .Values.telemetry.v1.enabled
type MixerTelemetryConfig struct {
	// .Values.mixer.telemetry.sessionAffinityEnabled, maps to MeshConfig.sidecarToTelemetrySessionAffinity
	SessionAffinity bool
	Batching        TelemetryBatchingConfig
	Runtime         *DeploymentRuntimeConfig
	Adapters        *MixerTelemetryAdaptersConfig
}

type TelemetryBatchingConfig struct {
	// .Values.mixer.telemetry.reportBatchMaxEntries, maps to MeshConfig.reportBatchMaxEntries
	// Set reportBatchMaxEntries to 0 to use the default batching behavior (i.e., every 100 requests).
	// A positive value indicates the number of requests that are batched before telemetry data
	// is sent to the mixer server
	MaxEntries int32
	// .Values.mixer.telemetry.reportBatchMaxTime, maps to MeshConfig.reportBatchMaxTime
	// Set reportBatchMaxTime to 0 to use the default batching behavior (i.e., every 1 second).
	// A positive time value indicates the maximum wait time since the last request will telemetry data
	// be batched before being sent to the mixer server
	MaxTime string
}

type MixerTelemetryAdaptersConfig struct {
	// .Values.mixer.adapters.useAdapterCRDs, removed in istio 1.4, defaults to false
	UseAdapterCRDs bool
	// .Values.mixer.adapters.kubernetesenv.enabled, defaults to true
	KubernetesEnv bool
	// .Values.mixer.adapters.stdio.enabled, defaults to false (null)
	Stdio *MixerTelemetryStdioConfig
	// .Values.mixer.adapters.prometheus.enabled, defaults to true (non-null)
	Prometheus *MixerTelemetryPrometheusConfig
	// .Values.mixer.adapters.stackdriver.enabled, defaults to false (null)
	Stackdriver *MixerTelemetryStackdriverConfig
}

type MixerTelemetryStdioConfig struct {
	// .Values.mixer.adapters.stdio.outputAsJson, defaults to false
	OutputAsJSON bool
}

type MixerTelemetryPrometheusConfig struct {
	// .Values.mixer.adapters.prometheus.metricsExpiryDuration, defaults to 10m
	MetricsExpirationDuration string
}

type MixerTelemetryStackdriverConfig struct {
	Auth *MixerTelemetryStackdriverAuthConfig
	// .Values.mixer.adapters.stackdriver.tracer.enabled, defaults to false (null)
	Tracer *MixerTelemetryStackdriverTracerConfig
	// .Values.mixer.adapters.stackdriver.contextGraph.enabled, defaults to false
	EnableContextGraph bool
	// .Values.mixer.adapters.stackdriver.logging.enabled, defaults to true
	EnableLogging bool
	// .Values.mixer.adapters.stackdriver.metrics.enabled, defaults to true
	EnableMetrics bool
}

type MixerTelemetryStackdriverAuthConfig struct {
	// .Values.mixer.adapters.stackdriver.auth.appCredentials, defaults to false
	AppCredentials bool
	// .Values.mixer.adapters.stackdriver.auth.apiKey
	APIKey string
	// .Values.mixer.adapters.stackdriver.auth.serviceAccountPath
	ServiceAccountPath string
}

type MixerTelemetryStackdriverTracerConfig struct {
	// .Values.mixer.adapters.stackdriver.tracer.sampleProbability
	SampleProbability int
}

type RemoteTelemetryConfig struct {
	// .Values.global.remoteTelemetryAddress, maps to MeshConfig.mixerReportServer
	Address string
	// .Values.global.createRemoteSvcEndpoints
	CreateServices bool
	Batching       TelemetryBatchingConfig
}

// .Values.telemetry.v2.enabled
type IstiodTelemetryConfig struct {
	// always enabled
	MetadataExchange *MetadataExchangeConfig
	// .Values.telemetry.v2.prometheus.enabled
	PrometheusFilter *PrometheusFilterConfig
	// .Values.telemetry.v2.stackdriver.enabled
	StackDriverFilter *StackDriverFilterConfig
	// .Values.telemetry.v2.accessLogPolicy.enabled
	AccessLogTelemetryFilter *AccessLogTelemetryFilterConfig
}

type MetadataExchangeConfig struct {
	// .Values.telemetry.v2.metadataExchange.wasmEnabled
	// Indicates whether to enable WebAssembly runtime for metadata exchange filter.
	WASMEnabled bool
}

// previously enablePrometheusMerge
// annotates injected pods with prometheus.io annotations (scrape, path, port)
// overridden through prometheus.istio.io/merge-metrics
type PrometheusFilterConfig struct {
	// defaults to true
	Scrape bool
	// Indicates whether to enable WebAssembly runtime for stats filter.
	WASMEnabled bool
}

type StackDriverFilterConfig struct {
	// all default to false
	Logging         bool
	Monitoring      bool
	Topology        bool
	DisableOutbound bool
	ConfigOverride  map[string]string
}

type AccessLogTelemetryFilterConfig struct {
	// defaults to 43200s
	// To reduce the number of successful logs, default log window duration is
	// set to 12 hours.
	LogWindoDuration string
}
