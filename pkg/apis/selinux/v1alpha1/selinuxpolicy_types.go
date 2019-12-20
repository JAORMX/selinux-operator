package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SelinuxPolicySpec defines the desired state of SelinuxPolicy
type SelinuxPolicySpec struct {
	Apply  bool   `json:"apply,omitempty"`
	Policy string `json:"policy,omitempty"`
}

// SelinuxPolicyStatus defines the observed state of SelinuxPolicy
type SelinuxPolicyStatus struct {
	Installation string `json:"installation,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SelinuxPolicy is the Schema for the selinuxpolicies API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=selinuxpolicies,scope=Namespaced
type SelinuxPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SelinuxPolicySpec   `json:"spec,omitempty"`
	Status SelinuxPolicyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SelinuxPolicyList contains a list of SelinuxPolicy
type SelinuxPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SelinuxPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SelinuxPolicy{}, &SelinuxPolicyList{})
}
