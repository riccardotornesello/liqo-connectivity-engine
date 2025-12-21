/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TunnelPolicy defines the policy for the tunnel
// +kubebuilder:validation:Enum=allow;deny
type TunnelPolicy string

const (
	// TunnelPolicyAllow allows traffic through the tunnel
	TunnelPolicyAllow TunnelPolicy = "allow"
	// TunnelPolicyDeny denies traffic through the tunnel
	TunnelPolicyDeny TunnelPolicy = "deny"
)

// ResourceGroup defines the filter for entities
// +kubebuilder:validation:Enum=vc-local;vc-remote;remote
type ResourceGroup string

const (
	// ResourceGroupVcLocal represents entities in the virtual cluster hosted locally
	ResourceGroupVcLocal ResourceGroup = "vc-local"
	// ResourceGroupVcRemote represents entities in the virtual cluster hosted remotely
	ResourceGroupVcRemote ResourceGroup = "vc-remote"
	// ResourceGroupRemote represents the remote pod CIDR
	ResourceGroupRemote ResourceGroup = "remote"
	// ResourceGroupOffloaded represents entities offloaded on the own cluster
	ResourceGroupOffloaded ResourceGroup = "offloaded"
)

// RuleAction defines the action for a rule
// +kubebuilder:validation:Enum=allow;deny
type RuleAction string

const (
	// RuleActionAllow allows traffic matching the rule
	RuleActionAllow RuleAction = "allow"
	// RuleActionDeny denies traffic matching the rule
	RuleActionDeny RuleAction = "deny"
)

// PeeringSecurityRule defines a single peering security rule
type PeeringSecurityRule struct {
	// The group initiating the traffic
	Src *ResourceGroup `json:"src,omitempty"`

	// The group receiving the traffic
	Dst *ResourceGroup `json:"dst,omitempty"`

	// The action to take for traffic matching the rule
	Action RuleAction `json:"action"`
}

// PeeringSecuritySpec defines the desired state of PeeringSecurity
type PeeringSecuritySpec struct {
	// tunnelPolicy defines the policy for the tunnel
	// +kubebuilder:default=allow
	TunnelPolicy TunnelPolicy `json:"tunnelPolicy"`

	// rules defines the list of peering security rules
	Rules []PeeringSecurityRule `json:"rules,omitempty"`
}

// PeeringSecurityStatus defines the observed state of PeeringSecurity.
type PeeringSecurityStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the PeeringSecurity resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PeeringSecurity is the Schema for the peeringsecurities API
type PeeringSecurity struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of PeeringSecurity
	// +required
	Spec PeeringSecuritySpec `json:"spec"`

	// status defines the observed state of PeeringSecurity
	// +optional
	Status PeeringSecurityStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// PeeringSecurityList contains a list of PeeringSecurity
type PeeringSecurityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []PeeringSecurity `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PeeringSecurity{}, &PeeringSecurityList{})
}
