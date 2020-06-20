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

// ComponentRuntimeConfig allows for partial customization of a component's
// runtime configuration (Deployment, PodTemplate, auto scaling, pod disruption, etc.)
type ComponentRuntimeConfig struct {
	Deployment DeploymentRuntimeConfig `json:"deployment,omitempty"`
	Pod        PodRuntimeConfig `json:"pod,omitempty"`
}

// DeploymentRuntimeConfig allow customization of a component's Deployment
// resource, including additional labels/annotations, replica count, autoscaling,
// rollout strategy, etc.
type DeploymentRuntimeConfig struct {
	// Metadata specifies additional labels and annotations to be applied to the deployment
	Metadata MetadataConfig `json:"metadata,omitempty"`
	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	// .Values.*.replicaCount
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	// The deployment strategy to use to replace existing pods with new ones.
	// +optional
	// +patchStrategy=retainKeys
	// .Values.*.rollingMaxSurge, rollingMaxUnavailable, etc.
	Strategy *appsv1.DeploymentStrategy `json:"strategy,omitempty" patchStrategy:"retainKeys" protobuf:"bytes,4,opt,name=strategy"`

	// The number of old ReplicaSets to retain to allow rollback.
	// This is a pointer to distinguish between explicit zero and not specified.
	// Defaults to 10.
	// +optional
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty" protobuf:"varint,6,opt,name=revisionHistoryLimit"`

	// Autoscaling specifies the configuration for a HorizontalPodAutoscaler
	// to be applied to this deployment.  Null indicates no auto scaling.
	// .Values.*.autoscale* fields
	AutoScaling *AutoScalerConfig `json:"autoScaling,omitempty"`

	// .Values.global.podDisruptionBudget.enabled, if not null
	// XXX: this is currently a global setting, not per component.  perhaps
	// this should only be available on the defaults?
	Disruption *PodDisruptionBudget `json:"disruption,omitempty"`
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
	// Metadata allows additional annotations/labels to be applied to the pod
	// .Values.*.podAnnotations
	// XXX: currently, additional lables are not supported
	Metadata MetadataConfig `json:"metadata,omitempty"`
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	// .Values.nodeSelector
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,7,rep,name=nodeSelector"`

	// If specified, the pod's scheduling constraints
	// +optional
	// .Values.podAffinityLabelSelector, podAntiAffinityLabelSelector, nodeSelector
	// XXX: this is more descriptive than what is currently exposed (i.e. only pod affinities and nodeSelector)
	Affinity *corev1.Affinity `json:"affinity,omitempty" protobuf:"bytes,18,opt,name=affinity"`

	// If specified, the pod will be dispatched by specified scheduler.
	// If not specified, the pod will be dispatched by default scheduler.
	// +optional
	// XXX: not currently supported
	SchedulerName string `json:"schedulerName,omitempty" protobuf:"bytes,19,opt,name=schedulerName"`

	// If specified, the pod's tolerations.
	// +optional
	// .Values.tolerations
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`

	// .Values.global.priorityClassName
	// XXX: currently, this is only a global setting.  maybe only allow setting in global runtime defaults?
	PriorityClassName string `json:"priorityClassName,omitempty" protobuf:"bytes,24,opt,name=priorityClassName"`

	// XXX: is it too cheesy to use 'default' name for defaults?  default would apply to all containers
	// .Values.*.resource, imagePullPolicy, etc.
	Containers map[string]ContainerConfig
}

// ContainerConfig to be applied to containers in a pod, in a deployment
type ContainerConfig struct {
	ImagePullPolicy  corev1.PullPolicy `json:"imagePullPolicy,omitempty" protobuf:"bytes,14,opt,name=imagePullPolicy,casttype=PullPolicy"`
	ImagePullSecrets []corev1.LocalObjectReference          `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,15,rep,name=imagePullSecrets"`
	Resources        map[string]corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

// PodDisruptionBudget details
// XXX: currently not configurable (i.e. no values.yaml equivalent)
type PodDisruptionBudget struct {
	MinAvailable   *intstr.IntOrString `json:"minAvailable,omitempty" protobuf:"bytes,1,opt,name=minAvailable"`
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty" protobuf:"bytes,3,opt,name=maxUnavailable"`
}

type DefaultRuntimeConfig struct {
	Metadata  MetadataConfig
	Container *ContainerConfig
}

type MetadataConfig struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type ComponentServiceConfig struct {
	Metadata MetadataConfig
	// .Values.prometheus.service.nodePort.port, ...enabled is true if not null
	NodePort *int32
	Ingress  *ComponentIngressConfig
}

type ComponentIngressConfig struct {
	Metadata    MetadataConfig
	Hosts       []string
	ContextPath string
	TLS         map[string]string // RawExtension?
}

type ComponentPersistenceConfig struct {
	StorageClassName string
	AccessModes      []corev1.PersistentVolumeAccessMode
	Capacity         corev1.ResourceList
}
