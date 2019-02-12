package pilot

import (
	istioopv1alpha2 "github.com/maistra/istio-operator/pkg/apis/istio/v1alpha2"
	"github.com/maistra/istio-operator/pkg/components/common"
)

func Sync(config *istioopv1alpha2.IstioOperatorConfig) []error {

	templateParams := templateParams{
		TemplateParams: common.TemplateParams{
			Namespace:              config.Namespace,
			ReplicaCount:           *config.Spec.PilotConfig.ReplicaCount,
			ServiceAccountName:     "istio-pilot-service-account",
			ClusterRoleName:        "istio-pilot-" + config.Namespace,
			ClusterRoleBindingName: "istio-pilot-" + config.Namespace,
		},
		ConfigureValidation:         config.Spec.GeneralConfig.ConfigValidation,
		ControlPlaneSecurityEnabled: config.Spec.GeneralConfig.ControlPlaneSecurityEnabled,
		MonitoringPort:              *config.Spec.GeneralConfig.MonitoringPort,
		PriorityClassName:           *config.Spec.GeneralConfig.PriorityClassName,
	}

	return common.Sync(config, "Pilot", TemplatesInstance(), templateParams)
}
