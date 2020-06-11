package v2

// XXX: this currently deviates from upstream, which creates a jaeger all-in-one deployment manually
type JaegerTracerConfig struct {
	// Name of Jaeger CR, Namespace must match control plane namespace
	Name string
	// Create a Jaeger resource if not present.  If false, will use an existing
	// named Jaeger resource.  This allows full customization of the Jaeger CR.
	Create bool
	Config JaegerConfig
	// Used to configure resources and affinity.  runtime.pod.containers can be
	// used to override details for specific jaeger components, e.g. allInOne,
	// query, etc.  runtime.metadata.annotations maps to
	// .Values.tracing.jaeger.annotations
	Runtime *ComponentRuntimeConfig
	// .Values.tracing.jaeger.ingress.enabled, fals if null
	Ingress *JaegerIngressConfig
}

type JaegerConfig struct {
	Storage *JaegerStorageConfig
}

type JaegerStorageConfig struct {
	Type JaegerStorageType
	// implies all-in-one template
	Memory *JaegerMemoryStorageConfig
	// implies production-elasticsearch template
	Elasticsearch *JaegerElasticsearchStorageConfig
}

type JaegerStorageType string

const (
	JaegerStorageTypeMemory        JaegerStorageType = "Memory"
	JaegerStorageTypeElasticsearch JaegerStorageType = "Elasticsearch"
)

type JaegerMemoryStorageConfig struct {
	// .Values.tracing.jaeger.memory.max_traces, defaults to 100000
	MaxTraces int64
}

type JaegerElasticsearchStorageConfig struct {
	// .Values.tracing.jaeger.elasticsearch.nodeCount, defaults to 3
	NodeCount int32
	// .Values.tracing.jaeger.elasticsearch.storage, raw yaml
	Storage map[string]string // Raw?
	// .Values.tracing.jaeger.elasticsearch.redundancyPolicy, raw yaml
	RedundancyPolicy map[string]string
	// .Values.tracing.jaeger.elasticsearch.esIndexCleaner, raw yaml
	IndexCleaner map[string]string
	// used for node selector, etc., specific to elasticsearch config
	Runtime *PodRuntimeConfig
}

type JaegerIngressConfig struct {
	Metadata MetadataConfig
}
