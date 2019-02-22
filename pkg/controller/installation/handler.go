package installation

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"bytes"
	"context"

	"github.com/maistra/istio-operator/pkg/apis/istio/v1alpha1"

	batchv1	"k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	namespace = "istio-system"
	serviceAccountName = "openshift-ansible"

	inventoryDir = "/tmp/inventory/"
	inventoryFile = inventoryDir + "istio.inventory"

	playbookFile = "playbooks/openshift-istio/config.yml"
	playbookOptions = "-vvv"

	configurationDir = "/etc/origin/master"

	defaultIstioPrefix = "docker.io/maistra/"
	defaultIstioVersion = "0.1.0"
	defaultDeploymentType = "origin"

	newline = "\n"

	istioInstallerCRName = "istio-installation"

	istioInstalledState = "Istio Installer Job Created"
)

var (
	installationHandler *Handler
)

func RegisterHandler(h *Handler) {
	installationHandler = h
}

type Handler struct {
	// It is likely possible to determine these at runtime, we should investigate
	OpenShiftRelease string
	MasterPublicURL  string
	IstioPrefix      string
	IstioVersion     string
	DeploymentType   string
}

func (h *ReconcileInstallation) deleteJob(job *batchv1.Job) {
	objectKey, err := client.ObjectKeyFromObject(job)
	if err != nil {
		return
	}
	err = h.client.Get(context.TODO(), objectKey, job) ; if err == nil {
		uid := job.UID
		var parallelism int32 = 0
		job.Spec.Parallelism = &parallelism
		h.client.Update(context.TODO(), job)
		podList := corev1.PodList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
		}

		labelSelector := labels.SelectorFromSet(labels.Set(map[string]string{"controller-uid": string(uid)}))
		listOptions := &client.ListOptions{
			Raw: &metav1.ListOptions{
				LabelSelector:        labelSelector.String(),
				IncludeUninitialized: false,
			},
		}

		err := h.client.List(context.TODO(), listOptions, &podList) ; if err == nil {
			for _, pod := range podList.Items {
				h.client.Delete(context.TODO(), &pod)
			}
			orphanDependents := false
			h.client.Delete(context.TODO(), job, func(opts *client.DeleteOptions) { deleteOrphanedDependents(opts, &orphanDependents) })
		}
	}
}

func deleteOrphanedDependents(opts *client.DeleteOptions, orphanedDependents *bool) {
	raw := opts.Raw
	if raw == nil {
		raw = &metav1.DeleteOptions{}
		opts.Raw = raw
	}
	raw.OrphanDependents = orphanedDependents
}

func (h *ReconcileInstallation) deleteItem(object runtime.Object) {
	switch item := object.(type) {
	case *batchv1.Job:
		h.deleteJob(item)
	default:
		h.client.Delete(context.TODO(), item)
	}
}

func (h *ReconcileInstallation) deleteItems(items []runtime.Object) {
	lastItem := len(items)-1
	for i := range items {
		item:= items[lastItem-i]
		h.deleteItem(item)
	}
}

func (h *ReconcileInstallation) createItems(items []runtime.Object) error {
	for _, item := range items {
		if err := h.client.Create(context.TODO(), item); err != nil {
			h.deleteItems(items)
			return err
		}
	}
	return nil
}

func (h *Handler) getIstioImagePrefix(cr *v1alpha1.Installation) string {
	if cr.Spec != nil && cr.Spec.Istio != nil && cr.Spec.Istio.Prefix != nil {
		return *cr.Spec.Istio.Prefix
	} else if h.IstioPrefix != "" {
		return h.IstioPrefix
	} else {
		return defaultIstioPrefix
	}
}

func (h *Handler) getIstioImageVersion(cr *v1alpha1.Installation) string {
	if cr.Spec != nil && cr.Spec.Istio != nil && cr.Spec.Istio.Version != nil {
		return *cr.Spec.Istio.Version
	} else if h.IstioVersion != "" {
		return h.IstioVersion
	} else {
		return defaultIstioVersion
	}
}

func (h *Handler) getDeploymentType(cr *v1alpha1.Installation) string {
	if cr.Spec != nil && cr.Spec.DeploymentType != nil {
		return *cr.Spec.DeploymentType
	} else if h.DeploymentType != "" {
		return h.DeploymentType
	} else {
		return defaultDeploymentType
	}
}

func (h *Handler) getOpenShiftRelease() string {
	return h.OpenShiftRelease
}

func (h *Handler) getMasterPublicURL() *string {
	if h.MasterPublicURL == "" {
		return nil
	}
	return &h.MasterPublicURL
}

func addStringValue(b *bytes.Buffer, key string, value string) {
	b.WriteString(key)
	b.WriteString(value)
	b.WriteString(newline)
}

func addStringPtrValue(b *bytes.Buffer, key string, value *string) {
	if value != nil {
		addStringValue(b, key, *value)
	}
}

func addBooleanPtrValue(b *bytes.Buffer, key string, value *bool) {
	if value != nil {
		addBooleanValue(b, key, *value)
	}
}

func addBooleanValue(b *bytes.Buffer, key string, value bool) {
	if value {
		addStringValue(b, key, "True")
	} else {
		addStringValue(b, key, "False")
	}
}