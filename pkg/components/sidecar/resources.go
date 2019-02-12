package sidecar

import (
	"sync"
	"text/template"

	"github.com/maistra/istio-operator/pkg/components/common"
)

type templates struct {
	common.Templates
	MutatingWebHookTemplate *template.Template
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
				ServiceTemplate:            template.New("Service.yaml"),
				DeploymentTemplate:         template.New("Deployment.yaml"),
				ClusterRoleTemplate:        template.New("ClusterRole.yaml"),
				ConfigMapTemplate:          template.New("ConfigMap.yaml"),
			},
			MutatingWebHookTemplate: template.New("MutatingWebHook.yaml"),
		}
		_singleton.ServiceTemplate.Parse(serviceYamlTemplate)
		_singleton.DeploymentTemplate.Parse(deploymentYamlTemplate)
		_singleton.ClusterRoleTemplate.Parse(clusterRoleYamlTemplate)
		_singleton.ConfigMapTemplate.Parse(configMapYaml)
		_singleton.MutatingWebHookTemplate.Parse(mutatingWebhookYamlTemplate)
	})
	return _singleton
}

const serviceYamlTemplate = `
apiVersion: v1
kind: Service
metadata:
  name: istio-sidecar-injector
  namespace: {{ .Config.Namespace }}
  labels:
    app: istio
    istio: sidecar-injector
spec:
  ports:
  - port: 443
  selector:
    istio: sidecar-injector
`

const clusterRoleYamlTemplate = `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: {{ .ClusterRoleName }}
  labels:
    app: istio
    istio: sidecar-injector
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["admissionregistration.k8s.io"]
  resources: ["mutatingwebhookconfigurations"]
  verbs: ["get", "list", "watch", "patch"]
`

const mutatingWebhookYamlTemplate = `
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  name: istio-sidecar-injector
  namespace: {{ .Config.Namespace }}
  labels:
    app: sidecarinjectorWebhook
webhooks:
  - name: sidecar-injector.istio.io
    clientConfig:
      service:
        name: istio-sidecar-injector
        namespace: {{ .Config.Namespace }}
        path: "/inject"
      caBundle: ""
    rules:
      - operations: [ "CREATE" ]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["pods"]
    failurePolicy: Fail
    namespaceSelector:
{{- if .Config.Spec.SidecarInjector.EnableNamespacesByDefault }}
      matchExpressions:
      - key: istio-injection
        operator: NotIn
        values:
        - disabled
{{- else }}
      matchLabels:
        istio-injection: enabled
{{- end }}

`

const deploymentYamlTemplate = `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-sidecar-injector
  namespace: {{ .Config.Namespace }}
  labels:
    app: sidecarinjectorWebhook
    istio: sidecar-injector
spec:
  replicas: {{ .Config.Spec.SidecarInjector.ReplicaCount }}
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app: sidecarinjectorWebhook
        istio: sidecar-injector
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      serviceAccountName: istio-sidecar-injector-service-account
 {{- if .Config.Spec.General.PriorityClassName }}
      priorityClassName: "{{ .Config.Spec.General.PriorityClassName }}"
{{- end }}
      containers:
        - name: sidecar-injector-webhook
          image: "{{ .Config.Spec.SidecarInjector.Image }}"
          imagePullPolicy: {{ .Config.Spec.General.PullPolicy }}
          args:
            - --caCertFile=/etc/istio/certs/root-cert.pem
            - --tlsCertFile=/etc/istio/certs/cert-chain.pem
            - --tlsKeyFile=/etc/istio/certs/key.pem
            - --injectConfig=/etc/istio/inject/config
            - --meshConfig=/etc/istio/config/mesh
            - --healthCheckInterval=2s
            - --healthCheckFile=/health
          volumeMounts:
          - name: config-volume
            mountPath: /etc/istio/config
            readOnly: true
          - name: certs
            mountPath: /etc/istio/certs
            readOnly: true
          - name: inject-config
            mountPath: /etc/istio/inject
            readOnly: true
          livenessProbe:
            exec:
              command:
                - /usr/local/bin/sidecar-injector
                - probe
                - --probe-path=/health
                - --interval=4s
            initialDelaySeconds: 4
            periodSeconds: 4
          readinessProbe:
            exec:
              command:
                - /usr/local/bin/sidecar-injector
                - probe
                - --probe-path=/health
                - --interval=4s
            initialDelaySeconds: 4
            periodSeconds: 4
          resources:
      volumes:
      - name: config-volume
        configMap:
          name: istio
      - name: certs
        secret:
          secretName: istio.istio-sidecar-injector-service-account
      - name: inject-config
        configMap:
          name: istio-sidecar-injector
          items:
          - key: config
            path: config
      affinity:
`

