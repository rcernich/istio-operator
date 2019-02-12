package galley

import (
	"github.com/maistra/istio-operator/pkg/components/common"
)

type ResourceNames struct {
	common.ResourceNames
	ConfigMap string
}

var (
	Names = ResourceNames{
		ResourceNames: common.ResourceNames{
			Deployment:     "istio-galley",
			Service:        "istio-galley",
			Role:           "istio-galley",
			RoleBinding:    "istio-galley-admin-role-binding",
			ServiceAccount: "istio-galley-service-account",
		},
		ConfigMap: "istio-galley-configuration",
	}
)
