package istiocontrolplane

import (
	"context"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/helm/pkg/releaseutil"

	"github.com/go-logr/logr"

	istiov1alpha3 "github.com/maistra/istio-operator/pkg/apis/istio/v1alpha3"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/helm/pkg/manifest"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type controlPlaneReconciler struct {
	*ReconcileIstioControlPlane
	log        logr.Logger
	instance   *istiov1alpha3.IstioControlPlane
	status     istiov1alpha3.IstioControlPlaneStatus
	ownerRefs  []metav1.OwnerReference
	renderings map[string][]manifest.Manifest
}

func (r *controlPlaneReconciler) Reconcile() (reconcile.Result, error) {
	// XXX: figure out error handling

	var err error

	// Render the templates
	r.renderings, r.status.Release, err = RenderHelmChart(ChartPath, r.instance)

	// create project
	// XXX: I don't think this should be necessary, as we should be creating
	// the control plane in the same project containing CR

	// set the auto-injection flag

	// add-scc-to-user anyuid to service accounts: citadel, egressgateway, galley, ingressgateway, mixer, pilot, sidecar-injector
	// plus: grafana, prometheus

	// add-scc-to-user privileged service accounts: jaeger

	// install istio
	r.installIstio()

	// install grafana

	// install jaeger

	// install kiali

	// install 3scale

	// install launcher

}

func (r *controlPlaneReconciler) installIstio() error {
	// update injection label on namespace
	// XXX: this should probably only be done when installing a control plane
	// which is all we're supporting atm.  if the scope expands to allow
	// installing custom gateways, etc., we should revisit this.
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: r.instance.Namespace}}
	err := r.client.Get(context.TODO(), client.ObjectKey{Name: r.instance.Namespace}, namespace)
	if label, ok := namespace.Labels["istio.openshift.com/ignore-namespace"]; !ok || label != "ignore" {
		r.log.V(1).Info("Adding istio.openshift.com/ignore-namespace=ignore label to Request.Namespace")
		namespace.Labels["istio.openshift.com/ignore-namespace"] = "ignore"
		err = r.client.Update(context.TODO(), namespace)
	}

	// ensure crds - move to bootstrap at operator startup

	// wait for crd availability - we should block bootstrapping until the crds are available

	// create components

	r.ownerRefs = []metav1.OwnerReference{*metav1.NewControllerRef(r.instance, r.instance.GroupVersionKind())}

	// create core istio resources
	if manifests, ok := r.renderings["istio"]; ok {
		for _, manifest := range manifests {
			if !strings.HasSuffix(manifest.Name, ".yaml") {
				r.log.V(2).Info("Skipping rendering of file", "file", manifest.Name)
				continue
			}
			r.log.V(2).Info("Processing resources from file", "file", manifest.Name)
			// split the manifest into individual objects
			objects := releaseutil.SplitManifests(manifest.Content)
			// Add owner ref
			for _, raw := range objects {
				rawUnstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(raw)
				if err != nil {
					// XXX: error handling
					r.log.Error(err, "unable to decode object", "object", raw)
					continue
				}
				// XXX: do we need special processing for kind: List?
				unstructured := &unstructured.Unstructured{Object: rawUnstructured}
				unstructured.SetOwnerReferences(r.ownerRefs)
				key := istiov1alpha3.NewResourceKey(unstructured, unstructured)
				condition := r.instance.Status.ResourceConditions[key]
				receiver := key.ToUnstructured()
				objectKey, err := client.ObjectKeyFromObject(receiver)
				if err != nil {
					// XXX: error handling
					continue
				}
				err = r.client.Get(context.TODO(), objectKey, receiver)
				if err != nil {
					if errors.IsNotFound(err) {
						err = r.client.Create(context.TODO(), unstructured)
						condition = &istiov1alpha3.Condition{}
					} else {
						// XXX: error handling
						continue
					}
				} else if (condition != nil && receiver.GetGeneration() != condition.Generation) || // somebody changed the object out from under us
					!(reflect.DeepEqual(unstructured.GetAnnotations(), receiver.GetAnnotations()) &&
						reflect.DeepEqual(unstructured.GetLabels(), receiver.GetLabels()) &&
						reflect.DeepEqual(unstructured.UnstructuredContent()["spec"], receiver.UnstructuredContent()["spec"])) {
					err = r.client.Update(context.TODO(), unstructured)
				} else {
					// nothing to do
					r.status.ResourceConditions[key] = condition
					continue
				}
				if err != nil {
					// XXX: error handling
					continue
				}
				condition.Generation = unstructured.GetGeneration()
				condition.Type = istiov1alpha3.ConditionTypeDeployed
				condition.Message = ""
				condition.Reason = istiov1alpha3.ConditionReasonInstallSuccessful
				condition.Status = istiov1alpha3.ConditionStatusTrue
				r.status.ResourceConditions[key] = condition

				// special handling
			}
		}
	}

	// create galley

	// create pilot

	// create mixer policy

	// create mixer telemetry

	// create security ???  not sure who depends on this, nodeagent for sure

	// create certmanager ???

	// sidecar injector???  should this be further up the list?

	// ingress

	// gateways

	// XXX: waiting is important for the follow-on components
	// wait for galley

	// wait for validating webhook to reconfigure

	// wait for sidecar injector deployment

	// wait for mutating webhook

	// patch mutating webhook (looks like it adds owner ref and restricts deletion)

	// set authenticated cluster role to enable access for all users

	// create route for ingress gateway service

	// create route for prometheus service
}
