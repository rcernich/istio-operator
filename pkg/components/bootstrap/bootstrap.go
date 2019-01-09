package bootstrap

import (
	"bytes"
	"fmt"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apiextensionsclientv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	istioopv1alpha1 "github.com/maistra/istio-operator/pkg/apis/istio/v1alpha1"
	"github.com/maistra/istio-operator/pkg/components/common"
	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
)

var (
	coreScheme = runtime.NewScheme()
	coreCodecs = serializer.NewCodecFactory(coreScheme)
)

func Sync(config *istioopv1alpha1.IstioOperatorConfig) []error {
	extensionsClient, err := apiextensionsclientset.NewForConfig(k8sclient.GetKubeConfig())
	if err != nil {
		panic(err)
	}
	errors := []error{}
	params := TemplateParams{}
	templates := Templates()
	var data *bytes.Buffer

	data, err = common.ProcessTemplate(templates.CRDsTemplate, &params)
	if err == nil {
		list := readCustomResourceDefinitionListV1OrDie(data.Bytes())
		for _, crd := range list.Items {
			_, _, err = applyCRD(extensionsClient.ApiextensionsV1beta1(), &crd)
			if err != nil {
				errors = append(errors, fmt.Errorf("Bootstrapping: CRD: %v: %v", crd.Name, err))
				err = nil
			}
		}
	} else {
		panic(err)
	}
	return errors
}

func readCustomResourceDefinitionListV1OrDie(objBytes []byte) *apiextensionsv1beta1.CustomResourceDefinitionList {
	requiredObj, err := runtime.Decode(coreCodecs.UniversalDecoder(apiextensionsv1beta1.SchemeGroupVersion), objBytes)
	if err != nil {
		panic(err)
	}
	return requiredObj.(*apiextensionsv1beta1.CustomResourceDefinitionList)
}

func applyCRD(client apiextensionsclientv1beta1.CustomResourceDefinitionsGetter, required *apiextensionsv1beta1.CustomResourceDefinition) (*apiextensionsv1beta1.CustomResourceDefinition, bool, error) {
	existing, err := client.CustomResourceDefinitions().Get(required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		actual, err := client.CustomResourceDefinitions().Create(required)
		return actual, true, err
	}
	if err != nil {
		return nil, false, err
	}

	return existing, false, err
}
