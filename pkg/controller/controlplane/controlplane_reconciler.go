package controlplane

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
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/helm/pkg/manifest"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type controlPlaneReconciler struct {
	*ReconcileControlPlane
	log        logr.Logger
	instance   *istiov1alpha3.ControlPlane
	status     istiov1alpha3.ControlPlaneStatus
	ownerRefs  []metav1.OwnerReference
	renderings map[string][]manifest.Manifest
}

func (r *controlPlaneReconciler) Reconcile() (reconcile.Result, error) {
	// XXX: can we get away with this?
	if r.instance.Status.ObservedGeneration >= r.instance.GetGeneration() {
		return reconcile.Result{}, nil
	}
	var err error

	// Render the templates
	r.renderings, r.status.ReleaseInfo, err = RenderHelmChart(ChartPath, r.instance)

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

var seen = struct{}{}

type customizationHook func(object *unstructured.Unstructured) error

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

	componentsProcessed := map[string]struct{}{}

	// create core istio resources
	componentsProcessed["istio"] = seen
	componentStatus, ok := r.instance.Status.ComponentStatus["istio"]
	if !ok {
		componentStatus = istiov1alpha3.NewComponentStatus()
		r.instance.Status.ComponentStatus["istio"] = componentStatus
	}
	componentStatus.RemoveCondition(istiov1alpha3.ConditionTypeReconciled)
	err = r.processManifests(r.renderings["istio"], &componentStatus, nil, nil)
	if err == nil {
		if condition := componentStatus.GetCondition(istiov1alpha3.ConditionTypeInstalled); condition.Status != istiov1alpha3.ConditionStatusTrue {
			componentStatus.SetCondition(istiov1alpha3.Condition{
				Type:   istiov1alpha3.ConditionTypeInstalled,
				Reason: istiov1alpha3.ConditionReasonInstallSuccessful,
				Status: istiov1alpha3.ConditionStatusTrue,
			})
		}
		componentStatus.SetCondition(istiov1alpha3.Condition{
			Type:   istiov1alpha3.ConditionTypeReconciled,
			Reason: istiov1alpha3.ConditionReasonReconcileSuccessful,
			Status: istiov1alpha3.ConditionStatusTrue,
		})
	} else {
		if condition := componentStatus.GetCondition(istiov1alpha3.ConditionTypeInstalled); condition.Status == istiov1alpha3.ConditionStatusUnknown {
			componentStatus.SetCondition(istiov1alpha3.Condition{
				Type:    istiov1alpha3.ConditionTypeInstalled,
				Reason:  istiov1alpha3.ConditionReasonInstallError,
				Status:  istiov1alpha3.ConditionStatusFalse,
				Message: err.Error(),
			})
			componentStatus.SetCondition(istiov1alpha3.Condition{
				Type:    istiov1alpha3.ConditionTypeReconciled,
				Reason:  istiov1alpha3.ConditionReasonInstallError,
				Status:  istiov1alpha3.ConditionStatusFalse,
				Message: err.Error(),
			})
		} else {
			componentStatus.SetCondition(istiov1alpha3.Condition{
				Type:    istiov1alpha3.ConditionTypeReconciled,
				Reason:  istiov1alpha3.ConditionReasonReconcileError,
				Status:  istiov1alpha3.ConditionStatusFalse,
				Message: err.Error(),
			})
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

func noopCustimizationHook(_ *unstructured.Unstructured) error { return nil }

func (r *controlPlaneReconciler) processManifests(manifests []manifest.Manifest,
	componentStatus *istiov1alpha3.ComponentStatus,
	processNewObject customizationHook,
	processDeletedObject customizationHook) error {

	allErrors := []error{}
	resourcesProcessed := map[istiov1alpha3.ResourceKey]struct{}{}
	if processNewObject == nil {
		processNewObject = noopCustimizationHook
	}
	if processDeletedObject == nil {
		processDeletedObject = noopCustimizationHook
	}

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
				allErrors = append(allErrors, err)
				continue
			}

			// XXX: do we need special processing for kind: List?
			unstructured := &unstructured.Unstructured{Object: rawUnstructured}
			unstructured.SetOwnerReferences(r.ownerRefs)

			key := istiov1alpha3.NewResourceKey(unstructured, unstructured)
			resourcesProcessed[key] = seen
			status, ok := componentStatus.ResourceStatus[key]
			if !ok {
				status = istiov1alpha3.NewStatus()
				componentStatus.ResourceStatus[key] = status
			}

			status.RemoveCondition(istiov1alpha3.ConditionTypeReconciled)

			receiver := key.ToUnstructured()
			objectKey, err := client.ObjectKeyFromObject(receiver)
			if err != nil {
				// This can only happen if reciever isn't an unstructured.Unstructured
				status.SetCondition(istiov1alpha3.Condition{
					Type:    istiov1alpha3.ConditionTypeReconciled,
					Reason:  istiov1alpha3.ConditionReasonReconcileError,
					Status:  istiov1alpha3.ConditionStatusFalse,
					Message: err.Error(),
				})
				allErrors = append(allErrors, err)
				continue
			}
			err = r.client.Get(context.TODO(), objectKey, receiver)
			if err != nil {
				if errors.IsNotFound(err) {
					err = r.client.Create(context.TODO(), unstructured)
					if err == nil {
						status.SetCondition(istiov1alpha3.Condition{
							Type:   istiov1alpha3.ConditionTypeInstalled,
							Reason: istiov1alpha3.ConditionReasonInstallSuccessful,
							Status: istiov1alpha3.ConditionStatusTrue,
						})
						status.ObservedGeneration = unstructured.GetGeneration()
					} else {
						status.SetCondition(istiov1alpha3.Condition{
							Type:    istiov1alpha3.ConditionTypeInstalled,
							Reason:  istiov1alpha3.ConditionReasonInstallError,
							Status:  istiov1alpha3.ConditionStatusFalse,
							Message: err.Error(),
						})
						status.SetCondition(istiov1alpha3.Condition{
							Type:    istiov1alpha3.ConditionTypeReconciled,
							Reason:  istiov1alpha3.ConditionReasonInstallError,
							Status:  istiov1alpha3.ConditionStatusFalse,
							Message: err.Error(),
						})
						allErrors = append(allErrors, err)
						continue
					}
				} else {
					status.SetCondition(istiov1alpha3.Condition{
						Type:    istiov1alpha3.ConditionTypeReconciled,
						Reason:  istiov1alpha3.ConditionReasonReconcileError,
						Status:  istiov1alpha3.ConditionStatusFalse,
						Message: err.Error(),
					})
					allErrors = append(allErrors, err)
					continue
				}
			} else if (receiver.GetGeneration() > status.ObservedGeneration) || // somebody changed the object out from under us
				!(reflect.DeepEqual(unstructured.GetAnnotations(), receiver.GetAnnotations()) &&
					reflect.DeepEqual(unstructured.GetLabels(), receiver.GetLabels())) ||
				shouldUpdate(unstructured.UnstructuredContent(), receiver.UnstructuredContent()) {
				err = r.client.Update(context.TODO(), unstructured)
				if err == nil {
					status.SetCondition(istiov1alpha3.Condition{
						Type:   istiov1alpha3.ConditionTypeReconciled,
						Reason: istiov1alpha3.ConditionReasonReconcileSuccessful,
						Status: istiov1alpha3.ConditionStatusTrue,
					})
					status.ObservedGeneration = unstructured.GetGeneration()
				} else {
					status.SetCondition(istiov1alpha3.Condition{
						Type:    istiov1alpha3.ConditionTypeReconciled,
						Reason:  istiov1alpha3.ConditionReasonUpdateError,
						Status:  istiov1alpha3.ConditionStatusFalse,
						Message: err.Error(),
					})
					allErrors = append(allErrors, err)
					continue
				}
			} else {
				// nothing to do
				status.SetCondition(istiov1alpha3.Condition{
					Type:   istiov1alpha3.ConditionTypeReconciled,
					Reason: istiov1alpha3.ConditionReasonReconcileSuccessful,
					Status: istiov1alpha3.ConditionStatusTrue,
				})
			}

			// special handling
			processNewObject(unstructured)
		}
	}

	// handle deletions
	for key, status := range componentStatus.ResourceStatus {
		if _, ok := resourcesProcessed[key]; !ok {
			if condition := status.GetCondition(istiov1alpha3.ConditionTypeInstalled); condition.Status == istiov1alpha3.ConditionStatusTrue {
				unstructured := key.ToUnstructured()
				err := r.client.Delete(context.TODO(), unstructured, client.PropagationPolicy(metav1.DeletePropagationBackground))
				if err == nil || errors.IsNotFound(err) {
					status.SetCondition(istiov1alpha3.Condition{
						Type:   istiov1alpha3.ConditionTypeInstalled,
						Status: istiov1alpha3.ConditionStatusFalse,
						Reason: istiov1alpha3.ConditionReasonDeletionSuccessful,
					})
					status.SetCondition(istiov1alpha3.Condition{
						Type:   istiov1alpha3.ConditionTypeReconciled,
						Status: istiov1alpha3.ConditionStatusTrue,
						Reason: istiov1alpha3.ConditionReasonDeletionSuccessful,
					})
					// special handling
					processDeletedObject(unstructured)
				} else {
					status.SetCondition(istiov1alpha3.Condition{
						Type:    istiov1alpha3.ConditionTypeReconciled,
						Status:  istiov1alpha3.ConditionStatusFalse,
						Reason:  istiov1alpha3.ConditionReasonDeletionError,
						Message: err.Error(),
					})
					allErrors = append(allErrors, err)
				}
			}
		}
	}
	return utilerrors.NewAggregate(allErrors)
}

// shouldUpdate checks to see if the spec fields are the same for both objects.
// if the objects don't have a spec field, it checks all other fields, skipping
// known fields that shouldn't impact updates: kind, apiVersion, metadata, and status.
func shouldUpdate(o1, o2 map[string]interface{}) bool {
	if spec1, ok1 := o1["spec"]; ok1 {
		// we assume these are the same type of object
		return reflect.DeepEqual(spec1, o2["spec"])
	}
	for key, value := range o1 {
		if key == "status" || key == "kind" || key == "apiVersion" || key == "metadata" {
			continue
		}
		if !reflect.DeepEqual(value, o2[key]) {
			return true
		}
	}
	return false
}
