package mixer

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

type mixerTemplates struct {
	common.Templates
	DestinationRule *template.Template
}

type templates struct {
	common.Templates
	Policy    mixerTemplates
	Telemetry mixerTemplates
}

var (
	_singleton *templates
	_init      sync.Once
)

func TemplatesInstance() *templates {
	_init.Do(func() {
		commonTemplates := common.TemplatesInstance()
		_singleton = &templates{
			Templates: common.Templates{
				ServiceAccountTemplate:     commonTemplates.ServiceAccountTemplate,
				ClusterRoleBindingTemplate: commonTemplates.ClusterRoleBindingTemplate,
				ClusterRoleTemplate:        template.New("ClusterRole.yaml"),
			},
			Policy: mixerTemplates{
				Templates: common.Templates{
					ServiceTemplate:    template.New("PolicyService.yaml"),
					DeploymentTemplate: template.New("PolicyDeployment.yaml"),
				},
				DestinationRule: template.New("PolicyDestinationRule.yaml"),
			},
			Telemetry: mixerTemplates{
        Templates: common.Templates{
  				ServiceTemplate:    template.New("TelemetryService.yaml"),
	  			DeploymentTemplate: template.New("TelemetryDeployment.yaml"),
        },
				DestinationRule: template.New("TelemetryDestinationRule.yaml"),
			},
		}
		_singleton.ClusterRoleTemplate.Parse(clusterRoleYamlTemplate)
		_singleton.Policy.ServiceTemplate.Parse(policyServiceYamlTemplate)
		_singleton.Policy.DeploymentTemplate.Parse(policyDeploymentYamlTemplate)
    _singleton.Policy.DestinationRule.Parse(policyDestinationRuleYamlTemplate)
		_singleton.Policy.ServiceTemplate.Parse(telemetryServiceYamlTemplate)
		_singleton.Policy.DeploymentTemplate.Parse(telemetryDeploymentYamlTemplate)
    _singleton.Telemetry.DestinationRule.Parse(telemetryDestinationRuleYamlTemplate)
	})
	return _singleton
}

// XXX: ignoring mesh expansion for now

const policyServiceYamlTemplate = `
apiVersion: v1
kind: Service
metadata:
  name: istio-policy
  namespace: {{ $.Release.Namespace }}
  labels:
    app: mixer
    istio: mixer
spec:
  ports:
  - name: grpc-mixer
    port: 9091
  - name: grpc-mixer-mtls
    port: 15004
  - name: http-monitoring
    port: {{ $.Values.global.monitoringPort }}
  selector:
    istio: mixer
    istio-mixer-type: policy
`

const policyDestinationRuleYamlTemplate = `
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: istio-policy
  namespace: {{ .Release.Namespace }}
  labels:
    app: mixer
spec:
  host: istio-policy.{{ .Release.Namespace }}.svc.cluster.local
  trafficPolicy:
    {{- if .Values.global.controlPlaneSecurityEnabled }}
    portLevelSettings:
    - port:
        number: 15004
      tls:
        mode: ISTIO_MUTUAL
    {{- end}}
    connectionPool:
      http:
        http2MaxRequests: 10000
        maxRequestsPerConnection: 10000
`

