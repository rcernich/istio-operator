package pilot

import (
	"sync"
	"text/template"

	"github.com/maistra/istio-operator/pkg/components/common"
)

type templateParams struct {
	common.TemplateParams
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
  name: istio-{{ $name }}
  namespace: {{ $spec.namespace | default $.Release.Namespace }}
  annotations:
    {{- range $key, $val := $spec.serviceAnnotations }}
    {{ $key }}: {{ $val | quote }}
    {{- end }}
  labels:
    app: istio-{{ $name }}
    istio: {{ $name }}
spec:
{{- if $spec.loadBalancerIP }}
  loadBalancerIP: "{{ $spec.loadBalancerIP }}"
{{- end }}
{{- if $spec.loadBalancerSourceRanges }}
  loadBalancerSourceRanges:
{{ toYaml $spec.loadBalancerSourceRanges | indent 4 }}
{{- end }}
{{- if $spec.externalTrafficPolicy }}
  externalTrafficPolicy: {{$spec.externalTrafficPolicy }}
{{- end }}
  type: {{ .type }}
  selector:
    app: istio-{{ $name }}
    istio: {{ $name }}
  ports:
    {{- range $key, $val := $spec.ports }}
    -
      {{- range $pkey, $pval := $val }}
      {{ $pkey}}: {{ $pval }}
      {{- end }}
    {{- end }}
    {{- if $.Values.global.meshExpansion.enabled }}
    {{- range $key, $val := $spec.meshExpansionPorts }}
    -
      {{- range $pkey, $pval := $val }}
      {{ $pkey}}: {{ $pval }}
      {{- end }}
    {{- end }}
    {{- end }}
`

const deploymentYamlTemplate = `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-{{ $name }}
  namespace: {{ $spec.namespace | default $.Release.Namespace }}
  labels:
    app: istio-{{ $name }}
    istio: istio-{{ $name }}
spec:
  replicas: {{ $spec.replicaCount }}
  template:
    metadata:
      labels:
        app: istio-{{ $name }}
        istio: istio-{{ $name }}
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
{{- if $spec.podAnnotations }}
{{ toYaml $spec.podAnnotations | indent 8 }}
{{ end }}
    spec:
      serviceAccountName: istio-{{ $name }}-service-account
{{- if $.Values.global.priorityClassName }}
      priorityClassName: "{{ $.Values.global.priorityClassName }}"
{{- end }}
{{- if $.Values.global.proxy.enableCoreDump }}
      initContainers:
        - name: enable-core-dump
{{- if contains "/" $.Values.global.proxy_init.image }}
          image: "{{ $.Values.global.proxy_init.image }}"
{{- else }}
          image: "{{ $.Values.global.hub }}/{{ $.Values.global.proxy_init.image }}:{{ $.Values.global.tag }}"
{{- end }}
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
{{- if contains "/" $.Values.global.proxy.image }}
          image: "{{ $.Values.global.proxy.image }}"
{{- else }}
          image: "{{ $.Values.global.hub }}/{{ $.Values.global.proxy.image }}:{{ $.Values.global.tag }}"
{{- end }}
          imagePullPolicy: {{ $.Values.global.imagePullPolicy }}
          ports:
            {{- range $key, $val := $spec.ports }}
            - containerPort: {{ $val.port }}
            {{- end }}
            - containerPort: 15090
              protocol: TCP
              name: http-envoy-prom
          args:
          - proxy
          - router
{{- if $.Values.global.proxy.proxyDomain }}
          - --domain
          - {{ $.Values.global.proxy.proxyDomain }}
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
          - {{ $key }}
          - --zipkinAddress
        {{- if $.Values.global.tracer.zipkin.address }}
          - {{ $.Values.global.tracer.zipkin.address }}
        {{- else if $.Values.global.istioNamespace }}
          - zipkin.{{ $.Values.global.istioNamespace }}:9411
        {{- else }}
          - zipkin:9411
        {{- end }}
        {{- if $.Values.global.proxy.envoyStatsd.enabled }}
          - --statsdUdpAddress
          - {{ $.Values.global.proxy.envoyStatsd.host }}:{{ $.Values.global.proxy.envoyStatsd.port }}
        {{- end }}
          - --proxyAdminPort
          - "15000"
        {{- if $.Values.global.controlPlaneSecurityEnabled }}
          - --controlPlaneAuthPolicy
          - MUTUAL_TLS
          - --discoveryAddress
          {{- if $.Values.global.istioNamespace }}
          - istio-pilot.{{ $.Values.global.istioNamespace }}:15011
          {{- else }}
          - istio-pilot:15011
          {{- end }}
        {{- else }}
          - --controlPlaneAuthPolicy
          - NONE
          - --discoveryAddress
          {{- if $.Values.global.istioNamespace }}
          - istio-pilot.{{ $.Values.global.istioNamespace }}:15010
          {{- else }}
          - istio-pilot:15010
          {{- end }}
        {{- end }}
          resources:
{{- if $spec.resources }}
{{ toYaml $spec.resources | indent 12 }}
{{- else }}
{{ toYaml $.Values.global.defaultResources | indent 12 }}
{{- end }}
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
          {{- if $spec.env }}
          {{- range $key, $val := $spec.env }}
          - name: {{ $key }}
            value: {{ $val }}
          {{- end }}
          {{- end }}
          volumeMounts:
          {{- if $.Values.global.sds.enabled }}
          - name: sdsudspath
            mountPath: /var/run/sds
          {{- end }}
          - name: istio-certs
            mountPath: /etc/certs
            readOnly: true
          {{- range $spec.secretVolumes }}
          - name: {{ .name }}
            mountPath: {{ .mountPath | quote }}
            readOnly: true
          {{- end }}
{{- if $spec.additionalContainers }}
{{ toYaml $spec.additionalContainers | indent 8 }}
{{- end }}
      volumes:
      {{- if $.Values.global.sds.enabled }}
      - name: sdsudspath
        hostPath:
          path: /var/run/sds
      {{- end }}
      - name: istio-certs
        secret:
          secretName: istio.{{ $key }}-service-account
          optional: true
      {{- range $spec.secretVolumes }}
      - name: {{ .name }}
        secret:
          secretName: {{ .secretName | quote }}
          optional: true
      {{- end }}
      {{- range $spec.configVolumes }}
      - name: {{ .name }}
        configMap:
          name: {{ .configMapName | quote }}
          optional: true
      {{- end }}
      affinity:
      {{- include "nodeaffinity" $ | indent 6 }}
`

const clusterRoleYamlTemplate = `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: istio-{{ $name }}-{{ $.Release.Namespace }}
  labels:
    app: istio-{{ $name }}
    istio: istio-{{ $name }}
rules:
- apiGroups: ["networking.istio.io"]
  resources: ["virtualservices", "destinationrules", "gateways"]
  verbs: ["get", "watch", "list", "update"]
`
