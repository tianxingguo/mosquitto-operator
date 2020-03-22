package v1alpha1

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DefaultMosquittoContainerRepository is the default docker repo for the mosquitto
	// container
	DefaultMoContainerRepository = "192.168.1.99/library/eclipse-mosquitto"

	// DefaultMoContainerVersion is the default tag used for for the mosquitto
	// container
	DefaultMoContainerVersion = "1.6.8"

	// DefaultMoContainerPolicy is the default container pull policy used
	DefaultMoContainerPolicy = "Always"

	// DefaultTerminationGracePeriod is the default time given before the
	// container is stopped. This gives clients time to disconnect from a
	// specific node gracefully.
	DefaultTerminationGracePeriod = 30

	VolumeReclaimPolicyRetain VolumeReclaimPolicy = "Retain"

	VolumeReclaimPolicyDelete VolumeReclaimPolicy = "Delete"
)

// PodPolicy defines the common pod configuration for Pods, including when used
// in deployments, stateful-sets, etc.
type PodPolicy struct {
	// Labels specifies the labels to attach to pods the operator creates for
	// the mosquitto cluster.
	Labels map[string]string `json:"labels,omitempty"`

	// NodeSelector specifies a map of key-value pairs. For the pod to be
	// eligible to run on a node, the node must have each of the indicated
	// key-value pairs as labels.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// The scheduling constraints on pods.
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// Resources is the resource requirements for the container.
	// This field cannot be updated once the cluster is created.
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// Tolerations specifies the pod's tolerations.
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`

	// List of environment variables to set in the container.
	// This field cannot be updated.
	Env []v1.EnvVar `json:"env,omitempty"`

	// Annotations specifies the annotations to attach to pods the operator
	// creates.
	Annotations map[string]string `json:"annotations,omitempty"`

	// SecurityContext specifies the security context for the entire pod
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`

	// TerminationGracePeriodSeconds is the amount of time that kubernetes will
	// give for a pod instance to shutdown normally.
	// The default value is 1800.
	TerminationGracePeriodSeconds int64 `json:"terminationGracePeriodSeconds"`
}

// ContainerImage defines the fields needed for a Docker repository image. The
// format here matches the predominant format used in Helm charts.
type ContainerImage struct {
	Repository string        `json:"repository"`
	Tag        string        `json:"tag"`
	PullPolicy v1.PullPolicy `json:"pullPolicy"`
}

type VolumeReclaimPolicy string

type Persistence struct {
	// VolumeReclaimPolicy is a mosquitto operator configuration. If it's set to Delete,
	// the corresponding PVCs will be deleted by the operator when mosquitto cluster is deleted.
	// The default value is Retain.
	VolumeReclaimPolicy VolumeReclaimPolicy `json:"reclaimPolicy,omitempty"`
	// PersistentVolumeClaimSpec is the spec to describe PVC for the container
	// This field is optional. If no PVC spec, stateful containers will use
	// emptyDir as volume.
	PersistentVolumeClaimSpec v1.PersistentVolumeClaimSpec `json:"spec,omitempty"`
}

