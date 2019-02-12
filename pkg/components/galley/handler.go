package galley

import (
	istioopv1alpha2 "github.com/maistra/istio-operator/pkg/apis/istio/v1alpha2"
	"github.com/maistra/istio-operator/pkg/components/common"
)

func Sync(config *istioopv1alpha2.IstioControlPlane) []error {

	templateParams := templateParams{
		TemplateParams: common.TemplateParams{
			Config:                 config,
			ServiceAccountName:     "istio-galley-service-account",
			ClusterRoleName:        "istio-galley-" + config.Namespace,
			ClusterRoleBindingName: "istio-galley-admin-role-binding-" + config.Namespace,
		},
	}

	return common.Sync(config, "Galley", TemplatesInstance(), templateParams)
}
