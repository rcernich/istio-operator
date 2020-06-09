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