const policyDeploymentYamlTemplate = `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-{{ $key }}
  namespace: {{ $.Release.Namespace }}
  labels:
    app: istio-mixer
    chart: {{ template "mixer.chart" $ }}
    heritage: {{ $.Release.Service }}
    release: {{ $.Release.Name }}
    version: {{ $.Chart.Version }}
    istio: mixer
spec:
  replicas: {{ $spec.replicaCount }}
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: {{ $key }}
      chart: {{ template "mixer.chart" $ }}
      heritage: {{ $.Release.Service }}
      release: {{ $.Release.Name }}
      version: {{ $.Chart.Version }}
      istio: mixer
      istio-mixer-type: {{ $key }}
  template:
    metadata:
      labels:
        app: {{ $key }}
        chart: {{ template "mixer.chart" $ }}
        heritage: {{ $.Release.Service }}
        release: {{ $.Release.Name }}
        version: {{ $.Chart.Version }}
        istio: mixer
        istio-mixer-type: {{ $key }}
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
{{- with $.Values.podAnnotations }}
{{ toYaml . | indent 8 }}
{{- end }}
    spec:
      serviceAccountName: istio-mixer-service-account
{{- if $.Values.global.priorityClassName }}
      priorityClassName: "{{ $.Values.global.priorityClassName }}"
{{- end }}
      volumes:
      - name: istio-certs
        secret:
          secretName: istio.istio-mixer-service-account
          optional: true
      - name: uds-socket
        emptyDir: {}
      affinity:
      {{- include "nodeaffinity" . | indent 6 }}
      containers:
      - name: mixer
{{- if contains "/" .Values.image }}
        image: "{{ .Values.image }}"
{{- else }}
        image: "{{ $.Values.global.hub }}/{{ $.Values.image }}:{{ $.Values.global.tag }}"
{{- end }}
        imagePullPolicy: {{ $.Values.global.imagePullPolicy }}
        ports:
        - containerPort: {{ .Values.global.monitoringPort }}
        - containerPort: 42422
        args:
          - --monitoringPort={{ .Values.global.monitoringPort }}
          - --address
          - unix:///sock/mixer.socket
{{- if $.Values.global.useMCP }}
    {{- if $.Values.global.controlPlaneSecurityEnabled}}
          - --configStoreURL=mcps://istio-galley.{{ $.Release.Namespace }}.svc:9901
          - --certFile=/etc/certs/cert-chain.pem
          - --keyFile=/etc/certs/key.pem
          - --caCertFile=/etc/certs/root-cert.pem
    {{- else }}
          - --configStoreURL=mcp://istio-galley.{{ $.Release.Namespace }}.svc:9901
    {{- end }}
{{- else }}
          - --configStoreURL=k8s://
{{- end }}
          - --configDefaultNamespace={{ $.Release.Namespace }}
          {{- if $.Values.global.tracer.zipkin.address }}
          - --trace_zipkin_url=http://{{- $.Values.global.tracer.zipkin.address }}/api/v1/spans
          {{- else }}
          - --trace_zipkin_url=http://zipkin:9411/api/v1/spans
          {{- end }}
        {{- if .Values.env }}
        env:
        {{- range $key, $val := .Values.env }}
        - name: {{ $key }}
          value: "{{ $val }}"
        {{- end }}
        {{- end }}
        resources:
{{- if .Values.resources }}
{{ toYaml .Values.resources | indent 10 }}
{{- else }}
{{ toYaml .Values.global.defaultResources | indent 10 }}
{{- end }}
        volumeMounts:
{{- if $.Values.global.useMCP }}
        - name: istio-certs
          mountPath: /etc/certs
          readOnly: true
{{- end }}
        - name: uds-socket
          mountPath: /sock
        livenessProbe:
          httpGet:
            path: /version
            port: {{ .Values.global.monitoringPort }}
          initialDelaySeconds: 5
          periodSeconds: 5
      - name: istio-proxy
{{- if contains "/" $.Values.global.proxy.image }}
        image: "{{ $.Values.global.proxy.image }}"
{{- else }}
        image: "{{ $.Values.global.hub }}/{{ $.Values.global.proxy.image }}:{{ $.Values.global.tag }}"
{{- end }}
        imagePullPolicy: {{ $.Values.global.imagePullPolicy }}
        ports:
        - containerPort: 9091
        - containerPort: 15004
        - containerPort: 15090
          protocol: TCP
          name: http-envoy-prom
        args:
        - proxy
        - --serviceCluster
        - istio-policy
        - --templateFile
        - /etc/istio/proxy/envoy_policy.yaml.tmpl
      {{- if $.Values.global.controlPlaneSecurityEnabled }}
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
{{- if $.Values.global.proxy.resources }}
{{ toYaml $.Values.global.proxy.resources | indent 10 }}
{{- else }}
{{ toYaml .Values.global.defaultResources | indent 10 }}
{{- end }}
        volumeMounts:
        - name: istio-certs
          mountPath: /etc/certs
          readOnly: true
        - name: uds-socket
          mountPath: /sock
`

const telemetryServiceYamlTemplate = `
apiVersion: v1
kind: Service
metadata:
  name: istio-telemetry
  namespace: {{ $.Release.Namespace }}
  labels:
    app: mixer
    istio: mixer
spec:
  ports:
  - name: grpc-mixer
    port: 9091
  - name: grpc-mixer-mtls
    port: 15004
  - name: http-monitoring
    port: {{ $.Values.global.monitoringPort }}
  - name: prometheus
    port: 42422
  selector:
    istio: mixer
    istio-mixer-type: telemetry
`

const telemetryDestinationRuleYamlTemplate = `
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: istio-telemetry
  namespace: {{ .Release.Namespace }}
  labels:
    app: mixer
spec:
  host: istio-telemetry.{{ .Release.Namespace }}.svc.cluster.local
  trafficPolicy:
    {{- if .Values.global.controlPlaneSecurityEnabled }}
    portLevelSettings:
    - port:
        number: 15004
      tls:
        mode: ISTIO_MUTUAL
    {{- end}}
    connectionPool:
      http:
        http2MaxRequests: 10000
        maxRequestsPerConnection: 10000
`

