package gateway

import (
	"sync"
	"text/template"

	"github.com/maistra/istio-operator/pkg/components/common"
)

type templateParams struct {
	common.TemplateParams
	EnableCoreDump              bool
	InitImage                   string // Proxy Init Image
	ProxyImage                  string
	ProxyDomain                 string
	ControlPlaneSecurityEnabled bool
}

type gatewayTemplates struct {
	Ingress common.Templates
	Egress  common.Templates
}

var (
	_singleton *gatewayTemplates
	_init      sync.Once
)

func TemplatesInstance() *gatewayTemplates {
	_init.Do(func() {
		commonTemplates := common.TemplatesInstance()
		_singleton = &gatewayTemplates{
      Ingress: common.Templates{
        ServiceAccountTemplate:     commonTemplates.ServiceAccountTemplate,
        ClusterRoleBindingTemplate: commonTemplates.ClusterRoleBindingTemplate,
        ServiceTemplate:            template.New("IngressService.yaml"),
        DeploymentTemplate:         template.New("IngressDeployment.yaml"),
        ClusterRoleTemplate:        template.New("IngressClusterRole.yaml"),
      },
      Egress: common.Templates{
        ServiceAccountTemplate:     commonTemplates.ServiceAccountTemplate,
        ClusterRoleBindingTemplate: commonTemplates.ClusterRoleBindingTemplate,
        ServiceTemplate:            template.New("EgressService.yaml"),
        DeploymentTemplate:         template.New("EgressDeployment.yaml"),
        ClusterRoleTemplate:        template.New("EgressClusterRole.yaml"),
      },
		}
		_singleton.Ingress.ServiceTemplate.Parse(ingressServiceYamlTemplate)
		_singleton.Ingress.DeploymentTemplate.Parse(ingressDeploymentYamlTemplate)
		_singleton.Ingress.ClusterRoleTemplate.Parse(ingressClusterRoleYamlTemplate)
		_singleton.Egress.ServiceTemplate.Parse(egressServiceYamlTemplate)
		_singleton.Egress.DeploymentTemplate.Parse(egressDeploymentYamlTemplate)
		_singleton.Egress.ClusterRoleTemplate.Parse(egressClusterRoleYamlTemplate)
	})
	return _singleton
}

// XXX: ignoring mesh expansion for now

const ingressServiceYamlTemplate = `
apiVersion: v1
kind: Service
metadata:
  name: istio-ingressgateway
  namespace: {{ .Config.Namespace }}
  labels:
    app: istio-ingressgateway
    istio: istio-ingressgateway
spec:
  type: ClusterIP
  selector:
    app: istio-ingressgateway
    istio: istio-ingressgateway
  ports:
  - port: 80
    targetPort: 80
    name: http2
    nodePort: 31380
  - port: 443
    name: https
    nodePort: 31390
  # Example of a port to add. Remove if not needed
  - port: 31400
    name: tcp
    nodePort: 31400
  ### PORTS FOR UI/metrics #####
  ## Disable if not needed
  - port: 15029
    targetPort: 15029
    name: http-kiali
  - port: 15030
    targetPort: 15030
    name: http2-prometheus
  - port: 15031
    targetPort: 15031
    name: http2-grafana
  - port: 15032
    targetPort: 15032
    name: http2-tracing
    # This is the port where sni routing happens
  - port: 15443
    targetPort: 15443
    name: tls
`

