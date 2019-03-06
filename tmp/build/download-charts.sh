#!/usr/bin/env bash
DIR=$(pwd)/tmp/_output/helm

ISTIO_VERSION=1.1.0-rc.2
#ISTIO_BRANCH=release-1.1

RELEASES_DIR=${DIR}/istio-releases

PLATFORM=linux
if [ -n "${ISTIO_VERSION}" ] ; then
  ISTIO_NAME=istio-${ISTIO_VERSION}
  ISTIO_FILE="${ISTIO_NAME}-${PLATFORM}.tar.gz"
  ISTIO_URL="https://github.com/istio/istio/releases/download/${ISTIO_VERSION}/${ISTIO_FILE}"
  EXTRACT_CMD="tar --strip-components=4 -C ./${ISTIO_NAME} -xvzf ${ISTIO_FILE} ${ISTIO_NAME}/install/kubernetes/helm"
  RELEASE_DIR="${RELEASES_DIR}/${ISTIO_NAME}"
else
  ISTIO_NAME=istio-${ISTIO_BRANCH}
  ISTIO_FILE="${ISTIO_BRANCH}.zip"
  ISTIO_URL="https://github.com/istio/istio/archive/${ISTIO_FILE}"
  EXTRACT_CMD="unzip ${ISTIO_FILE} ${ISTIO_NAME}/install/kubernetes/helm"
  RELEASE_DIR="${RELEASES_DIR}/${ISTIO_NAME}"
fi

