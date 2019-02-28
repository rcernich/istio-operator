package installation

import (
	"context"
	"reflect"

	istiov1alpha1 "github.com/maistra/istio-operator/pkg/apis/istio/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_installation")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Installation Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileInstallation{Handler: installationHandler, client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("installation-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Installation
	err = c.Watch(&source.Kind{Type: &istiov1alpha1.Installation{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Installation
	// err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
	// 	IsController: true,
	// 	OwnerType:    &istiov1alpha1.Installation{},
	// })
	// if err != nil {
	// 	return err
	// }

	return nil
}

var _ reconcile.Reconciler = &ReconcileInstallation{}

// ReconcileInstallation reconciles a Installation object
type ReconcileInstallation struct {
	*Handler
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

const (
	finalizer = "istio-operator:Installation"
)

// Reconcile reads that state of the cluster for a Installation object and makes changes based on the state read
// and what is in the Installation.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileInstallation) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Installation")

	// Fetch the Installation instance
	instance := &istiov1alpha1.Installation{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{Requeue: true}, err
	}

	if instance.Name != istioInstallerCRName {
		reqLogger.Info("Ignoring istio installer CR %v, please redeploy using the %v name", instance.Name, istioInstallerCRName)
		return reconcile.Result{}, nil
	}

	deleted := instance.GetDeletionTimestamp() != nil
	finalizers := instance.GetFinalizers()
	finalizerIndex := indexOf(finalizers, finalizer)
	if !deleted && finalizerIndex < 0 {
		reqLogger.V(1).Info("Adding finalizer", "finalizer", finalizer)
		finalizers = append(finalizers, finalizer)
		instance.SetFinalizers(finalizers)
		err = r.client.Update(context.TODO(), instance)
		return reconcile.Result{Requeue: true}, err
	}

	if deleted {
		if finalizerIndex < 0 {
			// already deleted ourselves
			return reconcile.Result{}, nil
		}

		reqLogger.Info("Removing the Istio installation")
		if err := r.ensureProjectAndServiceAccount(); err != nil {
			return reconcile.Result{}, err
		}
		removalJob := r.getRemovalJob(instance)
		r.deleteJob(removalJob)
		installerJob := r.getInstallerJob(instance)
		r.deleteJob(installerJob)
		items := r.newRemovalJobItems(instance)
		if err := r.createItems(items); err != nil {
			reqLogger.Error(err, "Failed to create the istio removal job")
			return reconcile.Result{Requeue: true}, err
		}
		finalizers = append(finalizers[:finalizerIndex], finalizers[finalizerIndex+1:]...)
		instance.SetFinalizers(finalizers)
		err = r.client.Update(context.TODO(), instance)
		return reconcile.Result{Requeue: true}, err
	}
	if instance.Status != nil && instance.Status.State != nil {
		if *instance.Status.State == istioInstalledState {
			if reflect.DeepEqual(instance.Spec, instance.Status.Spec) {
				reqLogger.V(2).Info("Ignoring installed state for %v %v", instance.Kind, instance.Name)
				return reconcile.Result{}, nil
			}
		} else {
			reqLogger.Info("Reinstalling istio for %v %v", instance.Kind, instance.Name)
		}
	} else {
		reqLogger.Info("Installing istio for %v %v", instance.Kind, instance.Name)
	}

	if err := r.ensureProjectAndServiceAccount(); err != nil {
		return reconcile.Result{}, err
	}

	installerJob := r.getInstallerJob(instance)
	r.deleteJob(installerJob)
	removalJob := r.getRemovalJob(instance)
	r.deleteJob(removalJob)
	items := r.newInstallerJobItems(instance)
	if err := r.createItems(items); err != nil {
		reqLogger.Error(err, "Failed to create the istio installer job")
		// XXX: do we need to do something in the result?
		return reconcile.Result{}, err
	}
	state := istioInstalledState
	if instance.Status == nil {
		instance.Status = &istiov1alpha1.InstallationStatus{
			State: &state,
			Spec:  instance.Spec.DeepCopy(),
		}
	} else {
		instance.Status.State = &state
		instance.Status.Spec = instance.Spec.DeepCopy()
	}
	if err := r.client.Update(context.TODO(), instance); err != nil {
		reqLogger.Error(err, "Failed to update the installation state in the resource")
		// XXX: do we need to do something in the result?
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func indexOf(l []string, s string) int {
	for i, elem := range l {
		if elem == s {
			return i
		}
	}
	return -1
}