const ingressDeploymentYamlTemplate = `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-ingressgateway
  namespace: {{ .Config.Namespace }}
  labels:
    app: istio-ingressgateway
    istio: istio-ingressgateway
spec:
  replicas: {{ .Config.Spec.Gateways.IngressGateway.ReplicaCount }}
  template:
    metadata:
      labels:
        app: istio-ingressgateway
        istio: istio-ingressgateway
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      serviceAccountName: istio-ingressgateway-service-account
{{- if .Config.Spec.General.PriorityClassName }}
      priorityClassName: "{{ .Config.Spec.General.PriorityClassName }}"
{{- end }}
{{- if .Config.Spec.General.Debug.EnableCoreDump }}
      initContainers:
        - name: enable-core-dump
          image: "{{ .Config.Spec.Proxy.InitImage }}"
          imagePullPolicy: IfNotPresent
          command:
            - /bin/sh
          args:
            - -c
            - sysctl -w kernel.core_pattern=/var/lib/istio/core.proxy && ulimit -c unlimited
          securityContext:
            privileged: true
{{- end }}
      containers:
        - name: istio-proxy
          image: "{{ .Config.Spec.Proxy.Image }}"
          imagePullPolicy: {{ .Config.Spec.General.PullPolicy }}
          ports:
          - containerPort: 80
            name: http2
          - containerPort: 443
            name: https
          # Example of a port to add. Remove if not needed
          - containerPort: 31400
            name: tcp
          ### PORTS FOR UI/metrics #####
          ## Disable if not needed
          - containerPort: 15029
            name: http-kiali
          - containerPort: 15030
            name: http2-prometheus
          - containerPort: 15031
            name: http2-grafana
          - containerPort: 15032
            name: http2-tracing
            # This is the port where sni routing happens
          - containerPort: 15443
            name: tls
          - containerPort: 15090
            protocol: TCP
            name: http-envoy-prom
          args:
          - proxy
          - router
{{- if .ProxyDomain }}
          - --domain
          - {{ .Config.Spec.Proxy.ProxyDomain }}
{{- end }}
          - --log_output_level
          - 'info'
          - --drainDuration
          - '45s' #drainDuration
          - --parentShutdownDuration
          - '1m0s' #parentShutdownDuration
          - --connectTimeout
          - '10s' #connectTimeout
          - --serviceCluster
          - istio-ingressgateway
        {{- if eq .Config.Spec.Monitoring.Tracer.Type "lightstep" }}
          - --lightstepAddress
          - {{ .Config.Spec.Monitoring.Tracer.LightStep.Address }}
          - --lightstepAccessToken
          - {{ .Config.Spec.Monitoring.Tracer.LightStep.AccessToken }}
          - --lightstepSecure={{ .Config.Spec.Monitoring.Tracer.LightStep.Secure }}
          - --lightstepCacertPath
          - {{ .Config.Spec.Monitoring.Tracer.LightStep.CACertPath }}
        {{- else if eq .Config.Spec.Monitoring.Tracer.Type "zipkin" }}
          - --zipkinAddress
          - {{ .Config.Spec.Monitoring.Tracer.Zipkin.Address }}
        {{- end }}
        {{- if $.Values.global.proxy.envoyStatsd.enabled }}
          - --statsdUdpAddress
          - {{ $.Values.global.proxy.envoyStatsd.host }}:{{ $.Values.global.proxy.envoyStatsd.port }}
        {{- end }}
          - --proxyAdminPort
          - "15000"
        {{- if $.Config.Spec.Security.ControlPlaneSecurityEnabled }}
          - --controlPlaneAuthPolicy
          - MUTUAL_TLS
          - --discoveryAddress
          - istio-pilot:15011
        {{- else }}
          - --controlPlaneAuthPolicy
          - NONE
          - --discoveryAddress
          - istio-pilot:15010
        {{- end }}
          resources:
          env:
          - name: NODE_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: spec.nodeName
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
          - name: ISTIO_META_POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: ISTIO_META_CONFIG_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          - name: ISTIO_META_ROUTER_MODE
            value: sni-dnat
          volumeMounts:
          {{- if $.Values.global.sds.enabled }}
          - name: sdsudspath
            mountPath: /var/run/sds
          {{- end }}
          - name: istio-certs
            mountPath: /etc/certs
            readOnly: true
          - name: ingressgateway-certs
            mountPath: /etc/istio/ingressgateway-certs
            readOnly: true
          - name: ingressgateway-ca-certs
            mountPath: /etc/istio/ingressgateway-ca-certs
            readOnly: true
      volumes:
      {{- if $.Values.global.sds.enabled }}
      - name: sdsudspath
        hostPath:
          path: /var/run/sds
      {{- end }}
      - name: istio-certs
        secret:
          secretName: istio-ingressgateway-service-account
          optional: true
      - name: ingressgateway-certs
        secret:
          secretName: istio-ingressgateway-certs
          optional: true
      - name: ingressgateway-ca-certs
        secret:
          secretName: istio-ingressgateway-ca-certs
          optional: true
      affinity:
`

const ingressClusterRoleYamlTemplate = `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: {{ .ClusterRoleName }}
  labels:
    app: istio-ingressgateway
    istio: istio-ingressgateway
rules:
- apiGroups: ["networking.istio.io"]
  resources: ["virtualservices", "destinationrules", "gateways"]
  verbs: ["get", "watch", "list", "update"]
`

const egressServiceYamlTemplate = `
apiVersion: v1
kind: Service
metadata:
  name: istio-egressgateway
  namespace: {{ .Config.Namespace }}
  labels:
    app: istio-egressgateway
    istio: istio-egressgateway
spec:
  type: ClusterIP
  selector:
    app: istio-egressgateway
    istio: istio-egressgateway
  ports:
  - port: 80
    name: http2
  - port: 443
    name: https
    # This is the port where sni routing happens
  - port: 15443
    targetPort: 15443
    name: tls
`

