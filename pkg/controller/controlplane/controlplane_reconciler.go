package controlplane

import (
	"context"
	"fmt"
	"path"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"

	istiov1alpha3 "github.com/maistra/istio-operator/pkg/apis/istio/v1alpha3"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"k8s.io/apiserver/pkg/authentication/serviceaccount"

	"k8s.io/helm/pkg/manifest"
	"k8s.io/helm/pkg/releaseutil"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type controlPlaneReconciler struct {
	*ReconcileControlPlane
	log        logr.Logger
	instance   *istiov1alpha3.ControlPlane
	ownerRefs  []metav1.OwnerReference
	renderings map[string][]manifest.Manifest
}

var seen = struct{}{}

func (r *controlPlaneReconciler) Delete() (reconcile.Result, error) {
	allErrors := []error{}
	for key := range r.instance.Status.ComponentStatus {
		err := r.processComponentManifests(key)
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}

	err := utilerrors.NewAggregate(allErrors)
	updateDeleteStatus(&r.instance.Status.StatusType, err)

	updateErr := r.client.Status().Update(context.TODO(), r.instance)
	if updateErr != nil {
		r.log.Error(err, "error updating ControlPlane status for object", "object", r.instance.GetName())
		if err == nil {
			// XXX: is this the right thing to do?
			return reconcile.Result{}, updateErr
		}
	}
	return reconcile.Result{}, err
}

func (r *controlPlaneReconciler) Reconcile() (reconcile.Result, error) {
	allErrors := []error{}
	var err error
	var launcherRenderings, threeScaleRenderings map[string][]manifest.Manifest

	// prepare to write a new reconciliation status
	r.instance.Status.RemoveCondition(istiov1alpha3.ConditionTypeReconciled)
	// ensure ComponentStatus is ready
	if r.instance.Status.ComponentStatus == nil {
		r.instance.Status.ComponentStatus = map[string]*istiov1alpha3.ComponentStatus{}
	}

	// Render the templates
	istioRenderings, _, err := RenderHelmChart(path.Join(ChartPath, "istio"), r.instance.GetNamespace(), r.instance.Spec.Istio)
	if err != nil {
		allErrors = append(allErrors, err)
	}
	if isEnabled(r.instance.Spec.Launcher) {
		launcherRenderings, _, err = RenderHelmChart(path.Join(ChartPath, "maistra-launcher"), r.instance.GetNamespace(), r.instance.Spec.Launcher)
		if err != nil {
			allErrors = append(allErrors, err)
		}
	} else {
		launcherRenderings = map[string][]manifest.Manifest{}
	}
	if isEnabled(r.instance.Spec.ThreeScale) {
		threeScaleRenderings, _, err = RenderHelmChart(path.Join(ChartPath, "maistra-threescale"), r.instance.GetNamespace(), r.instance.Spec.ThreeScale)
		if err != nil {
			allErrors = append(allErrors, err)
		}
	} else {
		threeScaleRenderings = map[string][]manifest.Manifest{}
	}
	if len(allErrors) > 0 {
		err = utilerrors.NewAggregate(allErrors)
		// we can't progress here
		updateReconcileStatus(&r.instance.Status.StatusType, err)
		r.client.Status().Update(context.TODO(), r.instance)
		return reconcile.Result{}, err
	}

	// merge the rendernings
	r.renderings = map[string][]manifest.Manifest{}
	for key, value := range istioRenderings {
		r.renderings[key] = value
	}
	for key, value := range launcherRenderings {
		r.renderings[key] = value
	}
	for key, value := range threeScaleRenderings {
		r.renderings[key] = value
	}

	// create project
	// XXX: I don't think this should be necessary, as we should be creating
	// the control plane in the same project containing CR

	// set the auto-injection flag

	// install istio
	// update injection label on namespace
	// XXX: this should probably only be done when installing a control plane
	// e.g. spec.pilot.enabled || spec.mixer.enabled || spec.galley.enabled || spec.sidecarInjectorWebhook.enabled || ....
	// which is all we're supporting atm.  if the scope expands to allow
	// installing custom gateways, etc., we should revisit this.
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: r.instance.Namespace}}
	err = r.client.Get(context.TODO(), client.ObjectKey{Name: r.instance.Namespace}, namespace)
	if err == nil {
		if namespace.Labels == nil {
			namespace.Labels = map[string]string{}
		}
		if label, ok := namespace.Labels["istio.openshift.com/ignore-namespace"]; !ok || label != "ignore" {
			r.log.V(1).Info("Adding istio.openshift.com/ignore-namespace=ignore label to Request.Namespace")
			namespace.Labels["istio.openshift.com/ignore-namespace"] = "ignore"
			err = r.client.Update(context.TODO(), namespace)
		}
	} else {
		allErrors = append(allErrors, err)
	}

	// ensure crds - move to bootstrap at operator startup

	// wait for crd availability - we should block bootstrapping until the crds are available

	// create components
	owner := metav1.NewControllerRef(r.instance, istiov1alpha3.SchemeGroupVersion.WithKind("ControlPlane"))
	r.ownerRefs = []metav1.OwnerReference{*owner}

	componentsProcessed := map[string]struct{}{}

	// this ordering is based on the 1.0 resource ordering

	// create core istio resources
	componentsProcessed["istio"] = seen
	err = r.processComponentManifests("istio")
	if err != nil {
		allErrors = append(allErrors, err)
	}

	// create security
	componentsProcessed["istio/charts/security"] = seen
	err = r.processComponentManifests("istio/charts/security")
	if err != nil {
		allErrors = append(allErrors, err)
	}

	// create galley
	componentsProcessed["istio/charts/galley"] = seen
	err = r.processComponentManifests("istio/charts/galley")
	if err != nil {
		allErrors = append(allErrors, err)
	}

	// prometheus
	componentsProcessed["istio/charts/prometheus"] = seen
	err = r.processComponentManifests("istio/charts/prometheus")
	if err != nil {
		allErrors = append(allErrors, err)
	}

	// create mixer
	componentsProcessed["istio/charts/mixer"] = seen
	err = r.processComponentManifests("istio/charts/mixer")
	if err != nil {
		allErrors = append(allErrors, err)
	}

	// create pilot
	componentsProcessed["istio/charts/pilot"] = seen
	err = r.processComponentManifests("istio/charts/pilot")
	if err != nil {
		allErrors = append(allErrors, err)
	}

	// gateways
	componentsProcessed["istio/charts/gateways"] = seen
	err = r.processComponentManifests("istio/charts/gateways")
	if err != nil {
		allErrors = append(allErrors, err)
	}

	// sidecar injector
	componentsProcessed["istio/charts/sidecarInjectorWebhook"] = seen
	err = r.processComponentManifests("istio/charts/sidecarInjectorWebhook")
	if err != nil {
		allErrors = append(allErrors, err)
	}

	// install grafana
	componentsProcessed["istio/charts/grafana"] = seen
	err = r.processComponentManifests("istio/charts/grafana")
	if err != nil {
		allErrors = append(allErrors, err)
	}

	// install jaeger
	componentsProcessed["istio/charts/tracing"] = seen
	err = r.processComponentManifests("istio/charts/tracing")
	if err != nil {
		allErrors = append(allErrors, err)
	}

	// install kiali
	componentsProcessed["istio/charts/kiali"] = seen
	err = r.processComponentManifests("istio/charts/kiali")
	if err != nil {
		allErrors = append(allErrors, err)
	}

	// other components
	for key := range r.renderings {
		if !strings.HasPrefix(key, "istio/") {
			continue
		}
		if _, ok := componentsProcessed[key]; ok {
			// already processed this component
			continue
		}
		componentsProcessed[key] = seen
		err = r.processComponentManifests(key)
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}

	// install launcher
	componentsProcessed["maistra-launcher"] = seen
	err = r.processComponentManifests("maistra-launcher")
	if err != nil {
		allErrors = append(allErrors, err)
	}

	// install 3scale
	componentsProcessed["maistra-launcher"] = seen
	err = r.processComponentManifests("maistra-threescale")
	if err != nil {
		allErrors = append(allErrors, err)
	}

	// delete unseen components
	for key := range r.instance.Status.ComponentStatus {
		if _, ok := componentsProcessed[key]; ok {
			continue
		}
		componentsProcessed[key] = seen
		err = r.processComponentManifests(key)
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}

	// XXX: maybe we add additional charts to handle these resources
	// set authenticated cluster role to enable access for all users

	// create route for ingress gateway service

	// create route for prometheus service

	r.instance.Status.ObservedGeneration = r.instance.GetGeneration()

	err = utilerrors.NewAggregate(allErrors)
	updateReconcileStatus(&r.instance.Status.StatusType, err)

	updateErr := r.client.Status().Update(context.TODO(), r.instance)
	if updateErr != nil {
		r.log.Error(err, "error updating ControlPlane status")
		if err == nil {
			// XXX: is this the right thing to do?
			return reconcile.Result{}, updateErr
		}
	}
	return reconcile.Result{}, err
}

