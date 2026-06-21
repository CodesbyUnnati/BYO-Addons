package controller

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	platformv1alpha1 "github.com/example/byo-addons/api/v1alpha1"
)

func TestDesiredStateName(t *testing.T) {
	tests := []struct {
		name          string
		addonSetName  string
		componentName string
		want          string
	}{
		{
			name:          "lowercase names",
			addonSetName:  "platform",
			componentName: "cilium",
			want:          "platform-cilium-desired",
		},
		{
			name:          "normalizes uppercase and underscores",
			addonSetName:  "Prod_Addons",
			componentName: "OpenEBS",
			want:          "prod-addons-openebs-desired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := desiredStateName(tt.addonSetName, tt.componentName); got != tt.want {
				t.Fatalf("desiredStateName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestComponentStatus(t *testing.T) {
	component := platformv1alpha1.AddonComponent{
		Name:     "cilium",
		Type:     platformv1alpha1.AddonTypeCNI,
		Provider: "cilium",
	}

	status := componentStatus(component, true, "desired state recorded", "ConfigMap/platform-cilium-desired")

	if status.Name != component.Name || status.Type != component.Type || status.Provider != component.Provider {
		t.Fatalf("componentStatus did not copy component identity: %#v", status)
	}
	if !status.Ready {
		t.Fatalf("componentStatus Ready = false, want true")
	}
	if status.Message != "desired state recorded" {
		t.Fatalf("componentStatus Message = %q", status.Message)
	}
	if status.Observed != "ConfigMap/platform-cilium-desired" {
		t.Fatalf("componentStatus Observed = %q", status.Observed)
	}
	if status.UpdatedAt.IsZero() {
		t.Fatalf("componentStatus UpdatedAt was not set")
	}
}

func TestUpsertCondition(t *testing.T) {
	oldTransition := metav1.NewTime(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC))
	newTransition := metav1.NewTime(time.Date(2026, 1, 1, 12, 5, 0, 0, time.UTC))

	existing := []metav1.Condition{
		{
			Type:               "Ready",
			Status:             metav1.ConditionFalse,
			Reason:             "ComponentsPending",
			Message:            "0 of 1 enabled components reconciled",
			LastTransitionTime: oldTransition,
		},
	}

	t.Run("adds a new condition type", func(t *testing.T) {
		conditions := upsertCondition(existing, metav1.Condition{
			Type:               "Healthy",
			Status:             metav1.ConditionTrue,
			Reason:             "ChecksPassed",
			LastTransitionTime: newTransition,
		})

		if len(conditions) != 2 {
			t.Fatalf("len(conditions) = %d, want 2", len(conditions))
		}
		if conditions[1].Type != "Healthy" {
			t.Fatalf("new condition type = %q, want Healthy", conditions[1].Type)
		}
	})

	t.Run("preserves transition time when status is unchanged", func(t *testing.T) {
		conditions := upsertCondition(existing, metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionFalse,
			Reason:             "StillPending",
			Message:            "still waiting",
			LastTransitionTime: newTransition,
		})

		if len(conditions) != 1 {
			t.Fatalf("len(conditions) = %d, want 1", len(conditions))
		}
		if !conditions[0].LastTransitionTime.Equal(&oldTransition) {
			t.Fatalf("LastTransitionTime = %s, want %s", conditions[0].LastTransitionTime, oldTransition)
		}
		if conditions[0].Reason != "StillPending" {
			t.Fatalf("Reason = %q, want StillPending", conditions[0].Reason)
		}
	})

	t.Run("uses new transition time when status changes", func(t *testing.T) {
		conditions := upsertCondition(existing, metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionTrue,
			Reason:             "ComponentsReconciled",
			Message:            "1 of 1 enabled components reconciled",
			LastTransitionTime: newTransition,
		})

		if !conditions[0].LastTransitionTime.Equal(&newTransition) {
			t.Fatalf("LastTransitionTime = %s, want %s", conditions[0].LastTransitionTime, newTransition)
		}
		if conditions[0].Status != metav1.ConditionTrue {
			t.Fatalf("Status = %s, want True", conditions[0].Status)
		}
	})
}
