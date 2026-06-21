package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestAddonSetRegistersWithScheme(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme returned error: %v", err)
	}

	obj, err := scheme.New(GroupVersion.WithKind("AddonSet"))
	if err != nil {
		t.Fatalf("scheme.New returned error: %v", err)
	}
	if _, ok := obj.(*AddonSet); !ok {
		t.Fatalf("scheme.New returned %T, want *AddonSet", obj)
	}
}

func TestAddonSetDeepCopyDoesNotShareSlices(t *testing.T) {
	original := &AddonSet{
		ObjectMeta: metav1.ObjectMeta{Name: "cluster-addons", Namespace: "platform"},
		Spec: AddonSetSpec{
			ClusterName: "interview-cluster",
			Components: []AddonComponent{
				{Name: "cilium", Type: AddonTypeCNI, Provider: "cilium"},
			},
		},
		Status: AddonSetStatus{
			Components: []AddonComponentStatus{
				{Name: "cilium", Type: AddonTypeCNI, Provider: "cilium", Ready: true},
			},
			Conditions: []metav1.Condition{
				{Type: "Ready", Status: metav1.ConditionTrue, Reason: "ComponentsReconciled"},
			},
		},
	}

	copy := original.DeepCopy()
	copy.Spec.Components[0].Name = "openebs"
	copy.Status.Components[0].Name = "openebs"
	copy.Status.Conditions[0].Status = metav1.ConditionFalse

	if original.Spec.Components[0].Name != "cilium" {
		t.Fatalf("DeepCopy shared spec components slice")
	}
	if original.Status.Components[0].Name != "cilium" {
		t.Fatalf("DeepCopy shared status components slice")
	}
	if original.Status.Conditions[0].Status != metav1.ConditionTrue {
		t.Fatalf("DeepCopy shared status conditions slice")
	}
}
