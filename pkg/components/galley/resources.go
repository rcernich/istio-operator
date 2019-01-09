package galley

import (
	"sync"
	"text/template"

	"github.com/maistra/istio-operator/pkg/components/common"
)

type templateParams struct {
	common.TemplateParams
	PriorityClassName           string
	MonitoringPort              int
	ControlPlaneSecurityEnabled bool
	ConfigureValidation         bool
}

var (
	_singleton *common.Templates
	_init      sync.Once
)

func TemplatesInstance() *common.Templates {
	_init.Do(func() {
		commonTemplates := common.TemplatesInstance()
		_singleton = &common.Templates{
			ServiceAccountTemplate:     commonTemplates.ServiceAccountTemplate,
			ClusterRoleBindingTemplate: commonTemplates.ClusterRoleBindingTemplate,
			ServiceTemplate:            template.New("Service.yaml"),
			DeploymentTemplate:         template.New("Deployment.yaml"),
			ClusterRoleTemplate:        template.New("ClusterRole.yaml"),
			ConfigMapTemplate:          template.New("ConfigMap.yaml"),
		}
		_singleton.ServiceTemplate.Parse(serviceYamlTemplate)
		_singleton.DeploymentTemplate.Parse(deploymentYamlTemplate)
		_singleton.ClusterRoleTemplate.Parse(clusterRoleYamlTemplate)
		_singleton.ConfigMapTemplate.Parse(configMapYamlTemplate)
	})
	return _singleton
}

const serviceYamlTemplate = `
apiVersion: v1
kind: Service
metadata:
  name: istio-galley
  namespace: {{ .Namespace }}
  labels:
    istio: galley
spec:
  ports:
  - port: 443
    name: https-validation
  - port: {{ .MonitoringPort }}
    name: http-monitoring
  - port: 9901
    name: grpc-mcp
  selector:
    istio: galley
`

const deploymentYamlTemplate = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: istio-galley
  namespace: {{ .Namespace }}
  labels:
    istio: galley
spec:
  replicas: {{ .ReplicaCount }}
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        istio: galley
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      serviceAccountName: {{ .ServiceAccountName }}
{{- if .PriorityClassName }}
      priorityClassName: "{{ .PriorityClassName }}"
{{- end }}
      containers:
        - name: galley
          image: "{{ .Values.global.hub }}/{{ .Values.image }}:{{ .Values.global.tag }}"
          imagePullPolicy: {{ .Values.global.imagePullPolicy }}
          ports:
          - name: https-validation
            containerPort: 443
          - name: http-monitoring
            containerPort: {{ .MonitoringPort }}
          - name: grpc-mcp
            containerPort: 9901
          command:
          - /usr/local/bin/galley
          - --meshConfigFile=/etc/istio/mesh-config/mesh
          - --caCertFile=/etc/istio/certs/root-cert.pem
          - --tlsCertFile=/etc/istio/certs/cert-chain.pem
          - --tlsKeyFile=/etc/istio/certs/key.pem
          - --livenessProbeInterval=1s
          - --livenessProbePath=/healthliveness
          - --readinessProbePath=/healthready
          - --readinessProbeInterval=1s
{{- if $.ControlPlaneSecurityEnabled}}
          - --insecure=false
{{- else }}
          - --insecure=true
{{- end }}
          - --validation-webhook-config-file
          - /etc/istio/config/validatingwebhookconfiguration.yaml
          - --monitoringPort={{ .MonitoringPort }}
          volumeMounts:
          - name: certs
            mountPath: /etc/istio/certs
            readOnly: true
          - name: config
            mountPath: /etc/istio/config
            readOnly: true
          - name: mesh-config
            mountPath: /etc/istio/mesh-config
            readOnly: true
          livenessProbe:
            exec:
              command:
                - /usr/local/bin/galley
                - probe
                - --probe-path=/healthliveness
                - --interval=10s
            initialDelaySeconds: 5
            periodSeconds: 5
          readinessProbe:
            exec:
              command:
                - /usr/local/bin/galley
                - probe
                - --probe-path=/healthready
                - --interval=10s
            initialDelaySeconds: 5
            periodSeconds: 5
          resources:
{{- if .Values.resources }}
{{ toYaml .Values.resources | indent 12 }}
{{- else }}
{{ toYaml .Values.global.defaultResources | indent 12 }}
{{- end }}
      volumes:
      - name: certs
        secret:
          secretName: istio.istio-galley-service-account
      - name: config
        configMap:
          name: istio-galley-configuration
      - name: mesh-config
        configMap:
          name: istio
      affinity:
      {{- include "nodeaffinity" . | indent 6 }}
