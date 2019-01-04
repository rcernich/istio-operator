package galley

import (
	"bytes"
	"fmt"
	"text/template"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	istioopv1alpha1 "github.com/maistra/istio-operator/pkg/apis/istio/v1alpha1"
	"github.com/maistra/istio-operator/pkg/components/common"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	resourceread "github.com/openshift/library-go/pkg/operator/resource/resourcecread"
	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
)

func SyncGalley(config *istioopv1alpha1.IstioOperatorConfig) []error {

	templateParams := galleyTemplateParams{
		CommonTemplateParams: common.CommonTemplateParams{
			Namespace:              config.Namespace,
			ServiceAccountName:     "istio-galley-service-account",
			ClusterRoleName:        "istio-galley-" + config.Namespace,
			ClusterRoleBindingName: "istio-galley-admin-role-binding-" + config.Namespace,
		},
		ConfigureValidation:         config.Spec.GeneralConfig.ConfigValidation,
		ControlPlaneSecurityEnabled: config.Spec.GeneralConfig.ControlPlaneSecurityEnabled,
		MonitoringPort:              config.Spec.GeneralConfig.MonitoringPort,
		PriorityClassName:           config.Spec.GeneralConfig.PriorityClassName,
		ReplicaCount:                config.Spec.GalleyConfig.ReplicaCount,
	}

	templates := GalleyTemplates()

	kubeClient := k8sclient.GetKubeClient()
	errors := []error{}
	var err error
	var data *bytes.Buffer

	// create roles
	var role *rbacv1.ClusterRole
	data, err = processTemplate(templates.ClusterRoleTemplate, &templateParams)
	if err == nil {
		role = resourceread.ReadClusterRoleV1OrDie(data.Bytes())
		_, _, err = resourceapply.ApplyClusterRole(kubeClient.RbacV1(), role)
	}
	if err != nil {
		errors = append(errors, fmt.Errorf("Galley: ClusterRole: %v", err))
		err = nil
	}

	// create service account for galley pods
	var serviceAccount *corev1.ServiceAccount
	data, err = processTemplate(templates.ServiceAccountTemplate, &templateParams)
	if err == nil {
		serviceAccount = resourceread.ReadServiceAccountV1OrDie(data.Bytes())
		_, _, err = resourceapply.ApplyServiceAccount(kubeClient.CoreV1(), serviceAccount)
	}
	if err != nil {
		errors = append(errors, fmt.Errorf("Galley: ServiceAccount: %v", err))
		err = nil
	}

	// create role binding
	var roleBinding *rbacv1.ClusterRoleBinding
	data, err = processTemplate(templates.ClusterRoleBindingTemplate, &templateParams)
	if err == nil {
		roleBinding = resourceread.ReadClusterRoleBindingV1OrDie(data.Bytes())
		_, _, err = resourceapply.ApplyClusterRoleBinding(kubeClient.RbacV1(), roleBinding)
	}
	if err != nil {
		errors = append(errors, fmt.Errorf("Galley: ClusterRoleBinding: %v", err))
		err = nil
	}

	// create service
	var service *corev1.Service
	data, err = processTemplate(templates.ServiceTemplate, &templateParams)
	if err == nil {
		service = resourceread.ReadServiceV1OrDie(data.Bytes())
		_, _, err = resourceapply.ApplyService(kubeClient.CoreV1(), service)
	}
	if err != nil {
		errors = append(errors, fmt.Errorf("Galley: Service: %v", err))
		err = nil
	}

	// create config map
	var configMap *corev1.ConfigMap
	configMapModified := false
	data, err = processTemplate(templates.ConfigMapTemplate, &templateParams)
	if err == nil {
		configMap = resourceread.ReadConfigMapV1OrDie(data.Bytes())
		_, configMapModified, err = resourceapply.ApplyConfigMap(kubeClient.CoreV1(), configMap)
	}
	if err != nil {
		errors = append(errors, fmt.Errorf("Galley: ConfigMap: %v", err))
		err = nil
	}

	// create deployment
	var deployment *appsv1.Deployment
	data, err = processTemplate(templates.DeploymentTemplate, &templateParams)
	if err == nil {
		deployment = resourceread.ReadDeploymentV1OrDie(data.Bytes())
		_, _, err = resourceapply.ApplyDeployment(kubeClient.AppsV1(), deployment, config.Status.ObservedGeneration, configMapModified)
	}
	if err != nil {
		errors = append(errors, fmt.Errorf("Galley: Deployment: %v", err))
		err = nil
	}

	return errors
}

func processTemplate(template *template.Template, params interface{}) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	err := template.Execute(&buf, params)
	return &buf, err
}
