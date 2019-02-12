package mixer

import (
	"github.com/maistra/istio-operator/pkg/components/common"
)

type MixerResourceNames struct {
	common.ResourceNames
	DestinationRule string
}
type ResourceNames struct {
	Policy    MixerResourceNames
	Telemetry MixerResourceNames
}

var (
	Names = ResourceNames{
		Policy: MixerResourceNames{
			ResourceNames: common.ResourceNames{
				Deployment:     "istio-mixer-policy",
				Service:        "istio-policy",
				Role:           "istio-mixer",
				RoleBinding:    "istio-mixer",
				ServiceAccount: "istio-mixer-service-account",
			},
			DestinationRule: "istio-policy",
		},
		Telemetry: MixerResourceNames{
			ResourceNames: common.ResourceNames{
				Deployment:     "istio-mixer-telemetry",
				Service:        "istio-telemetry",
				Role:           "istio-mixer",
				RoleBinding:    "istio-mixer",
				ServiceAccount: "istio-mixer-service-account",
			},
			DestinationRule: "istio-telemetry",
		},
	}
)
