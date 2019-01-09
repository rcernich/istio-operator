package common

import (
	"bytes"
	"fmt"
	"text/template"

	istioopv1alpha1 "github.com/maistra/istio-operator/pkg/apis/istio/v1alpha1"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	resourceread "github.com/openshift/library-go/pkg/operator/resource/resourcecread"
	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
)

func Sync(config *istioopv1alpha1.IstioOperatorConfig, component string, templates *Templates, templateParams interface{}) []error {

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