const egressDeploymentYamlTemplate = `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-egressgateway
  namespace: {{ .Config.Namespace }}
  labels:
    app: istio-egressgateway
    istio: istio-egressgateway
spec:
  replicas: {{ .Config.Spec.Gateways.EgressGateway.ReplicaCount }}
  template:
    metadata:
      labels:
        app: istio-egressgateway
        istio: istio-egressgateway
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      serviceAccountName: istio-egressgateway-service-account
{{- if .Config.Spec.General.PriorityClassName }}
      priorityClassName: "{{ .Config.Spec.General.PriorityClassName }}"
{{- end }}
{{- if .Config.Spec.General.Debug.EnableCoreDump }}
      initContainers:
        - name: enable-core-dump
          image: "{{ .InitImage }}"
          imagePullPolicy: IfNotPresent
          command:
            - /bin/sh
          args:
            - -c
            - sysctl -w kernel.core_pattern=/var/lib/istio/core.proxy && ulimit -c unlimited
          securityContext:
            privileged: true
{{- end }}
      containers:
        - name: istio-proxy
          image: "{{ .ProxyImage }}"
          imagePullPolicy: {{ .ImagePullPolicy }}
          ports:
          - containerPort: 80
            name: http2
          - containerPort: 443
            name: https
            # This is the port where sni routing happens
          - containerPort: 15443
            name: tls
          args:
          - proxy
          - router
{{- if .Config.Spec.Proxy.ProxyDomain }}
          - --domain
          - {{ .Config.Spec.Proxy.ProxyDomain }}
{{- end }}
          - --log_output_level
          - 'info'
          - --drainDuration
          - '45s' #drainDuration
          - --parentShutdownDuration
          - '1m0s' #parentShutdownDuration
          - --connectTimeout
          - '10s' #connectTimeout
          - --serviceCluster
          - istio-egressgateway
        {{- if eq .Config.Spec.Monitoring.Tracer.Type "lightstep" }}
          - --lightstepAddress
          - {{ .Config.Spec.Monitoring.Tracer.LightStep.Address }}
          - --lightstepAccessToken
          - {{ .Config.Spec.Monitoring.Tracer.LightStep.AccessToken }}
          - --lightstepSecure={{ .Config.Spec.Monitoring.Tracer.LightStep.Secure }}
          - --lightstepCacertPath
          - {{ .Config.Spec.Monitoring.Tracer.LightStep.CACertPath }}
        {{- else if eq .Config.Spec.Monitoring.Tracer.Type "zipkin" }}
          - --zipkinAddress
          - {{ .Config.Spec.Monitoring.Tracer.Zipkin.Address }}
        {{- end }}
        {{- if $.Values.global.proxy.envoyStatsd.enabled }}
          - --statsdUdpAddress
          - {{ $.Values.global.proxy.envoyStatsd.host }}:{{ $.Values.global.proxy.envoyStatsd.port }}
        {{- end }}
          - --proxyAdminPort
          - "15000"
        {{- if $.Config.Spec.Security.ControlPlaneSecurityEnabled }}
          - --controlPlaneAuthPolicy
          - MUTUAL_TLS
          - --discoveryAddress
          - istio-pilot:15011
        {{- else }}
          - --controlPlaneAuthPolicy
          - NONE
          - --discoveryAddress
          - istio-pilot:15010
        {{- end }}
          resources:
          env:
          - name: NODE_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: spec.nodeName
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
          - name: ISTIO_META_POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: ISTIO_META_CONFIG_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          - name: ISTIO_META_ROUTER_MODE
            value: sni-dnat
          volumeMounts:
          {{- if $.Values.global.sds.enabled }}
          - name: sdsudspath
            mountPath: /var/run/sds
          {{- end }}
          - name: istio-certs
            mountPath: /etc/certs
            readOnly: true
          - name: ingressgateway-certs
            mountPath: /etc/istio/ingressgateway-certs
            readOnly: true
          - name: ingressgateway-ca-certs
            mountPath: /etc/istio/ingressgateway-ca-certs
            readOnly: true
      volumes:
      {{- if $.Values.global.sds.enabled }}
      - name: sdsudspath
        hostPath:
          path: /var/run/sds
      {{- end }}
      - name: istio-certs
        secret:
          secretName: istio-egressgateway-service-account
          optional: true
      - name: ingressgateway-certs
        secret:
          secretName: istio-egressgateway-certs
          optional: true
      - name: ingressgateway-ca-certs
        secret:
          secretName: istio-egressgateway-ca-certs
          optional: true
      affinity:
`

const egressClusterRoleYamlTemplate = `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: {{ .ClusterRoleName }}
  labels:
    app: istio-egressgateway
    istio: istio-egressgateway
rules:
- apiGroups: ["networking.istio.io"]
  resources: ["virtualservices", "destinationrules", "gateways"]
  verbs: ["get", "watch", "list", "update"]
`
