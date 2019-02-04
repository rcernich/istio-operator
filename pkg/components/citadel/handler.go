package citadel

import (
	"bytes"
	"fmt"

	"k8s.io/client-go/dynamic"

	istioopv1alpha1 "github.com/maistra/istio-operator/pkg/apis/istio/v1alpha1"
	"github.com/maistra/istio-operator/pkg/components/common"
	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
)

func Sync(config *istioopv1alpha1.IstioOperatorConfig) []error {

	templateParams := templateParams{
		TemplateParams: common.TemplateParams{
			Namespace:              config.Namespace,
			ReplicaCount:           *config.Spec.GalleyConfig.ReplicaCount,
			ServiceAccountName:     "istio-citadel-service-account",
			ClusterRoleName:        "istio-citadel-" + config.Namespace,
			ClusterRoleBindingName: "istio-citadel-admin-role-binding-" + config.Namespace,
		},
		ConfigureValidation:         config.Spec.GeneralConfig.ConfigValidation,
		ControlPlaneSecurityEnabled: config.Spec.GeneralConfig.ControlPlaneSecurityEnabled,
		MonitoringPort:              *config.Spec.GeneralConfig.MonitoringPort,
		PriorityClassName:           *config.Spec.GeneralConfig.PriorityClassName,
	}

	templates := TemplatesInstance()
	errors := common.Sync(config, "Citadel", &templates.Templates, templateParams)

	var err error
	var data *bytes.Buffer

	// XXX: add labels
	// XXX: add resource limits
	// XXX: add image pull secrets
	// XXX: add image pull policy
	// XXX: add ownership metadata

	dynamicInterface, err := dynamic.NewForConfig(k8sclient.GetKubeConfig())

	// XXX: distinguish between Policy and MeshPolicy for cluster vs namespaced installations
	if config.Spec.GeneralConfig.MTlsEnabled {
		data, err = common.ProcessTemplate(templates.MtlsMeshPolicyTemplate, &templateParams)
		if err == nil {
			meshPolicy := common.ReadMeshPolicyV1Alpha1OrDie(data.Bytes())
			_, _, err = common.ApplyMeshPolicy(dynamicInterface, meshPolicy)
		}
		if err != nil {
			errors = append(errors, fmt.Errorf("Citadel: MeshPolicy: %v", err))
		}

		data, err = common.ProcessTemplate(templates.MtlsDestinationRuleListTemplate, &templateParams)
		if err == nil {
			destinationRules := common.ReadDestinationRuleListV1Alpha3OrDie(data.Bytes())
			for _, dr := range destinationRules.Items {
				_, _, err = common.ApplyDestinationRule(dynamicInterface, &dr)
				if err != nil {
					errors = append(errors, fmt.Errorf("Citadel: DestinationRule: %v", err))
					err = nil
				}
			}
		}
	} else {
		data, err = common.ProcessTemplate(templates.PermissiveMeshPolicyTemplate, &templateParams)
		if err == nil {
			meshPolicy := common.ReadMeshPolicyV1Alpha1OrDie(data.Bytes())
			_, _, err = common.ApplyMeshPolicy(dynamicInterface, meshPolicy)
		}
		if err != nil {
			errors = append(errors, fmt.Errorf("Citadel: MeshPolicy: %v", err))
		}
	}

	return errors
}