ISTIO_NAME=${ISTIO_NAME//./-}

HELM_DIR=${RELEASE_DIR}

PATCH_1_0=$(echo "${ISTIO_VERSION}" | grep "^1\.0\..*")

COMMUNITY=${COMMUNITY:-true}

function retrieveIstioRelease() {
  if [ -d "${RELEASE_DIR}" ] ; then
    rm -rf "${RELEASE_DIR}"
  fi
  mkdir -p "${RELEASE_DIR}"

  if [ ! -f "${RELEASES_DIR}/${ISTIO_FILE}" ] ; then
    (
      cd "${RELEASES_DIR}"
      curl -LO "${ISTIO_URL}"
    )
  fi

  (
      cd "${RELEASES_DIR}"
      ${EXTRACT_CMD}
      (
        cd "${HELM_DIR}/istio"
        helm dep update
      )
  )
}

# The following modifications are made to the generated helm template for the Istio yaml files
# - remove the create customer resources job, we handle this in the installer to deal with potential races
# - remove the cleanup secrets job, we handle this in the installer
# - remove the kubernetes gateways
# - change privileged value on istio-proxy injection configmap to false
# - update the namespaceSelector to ignore namespaces with the label istio.openshift.com/ignore-namespace
# - add a maistra-version label to all objects which have a release label
# - remove GODEBUG from the pilot environment
# - remove istio-multi service account
# - remove istio-multi cluster role binding
# - remove istio-reader cluster role
# - switch prometheus init container image from busybox to prometheus
# - switch webhook ports to 8443
# - switch health check files into /tmp
function patchTemplates() {
  # - remove the create customer resources job, we handle this in the installer to deal with potential races
  rm ${HELM_DIR}/istio/charts/grafana/templates/create-custom-resources-job.yaml

  # - remove the cleanup secrets job, we handle this in the installer
  rm ${HELM_DIR}/istio/charts/security/templates/cleanup-secrets.yaml

  # - remove the kubernetes gateways
  # this no longer exists
  # rm ${HELM_DIR}istio/charts/pilot/templates/gateway.yaml

  # - remove GODEBUG from the pilot environment (and mixer too)
  sed -i -e '/GODEBUG/d' ${HELM_DIR}/istio/charts/pilot/values.yaml ${HELM_DIR}/istio/charts/mixer/values.yaml

  # - change privileged value on istio-proxy injection configmap to false
  # setting the proper values will fix this:
  # global.proxy.privileged=false
  # global.proxy.enableCoreDump=false

  # - update the namespaceSelector to ignore namespaces with the label istio.openshift.com/ignore-namespace
  # set sidecarInjectorWebhook.enableNamespacesByDefault=true
  sed -i -e '/if \.Values\.enableNamespacesByDefault/,/else/s/istio-injection/istio.openshift.com\/ignore-namespace/' \
         -e 's/NotIn/DoesNotExist/' \
         -e '/values/d' \
         -e '/disabled/d' ${HELM_DIR}/istio/charts/sidecarInjectorWebhook/templates/mutatingwebhook.yaml

  # - add a maistra-version label to all objects which have a release label
  find ${HELM_DIR} -name "*.yaml" -o -name "*.yaml.tpl" | \
    xargs sed -i -e 's/^\(.*\)release:\(.*\)$/\1maistra-version: MAISTRA_VERSION\
\1release:\2/' 

  # update the hub value
  # set global.hub=docker.io/istio
  sed -i -e 's+gcr.io/istio-release+docker.io/istio+g' ${HELM_DIR}/istio/values.yaml ${HELM_DIR}/istio-init/values.yaml

  # - remove istio-multi service account
  rm ${HELM_DIR}/istio/templates/serviceaccount.yaml
  # - remove istio-multi cluster role binding
  rm ${HELM_DIR}/istio/templates/clusterrolebinding.yaml
  # - remove istio-reader cluster role
  rm ${HELM_DIR}/istio/templates/clusterrole.yaml
  
  # - switch prometheus init container image from busybox to prometheus
  sed -i -e 's/"?busybox:?.*$/"docker.io\/prom\/prometheus:v2.3.1"/' ${HELM_DIR}/istio/charts/prometheus/templates/deployment.yaml

  # - switch webhook ports to 8443: add targetPort name to galley service
  # XXX: move upstream (add targetPort name)
  sed -i -e 's/^\(.*\)\(- port: 443.*\)$/\1\2\
\1  targetPort: webhook/' ${HELM_DIR}/istio/charts/galley/templates/service.yaml

  # - switch webhook ports to 8443
  # add name to webhook port (XXX: move upstream)
  # change the location of the healthCheckFile from /health to /tmp/health
  # add --validation-port=8443
  sed -i -e 's/^\(.*\)\- containerPort: 443.*$/\1- name: webhook\
\1  containerPort: 8443/' \
           -e 's/\/health/\/tmp\/health/' \
           -e 's/^\(.*\)\(- --monitoringPort.*\)$/\1\2\
\1- --validation-port=8443/' ${HELM_DIR}/istio/charts/galley/templates/deployment.yaml

  # - switch webhook ports to 8443
  # XXX: move upstream (add targetPort name)
  sed -i -e 's/^\(.*\)\(- port: 443.*\)$/\1\2\
\1  targetPort: webhook/' ${HELM_DIR}/istio/charts/sidecarInjectorWebhook/templates/service.yaml

  # - switch webhook ports to 8443
  # add webhook port
  sed -i -e 's/^\(.*\)\(volumeMounts:.*\)$/\1  - --port=8443\
\1ports:\
\1- name: webhook\
\1  containerPort: 8443\
\1\2/' ${HELM_DIR}/istio/charts/sidecarInjectorWebhook/templates/deployment.yaml

  # change the location of the healthCheckFile from /health to /tmp/health
  if [[ "${COMMUNITY,,}" != "true" ]]; then
    sed -i -e 's/\/health/\/tmp\/health/' ${HELM_DIR}/istio/charts/sidecarInjectorWebhook/templates/deployment.yaml
  fi

}

# The following modifications are made to the generated helm template to extract the CRDs
# - remove all content except for the crd configmaps
# - add maistra-version labels
function patchCRDs() {
  find ${HELM_DIR}/istio-init/files -name "crd-*.yaml" | xargs sed -i -e 's/^\(.*\)labels:$/\1labels:\
\1  maistra-version: MAISTRA_VERSION/'
}

# The following modifications are made to the generated helm template for the Grafana yaml file
# - add a service account for grafana
# - remove all non grafana configuration
# - remove the extraneous create custom resources job
# - add the service account to the deployment
# - add a maistra-version label to all objects which have a release label (done in patchTemplates())
function patchGrafanaTemplate() {
  # - add a service account for grafana
  echo 'apiVersion: v1
kind: ServiceAccount
metadata:
  name: grafana
  namespace: istio-system
  labels:
    app: grafana
' > ${HELM_DIR}/istio/charts/grafana/templates/serviceaccount.yaml

  # - remove the extraneous create custom resources job
  if [ -f ${HELM_DIR}/istio/charts/grafana/templates/create-custom-resources-job.yaml ]; then
    rm ${HELM_DIR}/istio/charts/grafana/templates/create-custom-resources-job.yaml
  fi

  # - add the service account to the deployment
  sed -i -e 's/^\(.*\)containers:\(.*\)$/\1serviceAccountName: grafana\
\1containers:\2/' ${HELM_DIR}/istio/charts/grafana/templates/deployment.yaml
}

# The following modifications are made to the generated helm template for the Kiali yaml file
# - remove all non kiali configuration
# - remove the kiali username/password secret
function patchKialiTemplate() {

  # - remove the kiali username/password secret
  rm ${HELM_DIR}/istio/charts/kiali/templates/demosecret.yaml
}

# The following modifications are made to the upstream kiali configuration for deployment on OpenShift
# - Add jaeger and grafana URLs to the configmap as well as the identity certs
# - Add the route.openshift.io api group to the cluster role
# - Add the openshift annotation to the service
# - Remove the prometheus, grafana environment from the deployment
# - Add the kiali-cert volume mount
# - Add the kiali-cert volume
# - Add istio namespace to the configmap
function patchKialiOpenShift() {
  # - Add jaeger and grafana URLs to the configmap as well as the identity certs
  #   these should be defined in the values file kiali.dashboard.grafanaURL
  #   and kiali.dashboard.jaegerURL.  If not specified, the URLs should be
  #   determined from the routes created for Jaeger and Grafana.  This needs to
  #   be done after rendering, unfortunately.
  # - Add istio namespace to the configmap (for 1.0 templates)
  # - Add the kiali-cert volume mount
  # - Add the kiali-cert volume
  if [ -n "${PATCH_1_0}" ]; then
    sed -i -e '/server:/ i\
\    istio_namespace: {{ .Release.Namespace }}' ${HELM_DIR}/istio/charts/kiali/templates/configmap.yaml
  fi
  sed -i -e '/port: 20001/ a\
\      static_content_root_directory: /opt/kiali/console' \
         -e '/grafana:/,/url:/ {
             /url:/ a\
\    identity:\
\      cert_file: /kiali-cert/tls.crt\
\      private_key_file: /kiali-cert/tls.key
             }' ${HELM_DIR}/istio/charts/kiali/templates/configmap.yaml

  # - Add the route.openshift.io api group to the cluster role
  sed -i -e '/apiGroups:.*config.istio.io/ i\
\- apiGroups: ["project.openshift.io"]\
\  resources:\
\  - projects\
\  verbs:\
\  - get\
\- apiGroups: ["route.openshift.io"]\
\  resources:\
\  - routes\
\  verbs:\
\  - get\
\- apiGroups: ["apps.openshift.io"]\
\  resources:\
\  - deploymentconfigs\
\  verbs:\
\  - get\
\  - list\
\  - watch' ${HELM_DIR}/istio/charts/kiali/templates/clusterrole.yaml

  # - Add the openshift annotation to the service
  sed -i -e '/metadata/ a\
\  annotations:\
\    service.alpha.openshift.io/serving-cert-secret-name: kiali-cert-secret' ${HELM_DIR}/istio/charts/kiali/templates/service.yaml
  
  # - Remove the prometheus, grafana environment from the deployment
  sed -i -e '/PROMETHEUS_SERVICE_URL/,/volumeMounts/ {
    /volumeMounts/b
    d
  }' ${HELM_DIR}/istio/charts/kiali/templates/deployment.yaml
  
  # - Add the kiali-cert volume mount
  # - Add the kiali-cert volume
  sed -i -e '/kind.*Deployment$/,/^.*affinity:/ {
    /volumeMounts:/ {
      N
      N
      a\
\        - name: kiali-cert\
\          mountPath: "/kiali-cert"
    }
    /configMap:/ {
      N
      a\
\      - name: kiali-cert\
\        secret:\
\          secretName: kiali-cert-secret
    }
  }' ${HELM_DIR}/istio/charts/kiali/templates/deployment.yaml

  # - add authorization details for authentication.istio.io (for 1.0 templates)
  if [ -n "${PATCH_1_0}" ]; then
    sed -i -e '/apiGroups:.*networking.istio.io/,/^-/ {
      /- watch/ a\
