package sidecar

import (
	"github.com/maistra/istio-operator/pkg/components/common"
)

type ResourceNames struct {
	common.ResourceNames
	ConfigMap       string
	MutatingWebhook string
}

var (
	Names = ResourceNames{
		ResourceNames: common.ResourceNames{
			Deployment:     "istio-sidecar-injector",
			Service:        "istio-sidecar-injector",
			Role:           "istio-sidecar-injector",
			RoleBinding:    "istio-sidecar-injector-admin-role-binding",
			ServiceAccount: "istio-sidecar-injector-service-account",
		},
		ConfigMap: "istio-sidecar-injector",
	}
)