func isEnabled(spec map[string]interface{}) bool {
	if enabledVal, ok := spec["enabled"]; ok {
		if enabled, ok := enabledVal.(bool); ok {
			return enabled
		}
	}
	return false
}

func (r *controlPlaneReconciler) processComponentManifests(componentName string) error {
	var err error
	status, hasStatus := r.instance.Status.ComponentStatus[componentName]
	renderings, hasRenderings := r.renderings[componentName]
	origLogger := r.log
	r.log = r.log.WithValues("Component", componentName)
	defer func() { r.log = origLogger }()
	if hasRenderings {
		if !hasStatus {
			status = istiov1alpha3.NewComponentStatus()
			r.instance.Status.ComponentStatus[componentName] = status
		}
		r.log.Info("reconciling component resources")
		status.RemoveCondition(istiov1alpha3.ConditionTypeReconciled)
		err := r.processManifests(renderings, status)
		updateReconcileStatus(&status.StatusType, err)
		status.ObservedGeneration = r.instance.GetGeneration()
		r.processNewComponent(componentName, status)
	} else if hasStatus && status.GetCondition(istiov1alpha3.ConditionTypeInstalled).Status != istiov1alpha3.ConditionStatusFalse {
		// delete resources
		r.log.Info("deleting component resources")
		err := r.processManifests([]manifest.Manifest{}, status)
		updateDeleteStatus(&status.StatusType, err)
		status.ObservedGeneration = r.instance.GetGeneration()
		r.processDeletedComponent(componentName, status)
	} else {
		r.log.Info("no renderings for component")
	}
	return err
}