\- apiGroups: ["authentication.istio.io"]\
\  resources:\
\  - policies\
\  verbs:\
\  - create\
\  - delete\
\  - get\
\  - list\
\  - patch\
\  - watch
    }' ${HELM_DIR}/istio/charts/kiali/templates/clusterrole.yaml
  fi

  # add monitoring.kiali.io
  sed -i -e '/apiGroups:.*authentication.istio.io/,/^-/ {
      /- watch/ a\
\- apiGroups: ["monitoring.kiali.io"]\
\  resources:\
\  - monitoringdashboards\
\  verbs:\
\  - get
  }' ${HELM_DIR}/istio/charts/kiali/templates/clusterrole.yaml
  
  # - add create verb to config.istio.io (for 1.0 templates)
  if [ -n "${PATCH_1_0}" ]; then
    sed -i -e '/apiGroups:.*config.istio.io/,/^-/ {
      /verbs:/ a\
\  - create
    }' ${HELM_DIR}/istio/charts/kiali/templates/clusterrole.yaml
  fi
  
  # - add create verb to networking.istio.io  (for 1.0 templates)
  if [ -n "${PATCH_1_0}" ]; then
    sed -i -e '/apiGroups:.*networking.istio.io/,/^-/ {
      /verbs:/ a\
\  - create
    }' ${HELM_DIR}/istio/charts/kiali/templates/clusterrole.yaml
  fi
}

retrieveIstioRelease

patchTemplates
patchCRDs
patchGrafanaTemplate
patchKialiTemplate
patchKialiOpenShift
