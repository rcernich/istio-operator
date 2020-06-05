package v2

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

type JaegerTracerConfig struct {
	// TODO....
}

type ZipkinTracerConfig struct {
	// TODO....
}

type LightstepTracerConfig struct {
	// TODO....
}

type DatadogTracerConfig struct {
	// TODO....
}

type StackdriverTracerConfig struct {
	// TODO....
}