const telemetryDeploymentYamlTemplate = `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-{{ $key }}
  namespace: {{ $.Release.Namespace }}
  labels:
    app: istio-mixer
    chart: {{ template "mixer.chart" $ }}
    heritage: {{ $.Release.Service }}
    release: {{ $.Release.Name }}
    version: {{ $.Chart.Version }}
    istio: mixer
spec:
  replicas: {{ $spec.replicaCount }}
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: {{ $key }}
      chart: {{ template "mixer.chart" $ }}
      heritage: {{ $.Release.Service }}
      release: {{ $.Release.Name }}
      version: {{ $.Chart.Version }}
      istio: mixer
      istio-mixer-type: {{ $key }}
  template:
    metadata:
      labels:
        app: {{ $key }}
        chart: {{ template "mixer.chart" $ }}
        heritage: {{ $.Release.Service }}
        release: {{ $.Release.Name }}
        version: {{ $.Chart.Version }}
        istio: mixer
        istio-mixer-type: {{ $key }}
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
{{- with $.Values.podAnnotations }}
{{ toYaml . | indent 8 }}
{{- end }}
    spec:
      serviceAccountName: istio-mixer-service-account
      volumes:
      - name: istio-certs
        secret:
          secretName: istio.istio-mixer-service-account
          optional: true
      - name: uds-socket
        emptyDir: {}
    {{- if $.Values.nodeSelector }}
      nodeSelector:
{{ toYaml $.Values.nodeSelector | indent 8 }}
    {{- end }}
      containers:
      - name: mixer
{{- if contains "/" .Values.image }}
        image: "{{ .Values.image }}"
{{- else }}
        image: "{{ $.Values.global.hub }}/{{ $.Values.image }}:{{ $.Values.global.tag }}"
{{- end }}
        imagePullPolicy: {{ $.Values.global.imagePullPolicy }}
        ports:
        - containerPort: {{ .Values.global.monitoringPort }}
        - containerPort: 42422
        args:
          - --monitoringPort={{ .Values.global.monitoringPort }}
          - --address
          - unix:///sock/mixer.socket
{{- if $.Values.global.useMCP }}
    {{- if $.Values.global.controlPlaneSecurityEnabled}}
          - --configStoreURL=mcps://istio-galley.{{ $.Release.Namespace }}.svc:9901
          - --certFile=/etc/certs/cert-chain.pem
          - --keyFile=/etc/certs/key.pem
          - --caCertFile=/etc/certs/root-cert.pem
    {{- else }}
          - --configStoreURL=mcp://istio-galley.{{ $.Release.Namespace }}.svc:9901
    {{- end }}
{{- else }}
          - --configStoreURL=k8s://
{{- end }}
          - --configDefaultNamespace={{ $.Release.Namespace }}
          {{- if $.Values.global.tracer.zipkin.address }}
          - --trace_zipkin_url=http://{{- $.Values.global.tracer.zipkin.address }}/api/v1/spans
          {{- else }}
          - --trace_zipkin_url=http://zipkin:9411/api/v1/spans
          {{- end }}
        {{- if .Values.env }}
        env:
        {{- range $key, $val := .Values.env }}
        - name: {{ $key }}
          value: "{{ $val }}"
        {{- end }}
        {{- end }}
        resources:
{{- if .Values.resources }}
{{ toYaml .Values.resources | indent 10 }}
{{- else }}
{{ toYaml .Values.global.defaultResources | indent 10 }}
{{- end }}
        volumeMounts:
{{- if $.Values.global.useMCP }}
        - name: istio-certs
          mountPath: /etc/certs
          readOnly: true
{{- end }}
        - name: uds-socket
          mountPath: /sock
        livenessProbe:
          httpGet:
            path: /version
            port: {{ .Values.global.monitoringPort }}
          initialDelaySeconds: 5
          periodSeconds: 5
      - name: istio-proxy
{{- if contains "/" $.Values.global.proxy.image }}
        image: "{{ $.Values.global.proxy.image }}"
{{- else }}
        image: "{{ $.Values.global.hub }}/{{ $.Values.global.proxy.image }}:{{ $.Values.global.tag }}"
{{- end }}
        imagePullPolicy: {{ $.Values.global.imagePullPolicy }}
        ports:
        - containerPort: 9091
        - containerPort: 15004
        - containerPort: 15090
          protocol: TCP
          name: http-envoy-prom
        args:
        - proxy
{{- if $.Values.global.proxy.proxyDomain }}
        - --domain
        - {{ $.Values.global.proxy.proxyDomain }}
{{- end }}
        - --serviceCluster
        - istio-telemetry
        - --templateFile
        - /etc/istio/proxy/envoy_telemetry.yaml.tmpl
      {{- if $.Values.global.controlPlaneSecurityEnabled }}
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
{{- if $.Values.global.proxy.resources }}
{{ toYaml $.Values.global.proxy.resources | indent 10 }}
{{- else }}
{{ toYaml .Values.global.defaultResources | indent 10 }}
{{- end }}
        volumeMounts:
        - name: istio-certs
          mountPath: /etc/certs
          readOnly: true
        - name: uds-socket
          mountPath: /sock
`

const clusterRoleYamlTemplate = `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: istio-mixer-{{ .Release.Namespace }}
  labels:
    app: mixer
rules:
- apiGroups: ["config.istio.io"] # istio CRD watcher
  resources: ["*"]
  verbs: ["create", "get", "list", "watch", "patch"]
- apiGroups: ["apiextensions.k8s.io"]
  resources: ["customresourcedefinitions"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["configmaps", "endpoints", "pods", "services", "namespaces", "secrets", "replicationcontrollers"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["extensions"]
  resources: ["replicasets"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["replicasets"]
  verbs: ["get", "list", "watch"]
`
