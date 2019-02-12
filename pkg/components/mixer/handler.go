package mixer

import (
	"bytes"
	"fmt"

	istioopv1alpha2 "github.com/maistra/istio-operator/pkg/apis/istio/v1alpha2"
	"github.com/maistra/istio-operator/pkg/components/common"
	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
	"k8s.io/client-go/dynamic"
)

func Sync(config *istioopv1alpha2.IstioControlPlane) []error {

	templateParams := templateParams{
		TemplateParams: common.TemplateParams{
			Config:                 config,
			ServiceAccountName:     "istio-mixer-service-account",
			ClusterRoleName:        "istio-mixer-" + config.Namespace,
			ClusterRoleBindingName: "istio-mixer-" + config.Namespace,
		},
	}

	var err error
	var data *bytes.Buffer
	dynamicInterface, err := dynamic.NewForConfig(k8sclient.GetKubeConfig())

	templates := TemplatesInstance()
	errors := common.Sync(config, "Mixer", &templates.Templates, templateParams)

	errors = append(errors, common.Sync(config, "Mixer: Policy", &templates.Policy.Templates, templateParams)...)
	data, err = common.ProcessTemplate(TemplatesInstance().Policy.DestinationRule, &templateParams)
	if err == nil {
		destinationRule := common.ReadDestinationRuleV1Alpha3OrDie(data.Bytes())
		_, _, err = common.ApplyDestinationRule(dynamicInterface, destinationRule)
	}
	if err != nil {
		errors = append(errors, fmt.Errorf("Mixer: Policy: DestinationRule: %v", err))
		err = nil
	}

	errors = append(errors, common.Sync(config, "Mixer: Telemetry", &templates.Telemetry.Templates, templateParams)...)
	data, err = common.ProcessTemplate(TemplatesInstance().Telemetry.DestinationRule, &templateParams)
	if err == nil {
		destinationRule := common.ReadDestinationRuleV1Alpha3OrDie(data.Bytes())
		_, _, err = common.ApplyDestinationRule(dynamicInterface, destinationRule)
	}
	if err != nil {
		errors = append(errors, fmt.Errorf("Mixer: Telemetry: DestinationRule: %v", err))
		err = nil
	}

	return errors
}
