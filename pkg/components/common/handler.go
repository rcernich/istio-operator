package common

import (
	"bytes"
	"fmt"
	"text/template"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"

	istioopv1alpha2 "github.com/maistra/istio-operator/pkg/apis/istio/v1alpha2"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	resourceread "github.com/openshift/library-go/pkg/operator/resource/resourcecread"
	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
)

var (
	authenticationV1alpha1GV = schema.GroupVersion{Group: "authentication.istio.io", Version: "v1alpha1"}
	networkingV1alpha3GV     = schema.GroupVersion{Group: "networking.istio.io", Version: "v1alpha3"}
	destinationRuleGVK       = networkingV1alpha3GV.WithKind("DestinationRule")
	destinationRuleListGVK   = networkingV1alpha3GV.WithKind("DestinationRuleList")
	destinationRuleResource  = "destinationrules"
	policyGVK                = authenticationV1alpha1GV.WithKind("Policy")
	meshPolicyGVK            = authenticationV1alpha1GV.WithKind("MeshPolicy")
	meshPolicyResource       = "meshpolicies"
	dynamicScheme            = runtime.NewScheme()
	dynamicCodecs            = serializer.NewCodecFactory(dynamicScheme)
)

func init() {
	dynamicScheme.AddKnownTypeWithName(destinationRuleGVK, &unstructured.Unstructured{})
	dynamicScheme.AddKnownTypeWithName(destinationRuleListGVK, &unstructured.UnstructuredList{})
	dynamicScheme.AddKnownTypeWithName(meshPolicyGVK, &unstructured.Unstructured{})
}

func Sync(config *istioopv1alpha2.IstioControlPlane, component string, templates *Templates, templateParams interface{}) []error {

	kubeClient := k8sclient.GetKubeClient()
	errors := []error{}
	var err error
	var data *bytes.Buffer

	// XXX: add labels
	// XXX: add resource limits
	// XXX: add image pull secrets
	// XXX: add image pull policy
	// XXX: add ownership metadata

	// create roles
	if templates.ClusterRoleTemplate != nil {
		data, err = ProcessTemplate(templates.ClusterRoleTemplate, &templateParams)
		if err == nil {
			role := resourceread.ReadClusterRoleV1OrDie(data.Bytes())
			_, _, err = resourceapply.ApplyClusterRole(kubeClient.RbacV1(), role)
		}
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: ClusterRole: %v", component, err))
			err = nil
		}
	}

	// create service account for galley pods
	if templates.ServiceAccountTemplate != nil {
		data, err = ProcessTemplate(templates.ServiceAccountTemplate, &templateParams)
		if err == nil {
			serviceAccount := resourceread.ReadServiceAccountV1OrDie(data.Bytes())
			_, _, err = resourceapply.ApplyServiceAccount(kubeClient.CoreV1(), serviceAccount)
		}
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: ServiceAccount: %v", component, err))
			err = nil
		}
	}

	// create role binding
	if templates.ClusterRoleBindingTemplate != nil {
		data, err = ProcessTemplate(templates.ClusterRoleBindingTemplate, &templateParams)
		if err == nil {
			roleBinding := resourceread.ReadClusterRoleBindingV1OrDie(data.Bytes())
			_, _, err = resourceapply.ApplyClusterRoleBinding(kubeClient.RbacV1(), roleBinding)
		}
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: ClusterRoleBinding: %v", component, err))
			err = nil
		}
	}

	// create service
	if templates.ServiceTemplate != nil {
		data, err = ProcessTemplate(templates.ServiceTemplate, &templateParams)
		if err == nil {
			service := resourceread.ReadServiceV1OrDie(data.Bytes())
			_, _, err = resourceapply.ApplyService(kubeClient.CoreV1(), service)
		}
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: Service: %v", component, err))
			err = nil
		}
	}

	// create config map
	configMapModified := false
	if templates.ConfigMapTemplate != nil {
		data, err = ProcessTemplate(templates.ConfigMapTemplate, &templateParams)
		if err == nil {
			configMap := resourceread.ReadConfigMapV1OrDie(data.Bytes())
			_, configMapModified, err = resourceapply.ApplyConfigMap(kubeClient.CoreV1(), configMap)
		}
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: ConfigMap: %v", component, err))
			err = nil
		}
	}

	// create deployment
	if templates.DeploymentTemplate != nil {
		data, err = ProcessTemplate(templates.DeploymentTemplate, &templateParams)
		if err == nil {
			deployment := resourceread.ReadDeploymentV1OrDie(data.Bytes())
			_, _, err = resourceapply.ApplyDeployment(kubeClient.AppsV1(), deployment, config.Status.ObservedGeneration, configMapModified)
		}
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: Deployment: %v", component, err))
			err = nil
		}
	}
	return errors
}

func ProcessTemplate(template *template.Template, params interface{}) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	err := template.Execute(&buf, params)
	return &buf, err
}

func ReadDestinationRuleV1Alpha3OrDie(objBytes []byte) *unstructured.Unstructured {
	return readDynamicObjectOrDie(networkingV1alpha3GV, objBytes)
}

func ReadDestinationRuleListV1Alpha3OrDie(objBytes []byte) *unstructured.UnstructuredList {
	return readDynamicListOrDie(networkingV1alpha3GV, objBytes)
}

func ApplyDestinationRule(client dynamic.DynamicInterface, required *unstructured.Unstructured) (*unstructured.Unstructured, bool, error) {
	return applyDynamicObject(client, networkingV1alpha3GV.WithResource(destinationRuleResource), required, true)
}

func ReadMeshPolicyV1Alpha1OrDie(objBytes []byte) *unstructured.Unstructured {
	return readDynamicObjectOrDie(authenticationV1alpha1GV, objBytes)
}

func ApplyMeshPolicy(client dynamic.DynamicInterface, required *unstructured.Unstructured) (*unstructured.Unstructured, bool, error) {
	return applyDynamicObject(client, authenticationV1alpha1GV.WithResource(meshPolicyResource), required, false)
}

func readDynamicObjectOrDie(gv schema.GroupVersion, objBytes []byte) *unstructured.Unstructured {
	requiredObj, err := runtime.Decode(dynamicCodecs.UniversalDecoder(gv), objBytes)
	if err != nil {
		panic(err)
	}
	return requiredObj.(*unstructured.Unstructured)
}

func readDynamicListOrDie(gv schema.GroupVersion, objBytes []byte) *unstructured.UnstructuredList {
	requiredObj, err := runtime.Decode(dynamicCodecs.UniversalDecoder(gv), objBytes)
	if err != nil {
		panic(err)
	}
	return requiredObj.(*unstructured.UnstructuredList)
}

func applyDynamicObject(client dynamic.DynamicInterface, gvr schema.GroupVersionResource, required *unstructured.Unstructured, namespaced bool) (*unstructured.Unstructured, bool, error) {
	var resourceInterface dynamic.DynamicResourceInterface
	if namespaced {
		resourceInterface = client.NamespacedResource(gvr, required.GetNamespace())
	} else {
		resourceInterface = client.ClusterResource(gvr)
	}
	existing, err := resourceInterface.Get(required.GetName(), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		actual, err := resourceInterface.Create(required)
		return actual, true, err
	}
	// TODO: merge objects
	existing, err = resourceInterface.Update(required)

	return existing, false, err
}
