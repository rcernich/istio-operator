package v2

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ControlPlaneRuntimeConfig struct {
	Citadel *ComponentRuntimeConfig
	Galley  *ComponentRuntimeConfig
	Pilot   *ComponentRuntimeConfig
	// Defaults will be merged into specific component config.
	Defaults *DefaultRuntimeConfig
}

type ComponentRuntimeConfig struct {
	Deployment DeploymentRuntimeConfig
	Pod        PodRuntimeConfig
	Containers map[string]ContainerConfig
}

type DeploymentRuntimeConfig struct {
	Labels map[string]string
	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	// The deployment strategy to use to replace existing pods with new ones.
	// +optional
	// +patchStrategy=retainKeys
	Strategy *appsv1.DeploymentStrategy `json:"strategy,omitempty" patchStrategy:"retainKeys" protobuf:"bytes,4,opt,name=strategy"`

	// The number of old ReplicaSets to retain to allow rollback.
	// This is a pointer to distinguish between explicit zero and not specified.
	// Defaults to 10.
	// +optional
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty" protobuf:"varint,6,opt,name=revisionHistoryLimit"`

	AutoScaling *AutoScalerConfig

	Disruption *PodDisruptionBudget
}

type AutoScalerConfig struct {
	// lower limit for the number of pods that can be set by the autoscaler, default 1.
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty" protobuf:"varint,2,opt,name=minReplicas"`
	// upper limit for the number of pods that can be set by the autoscaler; cannot be smaller than MinReplicas.
	MaxReplicas int32 `json:"maxReplicas" protobuf:"varint,3,opt,name=maxReplicas"`
	// target average CPU utilization (represented as a percentage of requested CPU) over all the pods;
	// if not specified the default autoscaling policy will be used.
	// +optional
	TargetCPUUtilizationPercentage *int32 `json:"targetCPUUtilizationPercentage,omitempty" protobuf:"varint,4,opt,name=targetCPUUtilizationPercentage"`
}

type PodRuntimeConfig struct {
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,7,rep,name=nodeSelector"`

	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty" protobuf:"bytes,18,opt,name=affinity"`

	// If specified, the pod will be dispatched by specified scheduler.
	// If not specified, the pod will be dispatched by default scheduler.
	// +optional
	SchedulerName string `json:"schedulerName,omitempty" protobuf:"bytes,19,opt,name=schedulerName"`

	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`

	PriorityClassName string `json:"priorityClassName,omitempty" protobuf:"bytes,24,opt,name=priorityClassName"`

	Annotations map[string]string
}

type ContainerConfig struct {
	ImagePullPolicy  corev1.PullPolicy
	ImagePullSecrets []corev1.LocalObjectReference          `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,15,rep,name=imagePullSecrets"`
	Resources        map[string]corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

type PodDisruptionBudget struct {
	MinAvailable   *intstr.IntOrString `json:"minAvailable,omitempty" protobuf:"bytes,1,opt,name=minAvailable"`
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty" protobuf:"bytes,3,opt,name=maxUnavailable"`
}

type DefaultRuntimeConfig struct {
	Labels      map[string]string
	Annotations map[string]string
	Container   *ContainerConfig
}
