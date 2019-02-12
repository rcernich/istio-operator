package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	defaultCPUResourceRequest         resource.Quantity
	defaultPilotCPUResourceRequest    resource.Quantity
	defaultPilotMemoryResourceRequest resource.Quantity
)

func init() {
	var err error
	defaultCPUResourceRequest, err = resource.ParseQuantity("10m")
	if err != nil {
		panic(err)
	}
	defaultPilotCPUResourceRequest, err = resource.ParseQuantity("500m")
	if err != nil {
		panic(err)
	}
	defaultPilotMemoryResourceRequest, err = resource.ParseQuantity("2048Mi")
	if err != nil {
		panic(err)
	}
}

func SetDefaults_GeneralConfig(gc *GeneralConfig) {
	SetDefaults_DeploymentConfig(&gc.DeploymentDefaults)
	if len(gc.PullPolicy) == 0 {
		gc.PullPolicy = corev1.PullAlways
	}
	if gc.DeploymentDefaults.Resources == nil {
		rr := getDefaultResourceRequirements()
		gc.DeploymentDefaults.Resources = &rr
	}
}

func SetDefaults_DeploymentConfig(dc *DeploymentConfig) {
	if dc.ReplicaCount == nil {
		var intval int32 = 1
		dc.ReplicaCount = &intval
	}
}

func getDefaultResourceRequirements() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU: defaultCPUResourceRequest,
		},
	}
}

func SetDefaults_ArchSchedulingPrefs(asp *ArchSchedulingPrefs) {
	if asp.AMD64 == nil {
		intval := int32(2)
		asp.AMD64 = &intval
	}
	if asp.S390X == nil {
		intval := int32(2)
		asp.S390X = &intval
	}
	if asp.PPC64LE == nil {
		intval := int32(2)
		asp.PPC64LE = &intval
	}
}

// +k8s:defaulter-gen=covers
func SetDefaults_MonitoringConfig(mc *MonitoringConfig) {
	if mc.EnableTracing == nil {
		boolval := true
		mc.EnableTracing = &boolval
	}

	if *mc.EnableTracing && len(mc.Tracer.Type) == 0 {
		mc.Tracer.Type = ZipkinTracerType
		mc.Tracer.Zipkin = &ZipkinConfig{}
		SetDefaults_ZipkinConfig(mc.Tracer.Zipkin)
	}

	if mc.Port == nil {
		intval := int32(9093)
		mc.Port = &intval
	}

	if len(mc.AccessLogFile) == 0 {
		mc.AccessLogFile = "/dev/stdout"
	}

	if len(mc.AccessLogEncoding) == 0 {
		mc.AccessLogEncoding = "TEXT"
	}
}

func SetDefaults_LightStepConfig(lsc *LightStepConfig) {
	if lsc.Secure == nil {
		boolval := true
		lsc.Secure = &boolval
	}
}

func SetDefaults_ZipkinConfig(zc *ZipkinConfig) {
	if len(zc.Address) == 0 {
		zc.Address = "zipkin:9411"
	}
}

func SetDefaults_SecurityConfig(sc *SecurityConfig) {
	if len(sc.TrustDomain) == 0 {
		sc.TrustDomain = "cluster.local"
	}
}

// +k8s:defaulter-gen=covers
func SetDefaults_CitadelConfig(cc *CitadelConfig) {
	if len(cc.Image.Name) == 0 {
		cc.Image.Name = "citadel"
	}

	SetDefaults_DeploymentConfig(&cc.DeploymentConfig)

	if cc.SelfSigned == nil {
		boolval := true
		cc.SelfSigned = &boolval
	}
}

// +k8s:defaulter-gen=covers
func SetDefaults_GalleyConfig(gc *GalleyConfig) {
	if len(gc.Image.Name) == 0 {
		gc.Image.Name = "galley"
	}

	SetDefaults_DeploymentConfig(&gc.DeploymentConfig)
}

// +k8s:defaulter-gen=covers
func SetDefaults_PilotConfig(pc *PilotConfig) {
	if len(pc.Image.Name) == 0 {
		pc.Image.Name = "pilot"
	}

	setAutoscalerDefaults(&pc.Autoscaler, 1, 5, 80)

	SetDefaults_DeploymentConfig(&pc.DeploymentConfig)

	if pc.Resources == nil {
		pc.Resources = &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU: defaultPilotCPUResourceRequest,
				corev1.ResourceMemory: defaultPilotMemoryResourceRequest,
			},
		}
	}

	if len(pc.WatchedNamespaces) == 0 {
		pc.WatchedNamespaces = []string{"*"}
	}

    if pc.RandomTraceSampling == nil {
		floatval := float64(100.0)
		pc.RandomTraceSampling = &floatval
	}

	if pc.Sidecar == nil {
		boolval := true
		pc.Sidecar = &boolval
    }
    
    if pc.PushThrottleCount == nil {
        intval := int32(100)
        pc.PushThrottleCount = &intval
    }
}

// +k8s:defaulter-gen=covers
func SetDefaults_ProxyConfig(pc *ProxyConfig) {
	if len(pc.Image.Name) == 0 {
		pc.Image.Name = "proxyv2"
	}

	if pc.Resources == nil {
		pc.Resources = &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU: defaultCPUResourceRequest,
			},
		}
	}

	SetDefaults_DeploymentConfig(&pc.DeploymentConfig)

	if len(pc.InitContainer.Image.Name) == 0 {
		pc.InitContainer.Image.Name = "proxy_init"
	}

	SetDefaults_DeploymentConfig(&pc.InitContainer)

	if len(pc.AutoInject) == 0 {
		pc.AutoInject = "enabled"
	}

	if pc.EgressWhiteList == nil {
		pc.EgressWhiteList = &EgressWhiteList{
			IncludeIPRanges: []string{"*"},
			ExcludeIPRanges: []string{},
		}
	}

	if pc.IngressWhiteList == nil {
		pc.IngressWhiteList = &IngressWhiteList{
			IncludePorts: []string{"*"},
			ExcludePorts: []string{},
		}
	}

	if pc.Concurrency == nil {
		intval := int32(0)
		pc.Concurrency = &intval
	}

	SetDefaults_StatusConfig(&pc.Status)
}