func (r *controlPlaneReconciler) processManifests(manifests []manifest.Manifest,
	componentStatus *istiov1alpha3.ComponentStatus) error {

	allErrors := []error{}
	resourcesProcessed := map[istiov1alpha3.ResourceKey]struct{}{}

	origLogger := r.log
	defer func() { r.log = origLogger }()
	for _, manifest := range manifests {
		r.log = origLogger.WithValues("manifest", manifest.Name)
		if !strings.HasSuffix(manifest.Name, ".yaml") {
			r.log.V(2).Info("Skipping rendering of manifest")
			continue
		}
		r.log.Info("Processing resources from manifest")
		// split the manifest into individual objects
		objects := releaseutil.SplitManifests(manifest.Content)
		manifestLogger := r.log
		for _, raw := range objects {
			rawJSON, err := yaml.YAMLToJSON([]byte(raw))
			if err != nil {
				r.log.Error(err, "unable to convert raw data to JSON")
				allErrors = append(allErrors, err)
				continue
			}
			obj := &unstructured.Unstructured{}
			_, _, err = unstructured.UnstructuredJSONScheme.Decode(rawJSON, nil, obj)
			if err != nil {
				r.log.Error(err, "unable to decode object into Unstructured")
				allErrors = append(allErrors, err)
				continue
			}

			// Add owner ref
			obj.SetOwnerReferences(r.ownerRefs)

			key := istiov1alpha3.NewResourceKey(obj, obj)
			r.log = manifestLogger.WithValues("Resource", key)

			r.log.V(2).Info("beginning reconciliation of resource", "ResourceKey", key)

			resourcesProcessed[key] = seen
			status, ok := componentStatus.ResourceStatus[key]
			if !ok {
				newStatus := istiov1alpha3.NewStatus()
				status = &newStatus
				componentStatus.ResourceStatus[key] = status
			}

			status.RemoveCondition(istiov1alpha3.ConditionTypeReconciled)

			err = r.patchObject(obj)
			if err != nil {
				r.log.Error(err, "error patching object")
				updateReconcileStatus(status, err)
				continue
			}

			receiver := key.ToUnstructured()
			objectKey, err := client.ObjectKeyFromObject(receiver)
			if err != nil {
				r.log.Error(err, "client.ObjectKeyFromObject() failed for resource")
				r.log.V(5).Info("raw: object", "object", raw)
				// This can only happen if reciever isn't an unstructured.Unstructured
				// i.e. this should never happen
				updateReconcileStatus(status, err)
				allErrors = append(allErrors, err)
				continue
			}
			err = r.client.Get(context.TODO(), objectKey, receiver)
			if err != nil {
				if errors.IsNotFound(err) {
					r.log.Info("creating resource")
					err = r.client.Create(context.TODO(), obj)
					if err == nil {
						status.ObservedGeneration = obj.GetGeneration()
						// special handling
						r.processNewObject(obj)
					}
				}
			} else if (receiver.GetGeneration() > status.ObservedGeneration) || // somebody changed the object out from under us
				!(reflect.DeepEqual(obj.GetAnnotations(), receiver.GetAnnotations()) &&
					reflect.DeepEqual(obj.GetLabels(), receiver.GetLabels())) ||
				shouldUpdate(obj.UnstructuredContent(), receiver.UnstructuredContent()) {
				r.log.Info("updating existing resource")
				//r.log.Info("updates not supported at this time")
				// XXX: k8s barfs on some updates: metadata.resourceVersion: Invalid value: 0x0: must be specified for an update
				obj.SetResourceVersion(receiver.GetResourceVersion())
				err = r.client.Update(context.TODO(), obj)
				if err == nil {
					status.ObservedGeneration = obj.GetGeneration()
				}
			}
			r.log.V(2).Info("resource reconciliation complete")
			updateReconcileStatus(status, err)
			if err != nil {
				r.log.Error(err, "error occurred reconciling resource")
				allErrors = append(allErrors, err)
			}
		}
	}

	// handle deletions
	// XXX: should these be processed in reverse order of creation?
	for key, status := range componentStatus.ResourceStatus {
		r.log = origLogger.WithValues("Resource", key)
		if _, ok := resourcesProcessed[key]; !ok {
			if condition := status.GetCondition(istiov1alpha3.ConditionTypeInstalled); condition.Status != istiov1alpha3.ConditionStatusFalse {
				r.log.Info("deleting resource")
				unstructured := key.ToUnstructured()
				err := r.client.Delete(context.TODO(), unstructured, client.PropagationPolicy(metav1.DeletePropagationBackground))
				updateDeleteStatus(status, err)
				if err == nil || errors.IsNotFound(err) {
					status.ObservedGeneration = 0
					// special handling
					r.processDeletedObject(unstructured)
				} else {
					r.log.Error(err, "error deleting resource")
					allErrors = append(allErrors, err)
				}
			}
		}
	}
	return utilerrors.NewAggregate(allErrors)
}

