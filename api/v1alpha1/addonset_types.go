/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// AddonType describes the role an addon plays in the cluster platform.
// +kubebuilder:validation:Enum=CNI;CSI
type AddonType string

const (
	AddonTypeCNI AddonType = "CNI"
	AddonTypeCSI AddonType = "CSI"
)

// AddonSource describes where the reconciler or GitOps system can find the addon.
type AddonSource struct {
	// RepoURL is the Helm repository, OCI registry, or Git repository URL.
	// +optional
	RepoURL string `json:"repoURL,omitempty"`

	// Chart is the Helm chart name when the addon is Helm based.
	// +optional
	Chart string `json:"chart,omitempty"`

	// Path is the Kustomize path when the addon is Git based.
	// +optional
	Path string `json:"path,omitempty"`

	// TargetRevision is a chart version, Git tag, branch, or commit SHA.
	// +optional
	TargetRevision string `json:"targetRevision,omitempty"`
}

// AddonComponent declares a single desired cluster addon.
type AddonComponent struct {
	// Name is a stable component name, for example cilium or openebs.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Type is the cluster capability category for this addon.
	Type AddonType `json:"type"`

	// Provider is the upstream provider or project name.
	// +kubebuilder:validation:MinLength=1
	Provider string `json:"provider"`

	// Version is the desired addon version.
	// +optional
	Version string `json:"version,omitempty"`

	// Namespace is where the addon should run.
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Enabled controls whether the addon is reconciled.
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// Source tells GitOps/automation where to fetch the addon.
	// +optional
	Source AddonSource `json:"source,omitempty"`

	// Values stores provider-specific Helm values or Kustomize variables.
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Values runtime.RawExtension `json:"values,omitempty"`
}

// AddonSetSpec defines the desired state of AddonSet.
type AddonSetSpec struct {
	// ClusterName is the logical cluster name this addon set targets.
	// +kubebuilder:validation:MinLength=1
	ClusterName string `json:"clusterName"`

	// GitOpsEngine names the reconciler that should apply provider manifests.
	// +kubebuilder:validation:Enum=ArgoCD;Flux;None
	// +kubebuilder:default=ArgoCD
	GitOpsEngine string `json:"gitOpsEngine,omitempty"`

	// Components is the desired addon inventory.
	// +kubebuilder:validation:MinItems=1
	Components []AddonComponent `json:"components"`
}

// AddonComponentStatus reports reconciliation state for one component.
type AddonComponentStatus struct {
	Name      string      `json:"name"`
	Type      AddonType   `json:"type"`
	Provider  string      `json:"provider"`
	Ready     bool        `json:"ready"`
	Message   string      `json:"message,omitempty"`
	Observed  string      `json:"observed,omitempty"`
	UpdatedAt metav1.Time `json:"updatedAt,omitempty"`
}

// AddonSetStatus defines the observed state of AddonSet.
type AddonSetStatus struct {
	ObservedGeneration int64                  `json:"observedGeneration,omitempty"`
	ReadyComponents    int32                  `json:"readyComponents,omitempty"`
	DesiredComponents  int32                  `json:"desiredComponents,omitempty"`
	Components         []AddonComponentStatus `json:"components,omitempty"`
	Conditions         []metav1.Condition     `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=aset
// +kubebuilder:printcolumn:name="Cluster",type=string,JSONPath=`.spec.clusterName`
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=`.status.desiredComponents`
// +kubebuilder:printcolumn:name="Ready",type=integer,JSONPath=`.status.readyComponents`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// AddonSet is the Schema for the addonsets API.
type AddonSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AddonSetSpec   `json:"spec,omitempty"`
	Status AddonSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AddonSetList contains a list of AddonSet.
type AddonSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AddonSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AddonSet{}, &AddonSetList{})
}