// This should always be installed for istio control plane
const configMapYaml = `
[[- $statusPort := annotation .ObjectMeta "status.sidecar.istio.io/port" "{{ .Config.Spec.Status.Port }}" ]]
[[- $interceptionMode :=  annotation .ObjectMeta "sidecar.istio.io/interceptionMode" .ProxyConfig.InterceptionMode ]]
[[- $includeOutboundIPRanges := annotation .ObjectMeta "traffic.sidecar.istio.io/includeOutboundIPRanges" "{{ .Config.Spec.Proxy.EgressWhiteList.IncludeIPRanges }}" ]]
[[- $excludeOutboundIPRanges := annotation .ObjectMeta "traffic.sidecar.istio.io/excludeOutboundIPRanges" "{{ .Config.Spec.Proxy.EgressWhiteList.ExcludeIPRanges }}" ]]
[[- $includeInboundPorts := annotation .ObjectMeta "traffic.sidecar.istio.io/includeInboundPorts" (includeInboundPorts .Spec.Containers) ]]
[[- $excludeInboundPorts := excludeInboundPort ($statusPort) (annotation .ObjectMeta "traffic.sidecar.istio.io/excludeInboundPorts" "{{ .Config.Spec.Proxy.IngressWhiteList.ExcludeInboundPorts }}") ]]
[[- $proxyImage := annotation .ObjectMeta "sidecar.istio.io/proxyImage" "{{ .Config.Spec.Proxy.Image }}" ]]
[[- $discoveryAddress := annotation .ObjectMeta "sidecar.istio.io/discoveryAddress" .ProxyConfig.DiscoveryAddress ]]
[[- $controlPlaneAuthPolicy := annotation .ObjectMeta "sidecar.istio.io/controlPlaneAuthPolicy" .ProxyConfig.ControlPlaneAuthPolicy ]]
[[- $applicationPorts := annotation .ObjectMeta "readiness.status.sidecar.istio.io/applicationPorts" (applicationPorts .Spec.Containers) ]]
[[- $initialDelaySeconds := annotation .ObjectMeta "readiness.status.sidecar.istio.io/initialDelaySeconds" "{{ .Config.Spec.Status.InitialDelaySeconds }}" ]]
[[- $periodSeconds := annotation .ObjectMeta "readiness.status.sidecar.istio.io/periodSeconds" "{{ .Config.Spec.Status.PeriodSeconds }}" ]]
[[- $failureThreshold := annotation .ObjectMeta "readiness.status.sidecar.istio.io/failureThreshold" "{{ .Config.Spec.Status.FailureThreshold }}" ]]
[[- $proxyCPU := index .ObjectMeta.Annotations "sidecar.istio.io/proxyCPU" ]]
[[- $proxyMemory: = index .ObjectMeta.Annotations "sidecar.istio.io/proxyMemory" ]]

apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-sidecar-injector
  namespace: {{ .Config.Namespace }}
  labels:
    app: istio
    istio: sidecar-injector
data:
  config: |-
    policy: {{ .Config.Spec.Proxy.AutoInject }}
    template: |-
      initContainers:
      - name: istio-init
        image: "{{ .Config.Spec.Proxy.InitImage }}"
        args:
        - "-p"
        - "[[ .MeshConfig.ProxyListenPort ]]"
        - "-u"
        - 1337
        - "-m"
        - "[[ $interceptionMode ]]"
        - "-i"
        - "[[ $includeOutboundIPRanges ]]"
        - "-x"
        - "[[ $excludeOutboundIPRanges ]]"
        - "-b"
        - "[[ $includeInboundPorts ]]"
        - "-d"
        - "[[ $excludeInboundPorts ]]"
        imagePullPolicy: {{ .Config.Spec.General.PullPolicy }}
        resources:
          requests:
            cpu: 10m
            memory: 10Mi
          limits:
            cpu: 10m
            memory: 10Mi
        securityContext:
          capabilities:
            add:
            - NET_ADMIN
          {{- if .Config.Spec.Proxy.Privileged }}
          privileged: true
          {{- end }}
        restartPolicy: Always
      {{- if eq .Config.Spec.General.Debug.EnableCoreDump true }}
      - name: enable-core-dump
        args:
        - -c
        - sysctl -w kernel.core_pattern=/var/lib/istio/core.proxy && ulimit -c unlimited
        command:
          - /bin/sh
        image: "{{ .Config.Spec.Proxy.InitImage }}"
        imagePullPolicy: IfNotPresent
        resources: {}
        securityContext:
          privileged: true
      {{ end }}
{{- end }}
      containers:
      - name: istio-proxy
        image: "[[ $proxyImage ]]"
        ports:
        - containerPort: 15090
          protocol: TCP
          name: http-envoy-prom
        args:
        - proxy
        - sidecar
{{- if .Config.Spec.Proxy.ProxyDomain }}
        - --domain
        - {{ .Config.Spec.Proxy.ProxyDomain }}
{{- end }}
        - --configPath
        - {{ "[[ .ProxyConfig.ConfigPath ]]" }}
        - --binaryPath
        - {{ "[[ .ProxyConfig.BinaryPath ]]" }}
        - --serviceCluster
        {{ "[[ if ne \"\" (index .ObjectMeta.Labels \"app\") -]]" }}
        - {{ "[[ index .ObjectMeta.Labels \"app\" ]].[[ valueOrDefault .DeploymentMeta.Namespace \"default\" ]]" }}
        {{ "[[ else -]]" }}
        - {{ "[[ valueOrDefault .DeploymentMeta.Name \"istio-proxy\" ]].[[ valueOrDefault .DeploymentMeta.Namespace \"default\" ]]" }}
        {{ "[[ end -]]" }}
        - --drainDuration
        - {{ "[[ formatDuration .ProxyConfig.DrainDuration ]]" }}
        - --parentShutdownDuration
        - {{ "[[ formatDuration .ProxyConfig.ParentShutdownDuration ]]" }}
        - --discoveryAddress
        - "[[ $discoveryAddress ]]"
      {{- if eq .Config.Spec.Monitoring.Tracer.Type "lightstep" }}
        - --lightstepAddress
        - {{ "[[ .ProxyConfig.GetTracing.GetLightstep.GetAddress ]]" }}
        - --lightstepAccessToken
        - {{ "[[ .ProxyConfig.GetTracing.GetLightstep.GetAccessToken ]]" }}
        - --lightstepSecure={{ "[[ .ProxyConfig.GetTracing.GetLightstep.GetSecure ]]" }}
        - --lightstepCacertPath
        - {{ "[[ .ProxyConfig.GetTracing.GetLightstep.GetCacertPath ]]" }}
      {{- else if eq .Config.Spec.Monitoring.Tracer.Type "zipkin" }}
        - --zipkinAddress
        - {{ "[[ .ProxyConfig.GetTracing.GetZipkin.GetAddress ]]" }}
      {{- end }}
        - --connectTimeout
        - {{ "[[ formatDuration .ProxyConfig.ConnectTimeout ]]" }}
      {{- if .Values.global.proxy.envoyStatsd.enabled }}
        - --statsdUdpAddress
        - {{ "[[ .ProxyConfig.StatsdUdpAddress ]]" }}
      {{- end }}
        - --proxyAdminPort
        - {{ "[[ .ProxyConfig.ProxyAdminPort ]]" }}
        {{ "[[ if gt .ProxyConfig.Concurrency 0 -]]" }}
        - --concurrency
        - {{ "[[ .ProxyConfig.Concurrency ]]" }}
        {{ "[[ end -]]" }}
        - --controlPlaneAuthPolicy
        - "[[ $controlPlaneAuthPolicy ]]"
        [[- if (ne $statusPort "0") ]]
        - --statusPort
        - "[[ $statusPort ]]"
        - --applicationPorts
        - "[[ $applicationPorts ]]"
        [[- end ]]
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: INSTANCE_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        - name: ISTIO_META_POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: ISTIO_META_CONFIG_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: ISTIO_META_INTERCEPTION_MODE
          value: "[[ $interceptionMode ]]"
        {{- if .Values.global.network }}
        - name: ISTIO_META_NETWORK
          value: "{{ .Values.global.network }}"
        {{- end }}
        [[ if .ObjectMeta.Annotations ]]
        - name: ISTIO_METAJSON_ANNOTATIONS
          value: |
                 "[[ toJSON .ObjectMeta.Annotations ]]"
        [[ end ]]
        [[ if .ObjectMeta.Labels ]]
        - name: ISTIO_METAJSON_LABELS
          value: |
                 "[[ toJSON .ObjectMeta.Labels ]]"
        [[ end ]]
        imagePullPolicy: {{ .Config.Spec.General.PullPolicy }}
        [[ if (ne $statusPort "0") ]]
        readinessProbe:
          httpGet:
            path: /healthz/ready
            port: "[[ $statusPort ]]
          initialDelaySeconds: "[[ $initialDelaySeconds ]]"
          periodSeconds: "[[ $periodSeconds ]]"
          failureThreshold: "[[ $failureThreshold ]]"
        [[ end -]]
        securityContext:
          {{- if .Config.Spec.Proxy.Privileged }}
          privileged: true
          {{- end }}
          {{- if ne .Config.Spec.General.Debug.EnableCoreDump true }}
          readOnlyRootFilesystem: true
          {{- end }}
          [[ if eq $interceptionMode "TPROXY" -]]
          capabilities:
            add:
            - NET_ADMIN
          runAsGroup: 1337
          [[ else -]]
          runAsUser: 1337
          [[- end ]]
        resources:
          [[ if $proxyCPU -]]
          requests:
            cpu: "[[ $proxyCPU ]]"
            memory: "[[ $proxyMemory ]]"
          [[ else -]]
        {{ "[[ end -]]" }}
        volumeMounts:
        - mountPath: /etc/istio/proxy
          name: istio-envoy
        - mountPath: /etc/certs/
          name: istio-certs
          readOnly: true
        {{- if .Values.global.sds.enabled }}
        - mountPath: /var/run/sds
          name: sds-uds-path
        {{- if .Values.global.sds.enableTokenMount }}
        - mountPath: /var/run/secrets/tokens
          name: istio-token
        {{- end }}
        {{- end }}
        {{- if and (eq .Config.Spec.Monitoring.Tracer.Type "lightstep") .Config.Spec.Monitoring.Tracer.LightStep.CACertPath }}
        - mountPath: {{ "[[ directory .ProxyConfig.GetTracing.GetLightstep.GetCacertPath ]]" }}
          name: lightstep-certs
          readOnly: true
        {{- end }}
      volumes:
      {{- if .Values.global.sds.enabled }}
      - name: sds-uds-path
        hostPath:
          path: /var/run/sds
      {{- if .Values.global.sds.enableTokenMount }}
      - name: istio-token
        projected:
          sources:
          - serviceAccountToken:
              path: istio-token
              expirationSeconds: 43200
              audience: {{ .Config.Spec.Security.TrustDomain }}
      {{- end }}
      {{- end }}
      {{- if and (eq .Config.Spec.Monitoring.Tracer.Type "lightstep") .Config.Spec.Monitoring.Tracer.LightStep.CACertPath }}
      - name: lightstep-certs
        secret:
          optional: true
          secretName: lightstep.cacert
      {{- end }}
      - emptyDir:
          medium: Memory
        name: istio-envoy
      - name: istio-certs
        secret:
          optional: true
          {{ "[[ if eq .Spec.ServiceAccountName \"\" -]]" }}
          secretName: istio.default
          {{ "[[ else -]]" }}
          secretName: {{ "[[ printf \"istio.%s\" .Spec.ServiceAccountName ]]"  }}
          {{ "[[ end -]]" }}
{{- end }}
{{- if .Values.global.podDNSSearchNamespaces }}
      dnsConfig:
        searches:
          {{- range .Values.global.podDNSSearchNamespaces }}
          - {{ . }}
          {{- end }}
`
