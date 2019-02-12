package pilot

import (
	istioopv1alpha2 "github.com/maistra/istio-operator/pkg/apis/istio/v1alpha2"
	"github.com/maistra/istio-operator/pkg/components/common"
)

func Sync(config *istioopv1alpha2.IstioControlPlane) []error {

	templateParams := templateParams{
		TemplateParams: common.TemplateParams{
			Config:                 config,
			ServiceAccountName:     "istio-pilot-service-account",
			ClusterRoleName:        "istio-pilot-" + config.Namespace,
			ClusterRoleBindingName: "istio-pilot-" + config.Namespace,
		},
	}

	return common.Sync(config, "Pilot", TemplatesInstance(), templateParams)
}