`

const clusterRoleYamlTemplate = `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: istio-galley-{{ .Namespace }}
  labels:
    app: galley
rules:
- apiGroups: ["admissionregistration.k8s.io"]
  resources: ["validatingwebhookconfigurations"]
  verbs: ["*"]
- apiGroups: ["config.istio.io"] # istio mixer CRD watcher
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["networking.istio.io"]
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["authentication.istio.io"]
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["rbac.istio.io"]
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["*"]
  resources: ["deployments"]
  resourceNames: ["istio-galley"]
  verbs: ["get"]
- apiGroups: ["*"]
  resources: ["endpoints"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["extensions"]
  resources: ["ingresses"]
  verbs: ["get", "list", "watch"]
`

const configMapYamlTemplate = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-galley-configuration
  namespace: {{ .Namespace }}
  labels:
    istio: galley
data:
  validatingwebhookconfiguration.yaml: |-
    apiVersion: admissionregistration.k8s.io/v1beta1
    kind: ValidatingWebhookConfiguration
    metadata:
      name: istio-galley
      namespace: {{ .Namespace }}
      labels:
        istio: galley
    webhooks:
    {{- if .ConfigureValidation }}
      - name: pilot.validation.istio.io
        clientConfig:
          service:
            name: istio-galley
            namespace: {{ .Namespace }}
            path: "/admitpilot"
          caBundle: ""
        rules:
          - operations:
            - CREATE
            - UPDATE
            apiGroups:
            - config.istio.io
            apiVersions:
            - v1alpha2
            resources:
            - httpapispecs
            - httpapispecbindings
            - quotaspecs
            - quotaspecbindings
          - operations:
            - CREATE
            - UPDATE
            apiGroups:
            - rbac.istio.io
            apiVersions:
            - "*"
            resources:
            - "*"
          - operations:
            - CREATE
            - UPDATE
            apiGroups:
            - authentication.istio.io
            apiVersions:
            - "*"
            resources:
            - "*"
          - operations:
            - CREATE
            - UPDATE
            apiGroups:
            - networking.istio.io
            apiVersions:
            - "*"
            resources:
            - destinationrules
            - envoyfilters
            - gateways
            - serviceentries
            - virtualservices
        failurePolicy: Fail
      - name: mixer.validation.istio.io
        clientConfig:
          service:
            name: istio-galley
            namespace: {{ .Namespace }}
            path: "/admitmixer"
          caBundle: ""
        rules:
          - operations:
            - CREATE
            - UPDATE
            apiGroups:
            - config.istio.io
            apiVersions:
            - v1alpha2
            resources:
            - rules
            - attributemanifests
            - circonuses
            - deniers
            - fluentds
            - kubernetesenvs
            - listcheckers
            - memquotas
            - noops
            - opas
            - prometheuses
            - rbacs
            - servicecontrols
            - solarwindses
            - stackdrivers
            - cloudwatches
            - dogstatsds
            - statsds
            - stdios
            - apikeys
            - authorizations
            - checknothings
            # - kuberneteses
            - listentries
            - logentries
            - metrics
            - quotas
            - reportnothings
            - servicecontrolreports
            - tracespans
        failurePolicy: Fail
    {{- end }}
  accesslist.yaml: |-
    allowed:
        - spiffe://cluster.local/ns/{{ .Namespace }}/sa/istio-mixer-service-account
        - spiffe://cluster.local/ns/{{ .Namespace }}/sa/istio-pilot-service-account
`
