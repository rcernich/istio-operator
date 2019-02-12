package sidecar

import (
	"bytes"
	"fmt"

	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	admissionregistrationclientv1beta1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1beta1"

	istioopv1alpha2 "github.com/maistra/istio-operator/pkg/apis/istio/v1alpha2"
	"github.com/maistra/istio-operator/pkg/components/common"
	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
)

var (
	admissionScheme = runtime.NewScheme()
	admissionCodecs = serializer.NewCodecFactory(admissionScheme)
)

func init() {
	if err := admissionregistrationv1beta1.AddToScheme(admissionScheme); err != nil {
		panic(err)
	}
}

func Sync(config *istioopv1alpha2.IstioControlPlane) []error {

	templateParams := TemplateParams{
		TemplateParams: common.TemplateParams{
			Config:                 config,
			ServiceAccountName:     "istio-sidecar-injector-service-account",
			ClusterRoleName:        "istio-sidecar-injector-" + config.Namespace,
			ClusterRoleBindingName: "istio-sidecar-injector-admin-role-binding-" + config.Namespace,
		},
	}

	templates := TemplatesInstance()
	errors := common.Sync(config, "SideCarInjector", &templates.Templates, templateParams)

	kubeClient := k8sclient.GetKubeClient()
	var err error
	var data *bytes.Buffer

	// XXX: add labels
	// XXX: add resource limits
	// XXX: add image pull secrets
	// XXX: add image pull policy
	// XXX: add ownership metadata

	// create mutating webhook
	data, err = common.ProcessTemplate(templates.MutatingWebHookTemplate, &templateParams)
	if err == nil {
		mutatingWebhook := readMutatingWebhookConfigurationV1Beta1OrDie(data.Bytes())
		_, _, err = applyMutatingWebhook(kubeClient.AdmissionregistrationV1beta1(), mutatingWebhook)
	}
	if err != nil {
		errors = append(errors, fmt.Errorf("SideCarInjector: MutatingWebhook: %v", err))
		err = nil
	}

	return errors
}

func readMutatingWebhookConfigurationV1Beta1OrDie(objBytes []byte) *admissionregistrationv1beta1.MutatingWebhookConfiguration {
	requiredObj, err := runtime.Decode(admissionCodecs.UniversalDecoder(admissionregistrationv1beta1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	return requiredObj.(*admissionregistrationv1beta1.MutatingWebhookConfiguration)
}

func applyMutatingWebhook(client admissionregistrationclientv1beta1.MutatingWebhookConfigurationsGetter, required *admissionregistrationv1beta1.MutatingWebhookConfiguration) (*admissionregistrationv1beta1.MutatingWebhookConfiguration, bool, error) {
	existing, err := client.MutatingWebhookConfigurations().Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		actual, err := client.MutatingWebhookConfigurations().Create(required)
		return actual, true, err
	}
	if err != nil {
		return nil, false, err
	}

	return existing, false, err
}
