package pilot

import (
	istioopv1alpha1 "github.com/maistra/istio-operator/pkg/apis/istio/v1alpha1"
	"github.com/maistra/istio-operator/pkg/components/common"
)

func Sync(config *istioopv1alpha1.IstioOperatorConfig) []error {

	templateParams := templateParams{
		TemplateParams: common.TemplateParams{
			Namespace:              config.Namespace,
			ReplicaCount:           *config.Spec.PilotConfig.ReplicaCount,
			ServiceAccountName:     "istio-ingressgateway-service-account",
			ClusterRoleName:        "istio-ingressgateway-" + config.Namespace,
			ClusterRoleBindingName: "istio-ingressgateway-" + config.Namespace,
		},
	}

	templates := TemplatesInstance()
	errors := common.Sync(config, "Gateways: Ingress", templates, templateParams)

	templateParams.ServiceAccountName = "istio-egressgateway-service-account"
	templateParams.ClusterRoleName = "istio-egressgateway-" + config.Namespace
	templateParams.ClusterRoleBindingName = "istio-egressgateway-" + config.Namespace
	errors = append(errors, common.Sync(config, "Gateways: Egress", templates, templateParams)...)

	return errors
}