func updateReconcileStatus(status *istiov1alpha3.StatusType, err error) {
	installStatus := status.GetCondition(istiov1alpha3.ConditionTypeInstalled).Status
	if err == nil {
		if installStatus != istiov1alpha3.ConditionStatusTrue {
			status.SetCondition(istiov1alpha3.Condition{
				Type:   istiov1alpha3.ConditionTypeInstalled,
				Reason: istiov1alpha3.ConditionReasonInstallSuccessful,
				Status: istiov1alpha3.ConditionStatusTrue,
			})
			status.SetCondition(istiov1alpha3.Condition{
				Type:   istiov1alpha3.ConditionTypeReconciled,
				Reason: istiov1alpha3.ConditionReasonInstallSuccessful,
				Status: istiov1alpha3.ConditionStatusTrue,
			})
		} else {
			status.SetCondition(istiov1alpha3.Condition{
				Type:   istiov1alpha3.ConditionTypeReconciled,
				Reason: istiov1alpha3.ConditionReasonReconcileSuccessful,
				Status: istiov1alpha3.ConditionStatusTrue,
			})
		}
	} else if installStatus == istiov1alpha3.ConditionStatusUnknown {
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
	} else {
		status.SetCondition(istiov1alpha3.Condition{
			Type:    istiov1alpha3.ConditionTypeReconciled,
			Reason:  istiov1alpha3.ConditionReasonReconcileError,
			Status:  istiov1alpha3.ConditionStatusFalse,
			Message: err.Error(),
		})
	}
}

