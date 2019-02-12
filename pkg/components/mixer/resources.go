package mixer

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
	PodAnnotations              string          // TODO
	UseMCP                      bool            // TODO?
	ZipkinAddress               string          // TODO, default to zipkin:9411
	Resources                   string          // TODO
	NodeSelector                string          // TODO
  Env                         []corev1.EnvVar // TODO
  ProxyDomain                 string
  ProxyImage                  string
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
  namespace: {{ .Namespace }}
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
    port: {{ .MonitoringPort }}
  selector:
    istio: mixer
    istio-mixer-type: policy
`

const policyDestinationRuleYamlTemplate = `
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: istio-policy
  namespace: {{ .Namespace }}
  labels:
    app: mixer
spec:
  host: istio-policy.{{ .Namespace }}.svc.cluster.local
  trafficPolicy:
    {{- if .ControlPlaneSecurityEnabled }}
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
  name: istio-mixer-policy
  namespace: {{ .Namespace }}
  labels:
    app: istio-mixer
    istio: mixer
spec:
  replicas: {{ .ReplicaCount }}
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: istio-mixer-policy
      istio: mixer
      istio-mixer-type: policy
  template:
    metadata:
      labels:
        app: istio-mixer-policy
        istio: mixer
        istio-mixer-type: policy
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
{{- with .PodAnnotations }}
{{ toYaml . | indent 8 }}
{{- end }}
    spec:
      serviceAccountName: istio-mixer-service-account
{{- if .PriorityClassName }}
      priorityClassName: "{{ .PriorityClassName }}"
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
        image: "{{ .Image }}"
        imagePullPolicy: {{ .ImagePullPolicy }}
        ports:
        - containerPort: {{ .MonitoringPort }}
        - containerPort: 42422
        args:
          - --monitoringPort={{ .MonitoringPort }}
          - --address
          - unix:///sock/mixer.socket
{{- if .UseMCP }}
    {{- if .ControlPlaneSecurityEnabled }}
          - --configStoreURL=mcps://istio-galley.{{ .Namespace }}.svc:9901
          - --certFile=/etc/certs/cert-chain.pem
          - --keyFile=/etc/certs/key.pem
          - --caCertFile=/etc/certs/root-cert.pem
    {{- else }}
          - --configStoreURL=mcp://istio-galley.{{ .Namespace }}.svc:9901
    {{- end }}
{{- else }}
          - --configStoreURL=k8s://
{{- end }}
          - --configDefaultNamespace={{ .Namespace }}
          {{- if .ZipkinAddress }}
          - --trace_zipkin_url=http://{{- .ZipkinAddress }}/api/v1/spans
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
{{- if .UseMCP }}
        - name: istio-certs
          mountPath: /etc/certs
          readOnly: true
{{- end }}
        - name: uds-socket
          mountPath: /sock
        livenessProbe:
          httpGet:
            path: /version
            port: {{ .MonitoringPort }}
          initialDelaySeconds: 5
          periodSeconds: 5
      - name: istio-proxy
        image: "{{ .ProxyImage }}"
        imagePullPolicy: {{ .ImagePullPolicy }}
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
      {{- if .ControlPlaneSecurityEnabled }}
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
  namespace: {{ .Namespace }}
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
    port: {{ .MonitoringPort }}
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
  namespace: {{ .Namespace }}
  labels:
    app: mixer
spec:
  host: istio-telemetry.{{ .Namespace }}.svc.cluster.local
  trafficPolicy:
    {{- if .ControlPlaneSecurityEnabled }}
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
  name: istio-mixer-telemetry
  namespace: {{ .Namespace }}
  labels:
    app: istio-mixer-telemetry
    istio: mixer
spec:
  replicas: {{ .ReplicaCount }}
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: istio-mixer-telemetry
      istio: mixer
      istio-mixer-type: telemetry
  template:
    metadata:
      labels:
        app: istio-mixer-telemetry
        istio: mixer
        istio-mixer-type: telemetry
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
{{- with .PodAnnotations }}
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
        image: "{{ .Image }}"
        imagePullPolicy: {{ .ImagePullPolicy }}
        ports:
        - containerPort: {{ .Values.global.monitoringPort }}
        - containerPort: 42422
        args:
          - --monitoringPort={{ .Values.global.monitoringPort }}
          - --address
          - unix:///sock/mixer.socket
{{- if .UseMCP }}
    {{- if .ControlPlaneSecurityEnabled}}
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
          - --configDefaultNamespace={{ .Namespace }}
          {{- if .ZipkinAddress }}
          - --trace_zipkin_url=http://{{- .ZipkinAddress }}/api/v1/spans
          {{- else }}
          - --trace_zipkin_url=http://zipkin:9411/api/v1/spans
          {{- end }}
        {{- if .Env }}
        env:
        {{- range $key, $val := .Env }}
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
{{- if .UseMCP }}
        - name: istio-certs
          mountPath: /etc/certs
          readOnly: true
{{- end }}
        - name: uds-socket
          mountPath: /sock
        livenessProbe:
          httpGet:
            path: /version
            port: {{ .MonitoringPort }}
          initialDelaySeconds: 5
          periodSeconds: 5
      - name: istio-proxy
        image: "{{ .Image }}"
        imagePullPolicy: {{ .ImagePullPolicy }}
        ports:
        - containerPort: 9091
        - containerPort: 15004
        - containerPort: 15090
          protocol: TCP
          name: http-envoy-prom
        args:
        - proxy
{{- if .ProxyDomain }}
        - --domain
        - {{ .ProxyDomain }}
{{- end }}
        - --serviceCluster
        - istio-telemetry
        - --templateFile
        - /etc/istio/proxy/envoy_telemetry.yaml.tmpl
      {{- if .ControlPlaneSecurityEnabled }}
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
  name: {{ .ClusterRoleName }}
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
