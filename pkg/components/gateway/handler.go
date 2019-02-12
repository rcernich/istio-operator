package gateway

import (
	istioopv1alpha2 "github.com/maistra/istio-operator/pkg/apis/istio/v1alpha2"
	"github.com/maistra/istio-operator/pkg/components/common"
)

func Sync(config *istioopv1alpha2.IstioControlPlane) []error {

	templateParams := common.TemplateParams{
		Config:                 config,
		ServiceAccountName:     "istio-ingressgateway-service-account",
		ClusterRoleName:        "istio-ingressgateway-" + config.Namespace,
		ClusterRoleBindingName: "istio-ingressgateway-" + config.Namespace,
	}

	templates := TemplatesInstance()
	errors := common.Sync(config, "Gateways: Ingress", &templates.Ingress, templateParams)

	templateParams.ServiceAccountName = "istio-egressgateway-service-account"
	templateParams.ClusterRoleName = "istio-egressgateway-" + config.Namespace
	templateParams.ClusterRoleBindingName = "istio-egressgateway-" + config.Namespace
	errors = append(errors, common.Sync(config, "Gateways: Egress", &templates.Egress, templateParams)...)

	return errors
}