func updateDeleteStatus(status *istiov1alpha3.StatusType, err error) {
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
	} else {
		status.SetCondition(istiov1alpha3.Condition{
			Type:    istiov1alpha3.ConditionTypeReconciled,
			Status:  istiov1alpha3.ConditionStatusFalse,
			Reason:  istiov1alpha3.ConditionReasonDeletionError,
			Message: err.Error(),
		})
	}
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
func (r *controlPlaneReconciler) processNewComponent(name string, status *istiov1alpha3.ComponentStatus) error {
	switch name {
	case "istio/charts/galley":
		r.waitForDeployments(status)
		if name == "istio/charts/galley" {
			for webhookKey, status := range status.FindResourcesOfKind("ValidatingWebhookConfiguration") {
				if installCondition := status.GetCondition(istiov1alpha3.ConditionTypeInstalled); installCondition.Status == istiov1alpha3.ConditionStatusTrue {
					r.waitForWebhookCABundleInitialization(webhookKey.ToUnstructured())
				}
			}
		}
	case "istio/charts/sidecarInjectorWebhook":
		for webhookKey, status := range status.FindResourcesOfKind("MutatingWebhookConfiguration") {
			if installCondition := status.GetCondition(istiov1alpha3.ConditionTypeInstalled); installCondition.Status == istiov1alpha3.ConditionStatusTrue {
				r.waitForWebhookCABundleInitialization(webhookKey.ToUnstructured())
			}
		}
		r.waitForDeployments(status)
	case "istio/charts/tracing":
		// XXX: not sure if we need to have elasticsearch ready before jaeger, nor how to accomplish that
		r.waitForDeployments(status)
	default:
		r.waitForDeployments(status)
	}
	return nil
}

func (r *controlPlaneReconciler) processDeletedComponent(name string, status *istiov1alpha3.ComponentStatus) error {
	return nil
}

func (r *controlPlaneReconciler) patchObject(object *unstructured.Unstructured) error {
	gk := object.GroupVersionKind().GroupKind()
	switch gk.String() {
	case "ConfigMap":
		if object.GetName() == "kiali" {
			return r.patchKialiConfig(object)
		}
	}
	return nil
}

func (r *controlPlaneReconciler) processNewObject(object *unstructured.Unstructured) error {
	gk := object.GroupVersionKind().GroupKind()
	switch gk.String() {
	case "ServiceAccount":
		return r.processNewServiceAccount(object)
	}
	return nil
}

func (r *controlPlaneReconciler) processDeletedObject(object *unstructured.Unstructured) error {
	gk := object.GroupVersionKind().GroupKind()
	switch gk.String() {
	case "ServiceAccount":
		return r.processDeletedServiceAccount(object)
	}
	return nil
}

var (
	grafanaRegexp = regexp.MustCompile("(grafana:\\s*url:).*?\n")
	jaegerRegexp  = regexp.MustCompile("(jaeger:\\s*url:).*?\n")
)

