package citadel

import (
	"github.com/maistra/istio-operator/pkg/components/common"
)

type ResourceNames struct {
	common.ResourceNames
	MeshPolicy               string
	DefaultDestinationRule   string
	APIServerDestinationRule string
}

var (
	Names = ResourceNames{
		ResourceNames: common.ResourceNames{
			Deployment:     "istio-citadel",
			Service:        "istio-citadel",
			Role:           "istio-citadel",
			RoleBinding:    "istio-citadel-admin-role-binding",
			ServiceAccount: "istio-citadel-service-account",
		},
		MeshPolicy:               "default",
		DefaultDestinationRule:   "default",
		APIServerDestinationRule: "api-server",
	}
)