// MosquittoConfig is the current configuration of each MosquittoConfig node, which
// sets these values in the configmap
type MosquittoConfig struct {
	// InitLimit is the amount of time, in ticks, to allow followers to connect
	// and sync to a leader.
	// Default value is 10.
	InitLimit int `json:"initLimit"`

	// TickTime is the length of a single tick, which is the basic time unit used
	// by mosquitto, as measured in milliseconds
	// The default value is 2000.
	TickTime int `json:"tickTime"`

	// SyncLimit is the amount of time, in ticks, to allow followers to sync with
	// mosquitto.
	// The default value is 2.
	SyncLimit int `json:"syncLimit"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// MosquittoClusterSpec defines the desired state of MosquittoCluster
type MosquittoClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Image is the  container image. default is mosquitto:latest
	Image ContainerImage `json:"image,omitempty"`

	// Labels specifies the labels to attach to pods the operator creates for
	// the mosquitto cluster.
	Labels map[string]string `json:"labels,omitempty"`

	// Replicas is the expected size of the mosquitto cluster.
	// The pravega-operator will eventually make the size of the running cluster
	// equal to the expected size.
	// The valid range of size is from 1 to 7.
	Replicas int32 `json:"replicas"`

	Ports []v1.ContainerPort `json:"ports,omitempty"`

	// Pod defines the policy to create pod for the mosquitto cluster.
	// Updating the Pod does not take effect on any existing pods.
	Pod PodPolicy `json:"pod,omitempty"`

	// Persistence is the configuration for mosquitto persistent layer.
	// PersistentVolumeClaimSpec and VolumeReclaimPolicy can be specified in here.
	Persistence *Persistence `json:"persistence,omitempty"`

	// Conf is the mosquitto configuration, which will be used to generate the
	// static mosquitto configuration. If no configuration is provided required
	// default values will be provided, and optional values will be excluded.
	Conf MosquittoConfig `json:"config,omitempty"`
}

// MosquittoClusterStatus defines the observed state of MosquittoCluster
type MosquittoClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// Members is the mosquitto members in the cluster
	Members MembersStatus `json:"members"`

	// Replicas is the number of number of desired replicas in the cluster
	Replicas int32 `json:"replicas"`

	// ReadyReplicas is the number of number of ready replicas in the cluster
	ReadyReplicas int32 `json:"readyReplicas"`

	// InternalClientEndpoint is the internal client IP and port
	InternalClientEndpoint string `json:"internalClientEndpoint"`

	// ExternalClientEndpoint is the internal client IP and port
	ExternalClientEndpoint string `json:"externalClientEndpoint"`
}

// MembersStatus is the status of the members of the cluster with both
// ready and unready node membership lists
type MembersStatus struct {
	Ready   []string `json:"ready"`
	Unready []string `json:"unready"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MosquittoCluster is the Schema for the mosquittoclusters API
// +k8s:openapi-gen=true
type MosquittoCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MosquittoClusterSpec   `json:"spec,omitempty"`
	Status MosquittoClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MosquittoClusterList contains a list of MosquittoCluster
type MosquittoClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MosquittoCluster `json:"items"`
}

// Ports groups the ports for a mosquitto cluster node for easy access
type Ports struct {
	Client    int32
	Websocket int32
	Monitor    int32
}

// setDefaults set default values when not defined in the spec.
func (z *MosquittoCluster) SetDefaults() bool {
	return z.Spec.setDefaults(z)
}

// ConfigMapName returns the name of the cluster config-map
func (z *MosquittoCluster) ConfigMapName() string {
	return fmt.Sprintf("%s-configmap", z.GetName())
}

//set default elements for image
func (c *ContainerImage) setDefaults() (changed bool) {
	if c.Repository == "" {
		changed = true
		c.Repository = DefaultMoContainerRepository
	}
	if c.Tag == "" {
		changed = true
		c.Tag = DefaultMoContainerVersion
	}
	if c.PullPolicy == "" {
		changed = true
		c.PullPolicy = DefaultMoContainerPolicy
	}
	return changed
}

// GetClientServiceName returns the name of the client service for the cluster
func (z *MosquittoCluster) GetClientServiceName() string {
	return fmt.Sprintf("%s-client", z.GetName())
}

// ToString formats a container image struct as a docker compatible repository
func (c *ContainerImage) ToString() string {
	return fmt.Sprintf("%s:%s", c.Repository, c.Tag)
}

// mosquittoPorts returns a struct of ports
func (z *MosquittoCluster) MosquittoPorts() Ports {
	ports := Ports{}
	for _, p := range z.Spec.Ports {
		if p.Name == "client" {
			ports.Client = p.ContainerPort
		} else if p.Name == "websocket" {
			ports.Websocket = p.ContainerPort
		} else if p.Name == "monitor" {
			ports.Monitor = p.ContainerPort
		}
	}

	return ports
}

