package mixer

import (
	"sync"
	"text/template"

	"github.com/maistra/istio-operator/pkg/components/common"
)

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
  namespace: {{ .Config.Namespace }}
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
    port: {{ .Config.Spec.Monitoring.Port }}
  selector:
    istio: mixer
    istio-mixer-type: policy
`

const policyDestinationRuleYamlTemplate = `
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: istio-policy
  namespace: {{ .Config.Namespace }}
  labels:
    app: mixer
spec:
  host: istio-policy.{{ .Config.Namespace }}.svc.cluster.local
  trafficPolicy:
    {{- if .Config.Spec.Security.ControlPlaneSecurityEnabled }}
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
  namespace: {{ .Config.Namespace }}
  labels:
    app: istio-mixer
    istio: mixer
spec:
  replicas: {{ .Config.Spec.Mixer.Policy.ReplicaCount }}
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
    spec:
      serviceAccountName: istio-mixer-service-account
{{- if .Config.Spec.General.PriorityClassName }}
      priorityClassName: "{{ .Config.Spec.General.PriorityClassName }}"
{{- end }}
      volumes:
      - name: istio-certs
        secret:
          secretName: istio.istio-mixer-service-account
          optional: true
      - name: uds-socket
        emptyDir: {}
      affinity:
      containers:
      - name: mixer
        image: "{{ .Config.Spec.Mixer.Policy.Image }}"
        imagePullPolicy: {{ .Config.Spec.General.PullPolicy }}
        ports:
        - containerPort: {{ .Config.Spec.Monitoring.Port }}
        - containerPort: 42422
        args:
          - --monitoringPort={{ .Config.Spec.Monitoring.Port }}
          - --address
          - unix:///sock/mixer.socket
    {{- if .Config.Spec.Security.ControlPlaneSecurityEnabled }}
          - --configStoreURL=mcps://istio-galley.{{ .Config.Namespace }}.svc:9901
          - --certFile=/etc/certs/cert-chain.pem
          - --keyFile=/etc/certs/key.pem
          - --caCertFile=/etc/certs/root-cert.pem
    {{- else }}
          - --configStoreURL=mcp://istio-galley.{{ .Config.Namespace }}.svc:9901
    {{- end }}
          - --configDefaultNamespace={{ .Config.Namespace }}
          {{- if eq .Config.Spec.Monitoring.Tracer.Type "zipkin" }}
          - --trace_zipkin_url=http://{{- .Config.Spec.Monitoring.Tracer.Zipkin.Address }}/api/v1/spans
          {{- else }}
          - --trace_zipkin_url=http://zipkin:9411/api/v1/spans
          {{- end }}
        resources:
        volumeMounts:
        - name: istio-certs
          mountPath: /etc/certs
          readOnly: true
        - name: uds-socket
          mountPath: /sock
        livenessProbe:
          httpGet:
            path: /version
            port: {{ .Config.Spec.Monitoring.Port }}
          initialDelaySeconds: 5
          periodSeconds: 5
      - name: istio-proxy
        image: "{{ .Config.Spec.Proxy.Image }}"
        imagePullPolicy: {{ .Config.Spec.Genral.PullPolicy }}
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
      {{- if .Config.Spec.Security.ControlPlaneSecurityEnabled }}
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
        - name: uds-socket
          mountPath: /sock
`

const telemetryServiceYamlTemplate = `
apiVersion: v1
kind: Service
metadata:
  name: istio-telemetry
  namespace: {{ .Config.Namespace }}
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
    port: {{ .Config.Spec.Monitoring.Port }}
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
  namespace: {{ .Config.Namespace }}
  labels:
    app: mixer
spec:
  host: istio-telemetry.{{ .Config.Namespace }}.svc.cluster.local
  trafficPolicy:
    {{- if .Config.Spec.Security.ControlPlaneSecurityEnabled }}
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
  namespace: {{ .Config.Namespace }}
  labels:
    app: istio-mixer-telemetry
    istio: mixer
spec:
  replicas: {{ .Config.Sepc.Mixer.Telemetry.ReplicaCount }}
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
    spec:
      serviceAccountName: istio-mixer-service-account
      volumes:
      - name: istio-certs
        secret:
          secretName: istio.istio-mixer-service-account
          optional: true
      - name: uds-socket
        emptyDir: {}
      containers:
      - name: mixer
        image: "{{ .Config.Spec.Mixer.Telemetry.Image }}"
        imagePullPolicy: {{ .Config.Spec.General.PullPolicy }}
        ports:
        - containerPort: {{ .Config.Spec.Monitoring.Port }}
        - containerPort: 42422
        args:
          - --monitoringPort={{ .Config.Spec.Monitoring.Port }}
          - --address
          - unix:///sock/mixer.socket
    {{- if .Config.Spec.Security.ControlPlaneSecurityEnabled}}
          - --configStoreURL=mcps://istio-galley.{{ $.Config.Namespace }}.svc:9901
          - --certFile=/etc/certs/cert-chain.pem
          - --keyFile=/etc/certs/key.pem
          - --caCertFile=/etc/certs/root-cert.pem
    {{- else }}
          - --configStoreURL=mcp://istio-galley.{{ $.Config.Namespace }}.svc:9901
    {{- end }}
          - --configDefaultNamespace={{ .Config.Namespace }}
          {{- if eq .Config.Spec.Monitoring.Tracer.Type "zipkin" }}
          - --trace_zipkin_url=http://{{- .Config.Spec.Monitoring.Tracer.Zipkin.Address }}/api/v1/spans
          {{- else }}
          - --trace_zipkin_url=http://zipkin:9411/api/v1/spans
          {{- end }}
        resources:
        volumeMounts:
        - name: istio-certs
          mountPath: /etc/certs
          readOnly: true
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
{{- if .Config.Spec.Proxy.ProxyDomain }}
        - --domain
        - {{ .Config.Spec.Proxy.ProxyDomain }}
{{- end }}
        - --serviceCluster
        - istio-telemetry
        - --templateFile
        - /etc/istio/proxy/envoy_telemetry.yaml.tmpl
      {{- if .Config.Spec.Security.ControlPlaneSecurityEnabled }}
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
