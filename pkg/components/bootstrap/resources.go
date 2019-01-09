package bootstrap

import (
    "sync"
    "text/template"
)

type TemplateParams struct {
}

type templates struct {
    CRDsTemplate *template.Template
}

var (
    _singleton *templates
    _init sync.Once
)
func Templates() *templates {
    _init.Do(func() {
        _singleton = &templates{
            CRDsTemplate: template.New("CRDsList.yaml"),
        }
        _singleton.CRDsTemplate.Parse(istioCRDsYaml)
    })
    return _singleton
}

// XXX: ideally, we'd pull these directly from the resource files in istio/istio
const istioCRDsYaml = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinitionList
items:

  # Pilot CRDs
  apiVersion: apiextensions.k8s.io/v1beta1
  kind: CustomResourceDefinition
  metadata:
    name: virtualservices.networking.istio.io
    labels:
      app: istio-operator
      istio: pilot
  spec:
    group: networking.istio.io
    names:
      kind: VirtualService
      listKind: VirtualServiceList
      plural: virtualservices
      singular: virtualservice
      categories:
      - istio-io
      - networking-istio-io
    scope: Namespaced
    version: v1alpha3

  apiVersion: apiextensions.k8s.io/v1beta1
  kind: CustomResourceDefinition
  metadata:
    name: destinationrules.networking.istio.io
    labels:
      app: istio-operator
      istio: pilot
  spec:
    group: networking.istio.io
    names:
      kind: DestinationRule
      listKind: DestinationRuleList
      plural: destinationrules
      singular: destinationrule
      categories:
      - istio-io
      - networking-istio-io
    scope: Namespaced
    version: v1alpha3

  apiVersion: apiextensions.k8s.io/v1beta1
  kind: CustomResourceDefinition
  metadata:
    name: serviceentries.networking.istio.io
    labels:
      app: istio-operator
      istio: pilot
  spec:
    group: networking.istio.io
    names:
      kind: ServiceEntry
      listKind: ServiceEntryList
      plural: serviceentries
      singular: serviceentry
      categories:
      - istio-io
      - networking-istio-io
    scope: Namespaced
    version: v1alpha3

  apiVersion: apiextensions.k8s.io/v1beta1
  kind: CustomResourceDefinition
  metadata:
    name: gateways.networking.istio.io
    labels:
      app: istio-operator
      istio: pilot
  spec:
    group: networking.istio.io
    names:
      kind: Gateway
      plural: gateways
      singular: gateway
      categories:
      - istio-io
      - networking-istio-io
    scope: Namespaced
    version: v1alpha3 

  apiVersion: apiextensions.k8s.io/v1beta1
  kind: CustomResourceDefinition
  metadata:
    name: envoyfilters.networking.istio.io
    labels:
      app: istio-operator
      istio: pilot
  spec:
    group: networking.istio.io
    names:
      kind: EnvoyFilter
      plural: envoyfilters
      singular: envoyfilter
      categories:
      - istio-io
      - networking-istio-io
    scope: Namespaced
    version: v1alpha3

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: clusterrbacconfigs.rbac.istio.io
    labels:
      app: istio-operator
      istio: rbac
  spec:
    group: rbac.istio.io
    names:
      kind: ClusterRbacConfig
      plural: clusterrbacconfigs
      singular: clusterrbacconfig
      categories:
      - istio-io
      - rbac-istio-io
    scope: Cluster
    version: v1alpha1

  # Citadel CRDs
  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: policies.authentication.istio.io
    labels:
      app: istio-operator
      istio: citadel
  spec:
    group: authentication.istio.io
    names:
      kind: Policy
      plural: policies
      singular: policy
      categories:
      - istio-io
      - authentication-istio-io
    scope: Namespaced
    version: v1alpha1

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: meshpolicies.authentication.istio.io
    labels:
      app: istio-operator
      istio: citadel
  spec:
    group: authentication.istio.io
    names:
      kind: MeshPolicy
      listKind: MeshPolicyList
      plural: meshpolicies
      singular: meshpolicy
      categories:
      - istio-io
      - authentication-istio-io
    scope: Cluster
    version: v1alpha1

  # Policy CRDs
  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: httpapispecbindings.config.istio.io
    labels:
      app: istio-operator
      istio: policy
  spec:
    group: config.istio.io
    names:
      kind: HTTPAPISpecBinding
      plural: httpapispecbindings
      singular: httpapispecbinding
      categories:
      - istio-io
      - apim-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: httpapispecs.config.istio.io
    labels:
      app: istio-operator
      istio: policy
  spec:
    group: config.istio.io
    names:
      kind: HTTPAPISpec
      plural: httpapispecs
      singular: httpapispec
      categories:
      - istio-io
      - apim-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: quotaspecbindings.config.istio.io
    labels:
      app: istio-operator
      istio: policy
  spec:
    group: config.istio.io
    names:
      kind: QuotaSpecBinding
      plural: quotaspecbindings
      singular: quotaspecbinding
      categories:
      - istio-io
      - apim-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: quotaspecs.config.istio.io
    labels:
      app: istio-operator
      istio: policy
  spec:
    group: config.istio.io
    names:
      kind: QuotaSpec
      plural: quotaspecs
      singular: quotaspec
      categories:
      - istio-io
      - apim-istio-io
    scope: Namespaced
    version: v1alpha2

  # Mixer CRDs
  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: rules.config.istio.io
    labels:
      app: istio-operator
      istio: mixer
      package: istio.io.mixer
  spec:
    group: config.istio.io
    names:
      kind: rule
      plural: rules
      singular: rule
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: attributemanifests.config.istio.io
    labels:
      app: istio-operator
      istio: mixer
      package: istio.io.mixer
  spec:
    group: config.istio.io
    names:
      kind: attributemanifest
      plural: attributemanifests
      singular: attributemanifest
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: bypasses.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-adapter
      package: bypass
  spec:
    group: config.istio.io
    names:
      kind: bypass
      plural: bypasses
      singular: bypass
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: circonuses.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-adapter
      package: circonus
  spec:
    group: config.istio.io
    names:
      kind: circonus
      plural: circonuses
      singular: circonus
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: deniers.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-adapter
      package: denier
  spec:
    group: config.istio.io
    names:
      kind: denier
      plural: deniers
      singular: denier
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: fluentds.config.istio.io
    annotations:
      "helm.sh/hook": crd-install
      labels:
      app: istio-operator
      istio: mixer-adapter
      package: fluentd
  spec:
    group: config.istio.io
    names:
      kind: fluentd
      plural: fluentds
      singular: fluentd
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: kubernetesenvs.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-adapter
      package: kubernetesenv
  spec:
    group: config.istio.io
    names:
      kind: kubernetesenv
      plural: kubernetesenvs
      singular: kubernetesenv
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: listcheckers.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-adapter
      package: listchecker
  spec:
    group: config.istio.io
    names:
      kind: listchecker
      plural: listcheckers
      singular: listchecker
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: memquotas.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-adapter
      package: memquota
  spec:
    group: config.istio.io
    names:
      kind: memquota
      plural: memquotas
      singular: memquota
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: noops.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-adapter
      package: noop
  spec:
    group: config.istio.io
    names:
      kind: noop
      plural: noops
      singular: noop
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: opas.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-adapter
      package: opa
  spec:
    group: config.istio.io
    names:
      kind: opa
      plural: opas
      singular: opa
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: prometheuses.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-adapter
      package: prometheus
  spec:
    group: config.istio.io
    names:
      kind: prometheus
      plural: prometheuses
      singular: prometheus
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: rbacs.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-adapter
      package: rbac
  spec:
    group: config.istio.io
    names:
      kind: rbac
      plural: rbacs
      singular: rbac
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: redisquotas.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-adapter
      package: redisquota
  spec:
    group: config.istio.io
    names:
      kind: redisquota
      plural: redisquotas
      singular: redisquota
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: servicecontrols.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-adapter
      package: servicecontrol
  spec:
    group: config.istio.io
    names:
      kind: servicecontrol
      plural: servicecontrols
      singular: servicecontrol
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: signalfxs.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-adapter
      package: signalfx
  spec:
    group: config.istio.io
    names:
      kind: signalfx
      plural: signalfxs
      singular: signalfx
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: solarwindses.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-adapter
      package: solarwinds
  spec:
    group: config.istio.io
    names:
      kind: solarwinds
      plural: solarwindses
      singular: solarwinds
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: stackdrivers.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-adapter
      package: stackdriver
  spec:
    group: config.istio.io
    names:
      kind: stackdriver
      plural: stackdrivers
      singular: stackdriver
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: cloudwatches.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-adapter
      package: cloudwatch
  spec:
    group: config.istio.io
    names:
      kind: cloudwatch
      plural: cloudwatches
      singular: cloudwatch
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: dogstatsds.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-adapter
      package: dogstatsd
  spec:
    group: config.istio.io
    names:
      kind: dogstatsd
      plural: dogstatsds
      singular: dogstatsd
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: statsds.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-adapter
      package: statsd
  spec:
    group: config.istio.io
    names:
      kind: statsd
      plural: statsds
      singular: statsd
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: stdios.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-adapter
      package: stdio
  spec:
    group: config.istio.io
    names:
      kind: stdio
      plural: stdios
      singular: stdio
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: apikeys.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-instance
      package: apikey
  spec:
    group: config.istio.io
    names:
      kind: apikey
      plural: apikeys
      singular: apikey
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: authorizations.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-instance
      package: authorization
  spec:
    group: config.istio.io
    names:
      kind: authorization
      plural: authorizations
      singular: authorization
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: checknothings.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-instance
      package: checknothing
  spec:
    group: config.istio.io
    names:
      kind: checknothing
      plural: checknothings
      singular: checknothing
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: kuberneteses.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-instance
      package: adapter.template.kubernetes
  spec:
    group: config.istio.io
    names:
      kind: kubernetes
      plural: kuberneteses
      singular: kubernetes
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: listentries.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-instance
      package: listentry
  spec:
    group: config.istio.io
    names:
      kind: listentry
      plural: listentries
      singular: listentry
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: logentries.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-instance
      package: logentry
  spec:
    group: config.istio.io
    names:
      kind: logentry
      plural: logentries
      singular: logentry
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: edges.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-instance
      package: edge
  spec:
    group: config.istio.io
    names:
      kind: edge
      plural: edges
      singular: edge
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: metrics.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-instance
      package: metric
  spec:
    group: config.istio.io
    names:
      kind: metric
      plural: metrics
      singular: metric
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: quotas.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-instance
      package: quota
  spec:
    group: config.istio.io
    names:
      kind: quota
      plural: quotas
      singular: quota
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: reportnothings.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-instance
      package: reportnothing
  spec:
    group: config.istio.io
    names:
      kind: reportnothing
      plural: reportnothings
      singular: reportnothing
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: servicecontrolreports.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-instance
      package: servicecontrolreport
  spec:
    group: config.istio.io
    names:
      kind: servicecontrolreport
      plural: servicecontrolreports
      singular: servicecontrolreport
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: tracespans.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-instance
      package: tracespan
  spec:
    group: config.istio.io
    names:
      kind: tracespan
      plural: tracespans
      singular: tracespan
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: rbacconfigs.rbac.istio.io
    labels:
      app: istio-operator
      istio: rbac
      package: istio.io.mixer
  spec:
    group: rbac.istio.io
    names:
      kind: RbacConfig
      plural: rbacconfigs
      singular: rbacconfig
      categories:
      - istio-io
      - rbac-istio-io
    scope: Namespaced
    version: v1alpha1

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: serviceroles.rbac.istio.io
    labels:
      app: istio-operator
      istio: rbac
      package: istio.io.mixer
  spec:
    group: rbac.istio.io
    names:
      kind: ServiceRole
      plural: serviceroles
      singular: servicerole
      categories:
      - istio-io
      - rbac-istio-io
    scope: Namespaced
    version: v1alpha1

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: servicerolebindings.rbac.istio.io
    labels:
      app: istio-operator
      istio: rbac
      package: istio.io.mixer
  spec:
    group: rbac.istio.io
    names:
      kind: ServiceRoleBinding
      plural: servicerolebindings
      singular: servicerolebinding
      categories:
      - istio-io
      - rbac-istio-io
    scope: Namespaced
    version: v1alpha1

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: adapters.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-adapter
      package: adapter
  spec:
    group: config.istio.io
    names:
      kind: adapter
      plural: adapters
      singular: adapter
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: instances.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-instance
      package: instance
  spec:
    group: config.istio.io
    names:
      kind: instance
      plural: instances
      singular: instance
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: templates.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-template
      package: template
  spec:
    group: config.istio.io
    names:
      kind: template
      plural: templates
      singular: template
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2

  kind: CustomResourceDefinition
  apiVersion: apiextensions.k8s.io/v1beta1
  metadata:
    name: handlers.config.istio.io
    labels:
      app: istio-operator
      istio: mixer-handler
      package: handler
  spec:
    group: config.istio.io
    names:
      kind: handler
      plural: handlers
      singular: handler
      categories:
      - istio-io
      - policy-istio-io
    scope: Namespaced
    version: v1alpha2
`