func (s *MosquittoClusterSpec) setDefaults(z *MosquittoCluster) (changed bool) {
	changed = s.Image.setDefaults()
	if s.Conf.setDefaults() {
		changed = true
	}
	if s.Replicas == 0 {
		s.Replicas = 3
		changed = true
	}
	if s.Ports == nil {
		s.Ports = []v1.ContainerPort{{
			Name:          "client",
			ContainerPort: 1883,
		}, {
			Name:          "websocket",
			ContainerPort: 9001,
		}, {
			Name:          "monitor",
			ContainerPort: 9888,
		},
		}
		changed = true
	}
	if z.Spec.Labels == nil {
		z.Spec.Labels = map[string]string{}
		changed = true
	}
	if _, ok := z.Spec.Labels["app"]; !ok {
		z.Spec.Labels["app"] = z.GetName()
		changed = true
	}
	if _, ok := z.Spec.Labels["release"]; !ok {
		z.Spec.Labels["release"] = z.GetName()
		changed = true
	}
	if s.Pod.setDefaults(z) {
		changed = true
	}
	if s.Persistence == nil {
		s.Persistence = &Persistence{}
		changed = true
	}
	if s.Persistence.setDefaults() {
		changed = true
	}
	return changed
}

func (p *PodPolicy) setDefaults(z *MosquittoCluster) (changed bool) {
	if p.Labels == nil {
		p.Labels = map[string]string{}
		changed = true
	}
	if p.TerminationGracePeriodSeconds == 0 {
		p.TerminationGracePeriodSeconds = DefaultTerminationGracePeriod
		changed = true
	}
	if z.Spec.Pod.Labels == nil {
		p.Labels = map[string]string{}
		changed = true
	}
	if _, ok := p.Labels["app"]; !ok {
		p.Labels["app"] = z.GetName()
		changed = true
	}
	if _, ok := p.Labels["release"]; !ok {
		p.Labels["release"] = z.GetName()
		changed = true
	}
	if p.Affinity == nil {
		p.Affinity = &v1.Affinity{
			PodAntiAffinity: &v1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{{
					Weight: 20,
					PodAffinityTerm: v1.PodAffinityTerm{
						TopologyKey: "kubernetes.io/hostname",
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{{
								Key:      "app",
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{z.GetName()},
							},
							},
						},
					},
				},
				},
			},
		}
		changed = true
	}
	return changed
}

func (c *MosquittoConfig) setDefaults() (changed bool) {
	if c.InitLimit == 0 {
		changed = true
		c.InitLimit = 10
	}
	if c.TickTime == 0 {
		changed = true
		c.TickTime = 2000
	}
	if c.SyncLimit == 0 {
		changed = true
		c.SyncLimit = 2
	}
	return changed
}

func (p *Persistence) setDefaults() (changed bool) {
	if !p.VolumeReclaimPolicy.isValid() {
		changed = true
		p.VolumeReclaimPolicy = VolumeReclaimPolicyRetain
	}

	p.PersistentVolumeClaimSpec.AccessModes = []v1.PersistentVolumeAccessMode{
		v1.ReadWriteOnce,
	}

	if len(p.PersistentVolumeClaimSpec.Resources.Requests) == 0 {
		p.PersistentVolumeClaimSpec.Resources.Requests = v1.ResourceList{
			v1.ResourceStorage: resource.MustParse("20Gi"),
		}
		changed = true
	}
	return changed
}

func (v VolumeReclaimPolicy) isValid() bool {
	if v != VolumeReclaimPolicyDelete && v != VolumeReclaimPolicyRetain {
		return false
	}
	return true
}

func init() {
	SchemeBuilder.Register(&MosquittoCluster{}, &MosquittoClusterList{})
}
