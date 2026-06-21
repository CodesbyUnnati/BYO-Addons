//go:build envtest

package controller

import (
	"context"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	platformv1alpha1 "github.com/example/byo-addons/api/v1alpha1"
)

var _ = Describe("AddonSetReconciler with envtest", Ordered, func() {
	var (
		testEnv   *envtest.Environment
		k8sClient client.Client
		cancel    context.CancelFunc
	)

	BeforeAll(func() {
		testEnv = &envtest.Environment{
			CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
			ErrorIfCRDPathMissing: true,
		}

		cfg, err := testEnv.Start()
		Expect(err).NotTo(HaveOccurred())

		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(platformv1alpha1.AddToScheme(scheme)).To(Succeed())

		manager, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme,
			Metrics: metricsserver.Options{
				BindAddress: "0",
			},
		})
		Expect(err).NotTo(HaveOccurred())

		reconciler := &AddonSetReconciler{
			Client: manager.GetClient(),
			Scheme: manager.GetScheme(),
		}
		Expect(reconciler.SetupWithManager(manager)).To(Succeed())
		k8sClient = manager.GetClient()

		ctx, stop := context.WithCancel(context.Background())
		cancel = stop
		go func() {
			defer GinkgoRecover()
			Expect(manager.Start(ctx)).To(Succeed())
		}()
	})

	AfterAll(func() {
		if cancel != nil {
			cancel()
		}
		if testEnv != nil {
			Expect(testEnv.Stop()).To(Succeed())
		}
	})

	It("reconciles an AddonSet through the real API server", func() {
		ctx := context.Background()
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: "platform-system"},
		}
		Expect(k8sClient.Create(ctx, namespace)).To(Succeed())

		addonSet := &platformv1alpha1.AddonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster-addons",
				Namespace: "platform-system",
			},
			Spec: platformv1alpha1.AddonSetSpec{
				ClusterName:  "interview-cluster",
				GitOpsEngine: "ArgoCD",
				Components: []platformv1alpha1.AddonComponent{
					{Name: "cilium", Type: platformv1alpha1.AddonTypeCNI, Provider: "cilium", Enabled: true},
				},
			},
		}
		Expect(k8sClient.Create(ctx, addonSet)).To(Succeed())

		configMapKey := ktypes.NamespacedName{Name: "cluster-addons-cilium-desired", Namespace: "platform-system"}
		Eventually(func() error {
			var configMap corev1.ConfigMap
			return k8sClient.Get(ctx, configMapKey, &configMap)
		}, 10*time.Second, 250*time.Millisecond).Should(Succeed())

		Eventually(func() int32 {
			var current platformv1alpha1.AddonSet
			err := k8sClient.Get(ctx, ktypes.NamespacedName{Name: "cluster-addons", Namespace: "platform-system"}, &current)
			if err != nil {
				return -1
			}
			return current.Status.ReadyComponents
		}, 10*time.Second, 250*time.Millisecond).Should(Equal(int32(1)))
	})
})
