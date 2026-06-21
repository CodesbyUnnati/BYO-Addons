package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	platformv1alpha1 "github.com/example/byo-addons/api/v1alpha1"
)

var _ = Describe("AddonSetReconciler with a fake Kubernetes client", func() {
	var (
		ctx        context.Context
		scheme     *runtime.Scheme
		k8sClient  client.Client
		reconciler *AddonSetReconciler
		addonSet   *platformv1alpha1.AddonSet
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(platformv1alpha1.AddToScheme(scheme)).To(Succeed())

		addonSet = &platformv1alpha1.AddonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster-addons",
				Namespace: "platform-system",
				UID:       ktypes.UID("addonset-uid"),
			},
			Spec: platformv1alpha1.AddonSetSpec{
				ClusterName:  "interview-cluster",
				GitOpsEngine: "ArgoCD",
			},
		}

		k8sClient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(addonSet).
			Build()
		reconciler = &AddonSetReconciler{Client: k8sClient, Scheme: scheme}
	})

	It("creates a desired-state ConfigMap for an enabled component", func() {
		component := platformv1alpha1.AddonComponent{
			Name:     "cilium",
			Type:     platformv1alpha1.AddonTypeCNI,
			Provider: "cilium",
			Version:  "1.16.0",
			Enabled:  true,
		}

		observed, err := reconciler.reconcileComponentConfigMap(ctx, addonSet, component)

		Expect(err).NotTo(HaveOccurred())
		Expect(observed).To(Equal("ConfigMap/cluster-addons-cilium-desired in namespace platform-system"))

		var configMap corev1.ConfigMap
		key := ktypes.NamespacedName{Name: "cluster-addons-cilium-desired", Namespace: "platform-system"}
		Expect(k8sClient.Get(ctx, key, &configMap)).To(Succeed())
		Expect(configMap.Labels).To(HaveKeyWithValue(managedLabel, "byo-addons-operator"))
		Expect(configMap.Labels).To(HaveKeyWithValue(componentLabel, "cilium"))
		Expect(configMap.Data).To(HaveKeyWithValue("engine", "ArgoCD"))
		Expect(configMap.Data).To(HaveKeyWithValue("cluster", "interview-cluster"))
		Expect(configMap.Data["addon.json"]).To(MatchJSON(`{
			"name": "cilium",
			"type": "CNI",
			"provider": "cilium",
			"version": "1.16.0",
			"enabled": true,
			"source": {},
			"values": null
		}`))
		Expect(configMap.OwnerReferences).To(HaveLen(1))
		Expect(configMap.OwnerReferences[0].Name).To(Equal("cluster-addons"))
	})

	It("honors a component namespace override", func() {
		component := platformv1alpha1.AddonComponent{
			Name:      "openebs",
			Type:      platformv1alpha1.AddonTypeCSI,
			Provider:  "openebs",
			Namespace: "storage-system",
			Enabled:   true,
		}

		observed, err := reconciler.reconcileComponentConfigMap(ctx, addonSet, component)

		Expect(err).NotTo(HaveOccurred())
		Expect(observed).To(Equal("ConfigMap/cluster-addons-openebs-desired in namespace storage-system"))

		var configMap corev1.ConfigMap
		key := ktypes.NamespacedName{Name: "cluster-addons-openebs-desired", Namespace: "storage-system"}
		Expect(k8sClient.Get(ctx, key, &configMap)).To(Succeed())
		Expect(configMap.Data).To(HaveKeyWithValue("engine", "ArgoCD"))
		Expect(configMap.OwnerReferences).To(BeEmpty())
	})

	It("deletes desired-state ConfigMaps during finalizer cleanup", func() {
		addonSet.Spec.Components = []platformv1alpha1.AddonComponent{
			{Name: "cilium", Type: platformv1alpha1.AddonTypeCNI, Provider: "cilium", Enabled: true},
			{Name: "openebs", Type: platformv1alpha1.AddonTypeCSI, Provider: "openebs", Namespace: "storage-system", Enabled: true},
		}

		for _, cm := range []corev1.ConfigMap{
			{ObjectMeta: metav1.ObjectMeta{Name: "cluster-addons-cilium-desired", Namespace: "platform-system"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "cluster-addons-openebs-desired", Namespace: "storage-system"}},
		} {
			configMap := cm
			Expect(k8sClient.Create(ctx, &configMap)).To(Succeed())
		}

		Expect(reconciler.deleteDesiredStateConfigMaps(ctx, addonSet)).To(Succeed())

		var configMap corev1.ConfigMap
		err := k8sClient.Get(ctx, ktypes.NamespacedName{Name: "cluster-addons-cilium-desired", Namespace: "platform-system"}, &configMap)
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
		err = k8sClient.Get(ctx, ktypes.NamespacedName{Name: "cluster-addons-openebs-desired", Namespace: "storage-system"}, &configMap)
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})
})