func (r *controlPlaneReconciler) patchKialiConfig(object *unstructured.Unstructured) error {
	configYaml, found, err := unstructured.NestedString(object.UnstructuredContent(), "data", "config.yaml")
	if err != nil {
		// This shouldn't occur if it's really a ConfigMap, but...
		r.log.Error(err, "could not parse kiali ConfigMap")
		return err
	} else if !found {
		return nil
	}

	// get jaeger route host
	jaegerRoute := &unstructured.Unstructured{}
	jaegerRoute.SetAPIVersion("route.openshift.io/v1")
	jaegerRoute.SetKind("Route")
	err = r.client.Get(context.TODO(), client.ObjectKey{Name: "jaeger-query", Namespace: object.GetNamespace()}, jaegerRoute)
	if err != nil && !errors.IsNotFound(err) {
		r.log.Error(err, "error retrieving jaeger route")
		return fmt.Errorf("could not retrieve jaeger route: %s", err)
	}

	// get grafana route host
	grafanaRoute := &unstructured.Unstructured{}
	grafanaRoute.SetAPIVersion("route.openshift.io/v1")
	grafanaRoute.SetKind("Route")
	err = r.client.Get(context.TODO(), client.ObjectKey{Name: "grafana", Namespace: object.GetNamespace()}, grafanaRoute)
	if err != nil && !errors.IsNotFound(err) {
		r.log.Error(err, "error retrieving grafana route")
		return fmt.Errorf("could not retrieve grafana route: %s", err)
	}

	// update config.yaml.external_services.grafana.url
	grafanaURL, _, _ := unstructured.NestedString(grafanaRoute.UnstructuredContent(), "spec", "host")
	configYaml = string(grafanaRegexp.ReplaceAll([]byte(configYaml), []byte("${1} "+grafanaURL)))
	// update config.yaml.external_services.jaeger.url
	jaegerURL, _, _ := unstructured.NestedString(jaegerRoute.UnstructuredContent(), "spec", "host")
	configYaml = string(jaegerRegexp.ReplaceAll([]byte(configYaml), []byte("${1} "+jaegerURL)))

	return unstructured.SetNestedField(object.UnstructuredContent(), configYaml, "data", "config.yaml")
}

// add-scc-to-user anyuid to service accounts: citadel, egressgateway, galley, ingressgateway, mixer, pilot, sidecar-injector
// plus: grafana, prometheus

// add-scc-to-user privileged service accounts: jaeger
func (r *controlPlaneReconciler) processNewServiceAccount(object *unstructured.Unstructured) error {
	switch object.GetName() {
	case "istio-ingressgateway-service-account",
		"istio-egressgateway-service-account",
		"istio-pilot-service-account",
		"istio-mixer-service-account",
		"istio-mixer-post-install-account",
		"istio-ca-service-account",
		"istio-sidecar-injector-service-account",
		"istio-citadel-service-account",
		"istio-ingress-service-account",
		"istio-galley-service-account",
		"istio-cleanup-old-ca-service-account",
		"prometheus",
		"default":
		return r.addUserToSCC("anyuid", serviceaccount.MakeUsername(object.GetNamespace(), object.GetName()))
	case "jaeger":
		return r.addUserToSCC("privileged", serviceaccount.MakeUsername(object.GetNamespace(), object.GetName()))
	}
	return nil
}

func (r *controlPlaneReconciler) addUserToSCC(sccName, user string) error {
	scc := &unstructured.Unstructured{}
	scc.SetAPIVersion("v1")
	scc.SetKind("SecurityContextConstraints")
	err := r.client.Get(context.TODO(), client.ObjectKey{Name: sccName}, scc)

	if err == nil {
		users, exists, _ := unstructured.NestedStringSlice(scc.UnstructuredContent(), "users")
		if !exists {
			users = []string{}
		}
		if indexOf(users, user) < 0 {
			r.log.Info("Adding ServiceAccount to SecurityContextConstraints", "ServiceAccount", user, "SecurityContextConstraints", sccName)
			users = append(users, user)
			unstructured.SetNestedStringSlice(scc.UnstructuredContent(), users, "users")
			err = r.client.Update(context.TODO(), scc)
		}
	}
	return err
}

