package common

import (
	"sync"
	"text/template"

	istioopv1alpha2 "github.com/maistra/istio-operator/pkg/apis/istio/v1alpha2"
)

type TemplateParams struct {
	Config                 *istioopv1alpha2.IstioControlPlane
	ServiceAccountName     string
	ClusterRoleName        string
	ClusterRoleBindingName string
	Autoscaler             AutoscalerParams
}

type AutoscalerParams struct {
	istioopv1alpha2.AutoscalerConfig
	Deployment string
}

type Templates struct {
	ClusterRoleTemplate        *template.Template
	ServiceAccountTemplate     *template.Template
	ClusterRoleBindingTemplate *template.Template
	ServiceTemplate            *template.Template
	DeploymentTemplate         *template.Template
	ConfigMapTemplate          *template.Template
}

var (
	_singleton *Templates
	_init      sync.Once
)

func TemplatesInstance() *Templates {
	_init.Do(func() {
		_singleton = &Templates{
			ServiceAccountTemplate:     template.New("ServiceAccount.yaml"),
			ClusterRoleBindingTemplate: template.New("ClusterRoleBinding.yaml"),
		}
		_singleton.ServiceAccountTemplate.Parse(serviceAccountYamlTemplate)
		_singleton.ClusterRoleBindingTemplate.Parse(clusterRoleBindingYamlTemplate)
	})
	return _singleton
}

const serviceAccountYamlTemplate = `
apiVersion: v1
kind: ServiceAccount
{{- if .Values.global.imagePullSecrets }}
imagePullSecrets:
{{- range .Values.global.imagePullSecrets }}
  - name: {{ . }}
{{- end }}
{{- end }}
metadata:
  name: {{ .ServiceAccountName }}
  namespace: {{ .Config.Namespace }}
`

const clusterRoleBindingYamlTemplate = `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: {{ .ClusterRoleBindingName }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: {{ .ClusterRoleName }}
subjects:
  - kind: ServiceAccount
    name: {{ .ServiceAccountName }}
    namespace: {{ .Config.Namespace }}
`

const horizontalAutoscalerYamlTemplate = `
apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
  name: {.Autoscaler.Deployment}
  namespace: {{ .Config.Namespace }}
  labels:
    app: istio
spec:
  maxReplicas: {{ .Autoscaler.Max }}
  minReplicas: {{ .Autoscaler.Min }}
  scaleTargetRef:
    apiVersion: apps/v1beta1
    kind: Deployment
    name: {.Autoscaler.Deployment}
  metrics:
  - type: Resource
    resource:
      name: cpu
      targetAverageUtilization: {{ .Autoscaler.TargetAverageCPUUtilization }}
`
