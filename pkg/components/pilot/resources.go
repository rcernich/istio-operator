package pilot

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
		}
		_singleton.ServiceTemplate.Parse(serviceYamlTemplate)
		_singleton.DeploymentTemplate.Parse(deploymentYamlTemplate)
		_singleton.ClusterRoleTemplate.Parse(clusterRoleYamlTemplate)
	})
	return _singleton
}

const serviceYamlTemplate = `
apiVersion: v1
kind: Service
metadata:
  name: istio-pilot
  namespace: {{ .Release.Namespace }}
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
  - port: {{ .Values.global.monitoringPort }}
    name: http-monitoring
  selector:
    istio: pilot
`

const deploymentYamlTemplate = `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-pilot
  namespace: {{ .Release.Namespace }}
  # TODO: default template doesn't have this, which one is right ?
  labels:
    app: pilot
    istio: pilot
  annotations:
    checksum/config-volume: {{ template "istio.configmap.checksum" . }}
spec:
  replicas: {{ .Values.replicaCount }}
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
{{- if .Values.global.priorityClassName }}
      priorityClassName: "{{ .Values.global.priorityClassName }}"
{{- end }}
      containers:
        - name: discovery
{{- if contains "/" .Values.image }}
          image: "{{ .Values.image }}"
{{- else }}
          image: "{{ .Values.global.hub }}/{{ .Values.image }}:{{ .Values.global.tag }}"
{{- end }}
          imagePullPolicy: {{ .Values.global.imagePullPolicy }}
          args:
          - "discovery"
          - --monitoringAddr=:{{ .Values.global.monitoringPort }}
{{- if $.Values.global.proxy.discoveryDomain }}
          - --domain
          - {{ $.Values.global.proxy.discoveryDomain }}
{{- end }}
{{- if .Values.global.oneNamespace }}
          - "-a"
          - {{ .Release.Namespace }}
{{- end }}
{{- if not .Values.sidecar }}
          - --secureGrpcAddr
          - ":15011"
{{- end }}
{{- if $.Values.global.useMCP }}
    {{- if $.Values.global.controlPlaneSecurityEnabled}}
          - --mcpServerAddrs=mcps://istio-galley.{{ $.Release.Namespace }}.svc:9901
          - --certFile=/etc/certs/cert-chain.pem
          - --keyFile=/etc/certs/key.pem
          - --caCertFile=/etc/certs/root-cert.pem
    {{- else }}
          - --mcpServerAddrs=mcp://istio-galley.{{ $.Release.Namespace }}.svc:9901
    {{- end }}
{{- end }}
          ports:
          - containerPort: 8080
          - containerPort: 15010
{{- if not .Values.sidecar }}
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
          {{- if .Values.env }}
          {{- range $key, $val := .Values.env }}
          - name: {{ $key }}
            value: "{{ $val }}"
          {{- end }}
          {{- end }}
{{- if .Values.traceSampling }}
          - name: PILOT_TRACE_SAMPLING
            value: "{{ .Values.traceSampling }}"
{{- end }}
          resources:
{{- if .Values.resources }}
{{ toYaml .Values.resources | indent 12 }}
{{- else }}
{{ toYaml .Values.global.defaultResources | indent 12 }}
{{- end }}
          volumeMounts:
          - name: config-volume
            mountPath: /etc/istio/config
          - name: istio-certs
            mountPath: /etc/certs
            readOnly: true
{{- if .Values.sidecar }}
        - name: istio-proxy
{{- if contains "/" .Values.global.proxy.image }}
          image: "{{ .Values.global.proxy.image }}"
{{- else }}
          image: "{{ .Values.global.hub }}/{{ .Values.global.proxy.image }}:{{ .Values.global.tag }}"
{{- end }}
          imagePullPolicy: {{ .Values.global.imagePullPolicy }}
          ports:
          - containerPort: 15003
          - containerPort: 15005
          - containerPort: 15007
          - containerPort: 15011
          args:
          - proxy
{{- if $.Values.global.proxy.proxyDomain }}
          - --domain
          - {{ $.Values.global.proxy.proxyDomain }}
{{- end }}
          - --serviceCluster
          - istio-pilot
          - --templateFile
          - /etc/istio/proxy/envoy_pilot.yaml.tmpl
        {{- if $.Values.global.controlPlaneSecurityEnabled}}
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
{{- if .Values.global.proxy.resources }}
{{ toYaml .Values.global.proxy.resources | indent 12 }}
{{- else }}
{{ toYaml .Values.global.defaultResources | indent 12 }}
{{- end }}
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
      {{- include "nodeaffinity" . | indent 6 }}
`

const clusterRoleYamlTemplate = `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: istio-pilot-{{ .Release.Namespace }}
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
