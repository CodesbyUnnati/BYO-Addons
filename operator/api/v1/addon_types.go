package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AddonSpec contains the desired state for one addon.
type AddonSpec struct {
	// Name is the stable name of the addon.
	Name string `json:"name"`

	// Type describes the addon category.
	// +optional
	Type string `json:"type,omitempty"`

	// Priority controls ordering when multiple addons compete for a slot.
	// +optional
	Priority *int32 `json:"priority,omitempty"`
}

// Addon is the schema for the Addon API.
type Addon struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AddonSpec `json:"spec,omitempty"`
}

// AddonList contains a list of Addon resources.
type AddonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Addon `json:"items"`
}
