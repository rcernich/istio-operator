package gateway

import (
	"github.com/maistra/istio-operator/pkg/components/common"
)

type ResourceNames struct {
	Egress  common.ResourceNames
	Ingress common.ResourceNames
}

var (
	Names = ResourceNames{
		Egress: common.ResourceNames{
			Deployment:     "istio-egressgateway",
			Service:        "istio-egressgateway",
			Role:           "istio-egressgateway",
			RoleBinding:    "istio-egressgateway",
			ServiceAccount: "istio-egressgateway-service-account",
		},
		Ingress: common.ResourceNames{
			Deployment:     "istio-ingressgateway",
			Service:        "istio-ingressgateway",
			Role:           "istio-ingressgateway",
			RoleBinding:    "istio-ingressgateway",
			ServiceAccount: "istio-ingressgateway-service-account",
		},
	}
)
