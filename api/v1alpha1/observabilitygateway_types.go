package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ObservabilityGatewaySpec defines the desired state of ObservabilityGateway
type ObservabilityGatewaySpec struct {
	Image           string         `json:"image"`
	ConfigConfigMap string         `json:"configConfigMap"`
	ExtraArgs       []string       `json:"extraArgs,omitempty"`
	Classes         []GatewayClass `json:"classes"`
}

type GatewayClass struct {
	Name           string                       `json:"name"`
	Replicas       int32                        `json:"replicas"`
	Resources      *corev1.ResourceRequirements `json:"resources,omitempty"`
	NodeSelector   map[string]string            `json:"nodeSelector,omitempty"`
	Tolerations    []corev1.Toleration          `json:"tolerations,omitempty"`
	Affinity       *corev1.Affinity             `json:"affinity,omitempty"`
	ExtraArgs      []string                     `json:"extraArgs,omitempty"`
	Ports          []corev1.ContainerPort       `json:"ports,omitempty"`
	PodAnnotations map[string]string            `json:"podAnnotations,omitempty"`
}

// ObservabilityGatewayStatus defines the observed state of ObservabilityGateway.
type ObservabilityGatewayStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type ObservabilityGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ObservabilityGatewaySpec   `json:"spec,omitempty"`
	Status ObservabilityGatewayStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ObservabilityGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ObservabilityGateway `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ObservabilityGateway{}, &ObservabilityGatewayList{})
}
