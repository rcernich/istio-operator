package citadel

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
	SelfSigned                  bool
	Resources                   string
}

type templates struct {
	common.Templates
	MtlsDestinationRuleListTemplate *template.Template
	MtlsMeshPolicyTemplate          *template.Template
	PermissiveMeshPolicyTemplate    *template.Template
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
			},
			MtlsDestinationRuleListTemplate: template.New("MtlsDestinationRules.yaml"),
			MtlsMeshPolicyTemplate:          template.New("MtlsMeshPolicy.yaml"),
			PermissiveMeshPolicyTemplate:    template.New("PermissiveMeshPolicy.yaml"),
		}
		_singleton.ServiceTemplate.Parse(serviceYamlTemplate)
		_singleton.DeploymentTemplate.Parse(deploymentYamlTemplate)
		_singleton.ClusterRoleTemplate.Parse(clusterRoleYamlTemplate)
		_singleton.MtlsDestinationRuleListTemplate.Parse(mtlsMeshDestinationRuleListYamlTemplate)
		_singleton.MtlsMeshPolicyTemplate.Parse(mtlsMeshPolicyYamlTemplate)
		_singleton.PermissiveMeshPolicyTemplate.Parse(permissiveMeshPolicyYamlTemplate)
	})
	return _singleton
}

const serviceYamlTemplate = `
apiVersion: v1
kind: Service
metadata:
  name: istio-citadel
  namespace: {{ .Namespace }}
  labels:
    app: security
    istio: citadel
spec:
  ports:
    - name: grpc-citadel
      port: 8060
      targetPort: 8060
      protocol: TCP
    - name: http-monitoring
      port: 9093
  selector:
    istio: citadel
`

const deploymentYamlTemplate = `
# istio CA watching all namespaces
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-citadel
  namespace: {{ .Namespace }}
  labels:
    app: security
    istio: citadel
spec:
  replicas: {{ .ReplicaCount }}
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app: security
        istio: citadel
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
    spec:
      serviceAccountName: istio-citadel-service-account
{{- if .PriorityClassName }}
      priorityClassName: "{{ .PriorityClassName }}"
{{- end }}
      containers:
        - name: citadel
          image: "{{ .Image }}"
          imagePullPolicy: {{ .ImagePullPolicy }}
          args:
            - --append-dns-names=true
            - --grpc-port=8060
            - --grpc-hostname=citadel
            - --citadel-storage-namespace={{ .Namespace }}
            - --custom-dns-names=istio-pilot-service-account.{{ .Namespace }}:istio-pilot.{{ .Namespace }}
          {{- if .SelfSigned }}
            - --self-signed-ca=true
          {{- else }}
            - --self-signed-ca=false
            - --signing-cert=/etc/cacerts/ca-cert.pem
            - --signing-key=/etc/cacerts/ca-key.pem
            - --root-cert=/etc/cacerts/root-cert.pem
            - --cert-chain=/etc/cacerts/cert-chain.pem
          {{- end }}
          {{- if .TrustDomain }}
            - --trust-domain={{ .TrustDomain }}
          {{- end }}
          resources:
{{- if .Resources }}
{{ toYaml .Resources | indent 12 }}
{{- end }}
{{- if not .SelfSigned }}
          volumeMounts:
          - name: cacerts
            mountPath: /etc/cacerts
            readOnly: true
      volumes:
      - name: cacerts
        secret:
         secretName: cacerts
         optional: true
{{- end }}
      affinity:
      {{- include "nodeaffinity" . | indent 6 }}
`

const clusterRoleYamlTemplate = `
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: {{ .ClusterRoleName }}
  labels:
    app: security
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["create", "get", "watch", "list", "update", "delete"]
- apiGroups: [""]
  resources: ["serviceaccounts"]
  verbs: ["get", "watch", "list"]
- apiGroups: [""]
  resources: ["services"]
  verbs: ["get", "watch", "list"]
`

const mtlsMeshPolicyYamlTemplate = `
# Authentication policy to enable mutual TLS for all services (that have sidecar) in the mesh.
apiVersion: "authentication.istio.io/v1alpha1"
kind: "MeshPolicy"
metadata:
  name: "default"
  labels:
    app: security
spec:
  peers:
  - mtls: {}
`

const mtlsMeshDestinationRuleListYamlTemplate = `
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRuleList
items:
  # Corresponding destination rule to configure client side to use mutual TLS when talking to
  # any service (host) in the mesh.
  - apiVersion: networking.istio.io/v1alpha3
    kind: DestinationRule
    metadata:
      name: "default"
      labels:
        app: security
    spec:
      host: "*.local"
      trafficPolicy:
        tls:
          mode: ISTIO_MUTUAL
  # Destination rule to disable (m)TLS when talking to API server, as API server doesn't have sidecar.
  # Customer should add similar destination rules for other services that don't have sidecar.
  - apiVersion: networking.istio.io/v1alpha3
    kind: DestinationRule
    metadata:
      name: "api-server"
      labels:
        app: security
    spec:
      host: "kubernetes.default.svc.cluster.local"
      trafficPolicy:
        tls:
          mode: DISABLE
`

const permissiveMeshPolicyYamlTemplate = `
# Authentication policy to enable permissive mode for all services (that have sidecar) in the mesh.
apiVersion: "authentication.istio.io/v1alpha1"
kind: "MeshPolicy"
metadata:
  name: "default"
  labels:
    app: security
spec:
  peers:
  - mtls:
      mode: PERMISSIVE
`
