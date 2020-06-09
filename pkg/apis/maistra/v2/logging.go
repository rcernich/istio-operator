package v2

type LoggingConfig struct {
	// .Values.global.proxy.logLevel, overridden by sidecar.istio.io/logLevel
	Level LogLevel
	// .Values.global.proxy.componentLogLevel, overridden by sidecar.istio.io/componentLogLevel
	// map of <component>:<level>
	ComponentLevel map[EnvoyComponent]LogLevel
	// .Values.global.logAsJson
	LogAsJSON bool
}

type LogLevel string

const (
	LogLevelTrace    LogLevel = "trace"
	LogLevelDebug    LogLevel = "debug"
	LogLevelInfo     LogLevel = "info"
	LogLevelWarning  LogLevel = "warning"
	LogLevelError    LogLevel = "error"
	LogLevelCritical LogLevel = "critical"
	LogLevelOff      LogLevel = "off"
)

type EnvoyComponent string