func SetDefaults_StatusConfig(sc *StatusConfig) {
	if sc.Port == nil {
		intval := int32(15020)
		sc.Port = &intval
	}
	if sc.InitialDelaySeconds == nil {
		intval := int32(1)
		sc.InitialDelaySeconds = &intval
	}
	if sc.PeriodSeconds == nil {
		intval := int32(2)
		sc.PeriodSeconds = &intval
	}
	if sc.FailureThreshold == nil {
		intval := int32(30)
		sc.FailureThreshold = &intval
	}
}

// +k8s:defaulter-gen=covers
func SetDefaults_SidecarInjectorConfig(sic *SidecarInjectorConfig) {
	if len(sic.Image.Name) == 0 {
		sic.Image.Name = "sidecar_injector"
	}

	SetDefaults_DeploymentConfig(&sic.DeploymentConfig)
}

// +k8s:defaulter-gen=covers
func SetDefaults_GatewaysConfig(gc *GatewaysConfig) {
	if gc.EgressGateway == nil {
		gc.EgressGateway = &CommonGatewayConfig{}
		SetDefaults_EgressGateway(gc.EgressGateway)
	}

	if gc.IngressGateway == nil {
		gc.IngressGateway = &CommonGatewayConfig{}
		SetDefaults_IngressGateway(gc.IngressGateway)
	}
}

func SetDefaults_CommonGatewayConfig(cgc *CommonGatewayConfig) {
	if len(cgc.Type) == 0 {
		cgc.Type = corev1.ServiceTypeClusterIP
	}

	if cgc.ReplicaCount == nil {
		intval := int32(1)
		cgc.ReplicaCount = &intval
	}

	setAutoscalerDefaults(&cgc.Autoscaler, 1, 5, 80)
}

func SetDefaults_EgressGateway(egc *CommonGatewayConfig) {
    SetDefaults_CommonGatewayConfig(egc)
	if len(egc.Ports) == 0 {
		egc.Ports = []corev1.ServicePort{
			corev1.ServicePort{
				Name: "http2",
				Port: 80,
			},
			corev1.ServicePort{
				Name: "https",
				Port: 443,
			},
			corev1.ServicePort{
				Name: "tls",
				Port: 15443,
			},
		}
	}
}

func SetDefaults_IngressGateway(igc *CommonGatewayConfig) {
    SetDefaults_CommonGatewayConfig(igc)
	if len(igc.Ports) == 0 {
		igc.Ports = []corev1.ServicePort{
			corev1.ServicePort{
				Name: "http2",
				Port: 80,
			},
			corev1.ServicePort{
				Name: "https",
				Port: 443,
			},
			corev1.ServicePort{
				Name: "tls",
				Port: 15443,
			},
			// UI/Metrics, maybe ignore
			corev1.ServicePort{
				Name: "http-kiali",
				Port: 15029,
			},
			corev1.ServicePort{
				Name: "http2-prometheus",
				Port: 15030,
			},
			corev1.ServicePort{
				Name: "http2-grafana",
				Port: 15031,
			},
			corev1.ServicePort{
				Name: "http2-tracing",
				Port: 15032,
			},
		}
	}
}

// +k8s:defaulter-gen=covers
func SetDefaults_MixerConfig(mc *MixerConfig) {
	if len(mc.Policy.Image.Name) == 0 {
		mc.Policy.Image.Name = "mixer"
	}
    SetDefaults_DeploymentConfig(&mc.Policy)
    setAutoscalerDefaults(&mc.Policy.Autoscaler, 1, 5, 80)

	if len(mc.Telemetry.Image.Name) == 0 {
		mc.Telemetry.Image.Name = "mixer"
	}
	SetDefaults_DeploymentConfig(&mc.Telemetry)
    setAutoscalerDefaults(&mc.Telemetry.Autoscaler, 1, 5, 80)

    if len(mc.Adapters) == 0 {
        mc.Adapters = []MixerAdapterConfig{
            MixerAdapterConfig{
                Type: MixerAdapterTypeKubernetesEnv,
                Enabled: true,
                KubernetesEnv: &KubernetesEnvMixerAdapterConfig{},
            },
            MixerAdapterConfig{
                Type: MixerAdapterTypePrometheus,
                Enabled: true,
                Prometheus: &PrometheusMixerAdapterConfig{
                    MetricExpiryDuration: "10m",
                },
            },
            MixerAdapterConfig{
                Type: MixerAdapterTypeStdio,
                Enabled: true,
                Stdio: &StdioMixerAdapterConfig{
                    OutputAsJSON: true,
                },
            },
        }
    }
}

func setAutoscalerDefaults(ac *AutoscalerConfig, min int32, max int32, cpuUtilization int32) {
	if ac.MinReplicas == nil {
		ac.MinReplicas = &min
	}
	if ac.MaxReplicas == nil {
		ac.MaxReplicas = &max
	}
	if ac.TargetCPUUtilizationPercentage == nil {
		ac.MinReplicas = &cpuUtilization
	}
}
