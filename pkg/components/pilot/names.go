package pilot

import (
	"github.com/maistra/istio-operator/pkg/components/common"
)

var (
	Names = common.ResourceNames{
		Deployment:     "istio-pilot",
		Service:        "istio-pilot",
		Role:           "istio-pilot",
		RoleBinding:    "istio-pilot",
		ServiceAccount: "istio-pilot-service-account",
	}
)
