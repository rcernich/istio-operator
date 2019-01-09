package galley

import (
	istioopv1alpha1 "github.com/maistra/istio-operator/pkg/apis/istio/v1alpha1"
	"github.com/maistra/istio-operator/pkg/components/common"
)

func Sync(config *istioopv1alpha1.IstioOperatorConfig) []error {

	templateParams := templateParams{
		TemplateParams: common.TemplateParams{
			Namespace:              config.Namespace,
			ReplicaCount:           *config.Spec.GalleyConfig.ReplicaCount,
			ServiceAccountName:     "istio-galley-service-account",
			ClusterRoleName:        "istio-galley-" + config.Namespace,
			ClusterRoleBindingName: "istio-galley-admin-role-binding-" + config.Namespace,
		},
		ConfigureValidation:         config.Spec.GeneralConfig.ConfigValidation,
		ControlPlaneSecurityEnabled: config.Spec.GeneralConfig.ControlPlaneSecurityEnabled,
		MonitoringPort:              *config.Spec.GeneralConfig.MonitoringPort,
		PriorityClassName:           *config.Spec.GeneralConfig.PriorityClassName,
	}

	return common.Sync(config, "Galley", TemplatesInstance(), templateParams)
}