func (r *controlPlaneReconciler) processDeletedServiceAccount(object *unstructured.Unstructured) error {
	switch object.GetName() {
	case "istio-ingressgateway-service-account",
		"istio-egressgateway-service-account",
		"istio-pilot-service-account",
		"istio-mixer-service-account",
		"istio-mixer-post-install-account",
		"istio-ca-service-account",
		"istio-sidecar-injector-service-account",
		"istio-citadel-service-account",
		"istio-ingress-service-account",
		"istio-galley-service-account",
		"istio-cleanup-old-ca-service-account",
		"prometheus",
		"default":
		return r.removeUserFromSCC("anyuid", serviceaccount.MakeUsername(object.GetNamespace(), object.GetName()))
	case "jaeger":
		return r.removeUserFromSCC("privileged", serviceaccount.MakeUsername(object.GetNamespace(), object.GetName()))
	}
	return nil
}

func (r *controlPlaneReconciler) removeUserFromSCC(sccName, user string) error {
	scc := &unstructured.Unstructured{}
	scc.SetAPIVersion("v1")
	scc.SetKind("SecurityContextConstraints")
	err := r.client.Get(context.TODO(), client.ObjectKey{Name: sccName}, scc)

	if err == nil {
		users, exists, _ := unstructured.NestedStringSlice(scc.UnstructuredContent(), "users")
		if !exists {
			return nil
		}
		if index := indexOf(users, user); index >= 0 {
			r.log.Info("Removing ServiceAccount from SecurityContextConstraints", "ServiceAccount", user, "SecurityContextConstraints", sccName)
			users = append(users[:index], users[index+1:]...)
			unstructured.SetNestedStringSlice(scc.UnstructuredContent(), users, "users")
			err = r.client.Update(context.TODO(), scc)
		}
	}
	return err
}

func (r *controlPlaneReconciler) waitForDeployments(status *istiov1alpha3.ComponentStatus) error {
	for deploymentKey, status := range status.FindResourcesOfKind("Deployment") {
		if installCondition := status.GetCondition(istiov1alpha3.ConditionTypeInstalled); installCondition.Status == istiov1alpha3.ConditionStatusTrue {
			r.waitForDeployment(deploymentKey.ToUnstructured())
		}
	}
	return nil
}

func (r *controlPlaneReconciler) waitForDeployment(object *unstructured.Unstructured) error {
	name := object.GetName()
	// wait for deployment replicas >= 1
	r.log.Info("waiting for deployment to become ready", "Deployment", name)
	for i := 0; i < 10; i++ {
		r.client.Get(context.TODO(), client.ObjectKey{Namespace: object.GetNamespace(), Name: name}, object)
		if val, _, _ := unstructured.NestedInt64(object.UnstructuredContent(), "status", "readyReplicas"); val > 0 {
			return nil
		}
		time.Sleep(6 * time.Second)
	}
	r.log.Error(nil, "deployment failed to become ready after 600s", "Deployment", name)
	return nil
}

func (r *controlPlaneReconciler) waitForWebhookCABundleInitialization(object *unstructured.Unstructured) error {
	name := object.GetName()
	kind := object.GetKind()
	r.log.Info("waiting for webhook CABundle initialization", kind, name)
outer:
	for i := 0; i < 10; i++ {
		r.client.Get(context.TODO(), client.ObjectKey{Name: name}, object)
		webhooks, found, _ := unstructured.NestedSlice(object.UnstructuredContent(), "webhooks")
		if !found || len(webhooks) == 0 {
			return nil
		}
		for _, webhook := range webhooks {
			typedWebhook, _ := webhook.(map[string]interface{})
			if caBundle, found, _ := unstructured.NestedString(typedWebhook, "clientConfig", "caBundle"); !found || len(caBundle) == 0 {
				time.Sleep(6 * time.Second)
				continue outer
			}
		}
		return nil
	}
	r.log.Error(nil, "webhook CABundle failed to become initialized after 600s", kind, name)
	return nil
}
