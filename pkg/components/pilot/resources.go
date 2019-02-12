package pilot

import (
	"sync"
	"text/template"

	"github.com/maistra/istio-operator/pkg/components/common"

	corev1 "k8s.io/api/core/v1"
)

type templateParams struct {
	common.TemplateParams
	PriorityClassName           string
	MonitoringPort              int
	ControlPlaneSecurityEnabled bool
	ConfigureValidation         bool
	DiscoveryDomain             string
	Sidecar                     bool
	UseMCP                      bool
  Env                         []corev1.EnvVar
  TraceSampling               string // TODO
  Resources                   string // TODO
  ProxyImage                  string
  ProxyDomain                 string
  NodeAffinity                string // TODO
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
		}
		_singleton.ServiceTemplate.Parse(serviceYamlTemplate)
		_singleton.DeploymentTemplate.Parse(deploymentYamlTemplate)
		_singleton.ClusterRoleTemplate.Parse(clusterRoleYamlTemplate)
	})
	return _singleton
}

// XXX: ignoring mesh expansion for now

const serviceYamlTemplate = `
apiVersion: v1
kind: Service
metadata:
  name: istio-pilot
  namespace: {{ .Config.Namespace }}
  labels:
    app: pilot
    istio: pilot
spec:
  ports:
  - port: 15010
    name: grpc-xds # direct
  - port: 15011
    name: https-xds # mTLS
  - port: 8080
    name: http-legacy-discovery # direct
  - port: {{ .Config.Spec.Monitoring.Port }}
    name: http-monitoring
  selector:
    istio: pilot
`

const deploymentYamlTemplate = `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-pilot
  namespace: {{ .Config.Namespace }}
  # TODO: default template doesn't have this, which one is right ?
  labels:
    app: pilot
    istio: pilot
  annotations:
    checksum/config-volume: {{ template "istio.configmap.checksum" . }}
spec:
  replicas: {{ .Config.Spec.Pilot.ReplicaCount }}
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app: pilot
        istio: pilot
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      serviceAccountName: istio-pilot-service-account
{{- if .Config.Spec.General.PriorityClassName }}
      priorityClassName: "{{ .Config.Spec.General.PriorityClassName }}"
{{- end }}
      containers:
        - name: discovery
        image: "{{ .Config.Spec.Pilot.Image }}"
        imagePullPolicy: {{ .Config.Spec.General.PullPolicy }}
        args:
          - "discovery"
          - --monitoringAddr=:{{ .Config.Spec.Monitoring.Port }}
{{- if .Config.Spec.Proxy.DiscoveryDomain }}
          - --domain
          - {{ .Config.Spec.Proxy.DiscoveryDomain }}
{{- end }}
{{- if .Config.Spec.Pilot.WatchedNamespaces }}
          {{- range $namespace := .Config.Spec.Pilot.WatchedNamespaces }}
          - "-a"
          - {{ $namespace }}
          {{- end }}
{{- end }}
{{- if not .Sidecar }}
          - --secureGrpcAddr
          - ":15011"
{{- end }}
{{- if .UseMCP }}
    {{- if .Config.Spec.Security.ControlPlaneSecurityEnabled}}
          - --mcpServerAddrs=mcps://istio-galley.{{ .Config.Namespace }}.svc:9901
          - --certFile=/etc/certs/cert-chain.pem
          - --keyFile=/etc/certs/key.pem
          - --caCertFile=/etc/certs/root-cert.pem
    {{- else }}
          - --mcpServerAddrs=mcp://istio-galley.{{ .Config.Namespace }}.svc:9901
    {{- end }}
{{- end }}
          ports:
          - containerPort: 8080
          - containerPort: 15010
{{- if not .Config.Spec.Pilot.Sidecar }}
          - containerPort: 15011
{{- end }}
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 30
            timeoutSeconds: 5
          env:
          - name: POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.namespace
          {{- if .Config.Spec.Pilot.PushThrottleCount }}
          - name: PILOT_PUSH_THROTTLE_COUNT
            value: "{{ Config.Spec.Pilot.PushThrottleCount }}"
          {{- end }}
{{- if .Config.Spec.Pilot.RandomTraceSampling }}
          - name: PILOT_TRACE_SAMPLING
            value: "{{ .Config.Spec.Pilot.RandomTraceSampling }}"
{{- end }}
          resources:
          volumeMounts:
          - name: config-volume
            mountPath: /etc/istio/config
          - name: istio-certs
            mountPath: /etc/certs
            readOnly: true
{{- if .Config.Spec.Pilot.Sidecar }}
        - name: istio-proxy
          image: "{{ .Config.Spec.Proxy.Image }}"
          imagePullPolicy: {{ .Config.Spec.General.PullPolicy }}
          ports:
          - containerPort: 15003
          - containerPort: 15005
          - containerPort: 15007
          - containerPort: 15011
          args:
          - proxy
{{- if .Config.Spec.Proxy.ProxyDomain }}
          - --domain
          - {{ .Config.Spec.Proxy.ProxyDomain }}
{{- end }}
          - --serviceCluster
          - istio-pilot
          - --templateFile
          - /etc/istio/proxy/envoy_pilot.yaml.tmpl
        {{- if $.Config.Spec.SecurityControlPlaneSecurityEnabled }}
          - --controlPlaneAuthPolicy
          - MUTUAL_TLS
        {{- else }}
          - --controlPlaneAuthPolicy
          - NONE
        {{- end }}
          env:
          - name: POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.namespace
          - name: INSTANCE_IP
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: status.podIP
          resources:
          volumeMounts:
          - name: istio-certs
            mountPath: /etc/certs
            readOnly: true
{{- end }}
      volumes:
      - name: config-volume
        configMap:
          name: istio
      - name: istio-certs
        secret:
          secretName: istio.istio-pilot-service-account
          optional: true
      affinity:
`

const clusterRoleYamlTemplate = `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: {{ .ClusterRoleName }}
  labels:
    app: pilot
rules:
- apiGroups: ["config.istio.io"]
  resources: ["*"]
  verbs: ["*"]
- apiGroups: ["rbac.istio.io"]
  resources: ["*"]
  verbs: ["get", "watch", "list"]
- apiGroups: ["networking.istio.io"]
  resources: ["*"]
  verbs: ["*"]
- apiGroups: ["authentication.istio.io"]
  resources: ["*"]
  verbs: ["*"]
- apiGroups: ["apiextensions.k8s.io"]
  resources: ["customresourcedefinitions"]
  verbs: ["*"]
- apiGroups: ["extensions"]
  resources: ["ingresses", "ingresses/status"]
  verbs: ["*"]
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["create", "get", "list", "watch", "update"]
- apiGroups: [""]
  resources: ["endpoints", "pods", "services"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["namespaces", "nodes", "secrets"]
  verbs: ["get", "list", "watch"]
`
