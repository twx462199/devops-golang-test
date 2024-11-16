package controller

import (
	"context"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	appsv1 "my.com/devops-golang-test/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = Describe("MyStatefulSet Controller", func() {
	var (
		ctx                  context.Context
		k8sClient            client.Client
		scheme               *runtime.Scheme
		controllerReconciler *MyStatefulSetReconciler
	)

	BeforeEach(func() {
		ctx = context.Background()

		// 创建 Scheme
		scheme = runtime.NewScheme()
		Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
		Expect(appsv1.AddToScheme(scheme)).To(Succeed())

		// 创建 fake client
		k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()

		// 初始化控制器
		controllerReconciler = &MyStatefulSetReconciler{
			Client: k8sClient,
			Scheme: scheme,
		}
	})

	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		var typeNamespacedName types.NamespacedName

		BeforeEach(func() {
			typeNamespacedName = types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}

			// 创建一个 MyStatefulSet 资源
			mystatefulset := &appsv1.MyStatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: appsv1.MyStatefulSetSpec{
					Replicas: int32Ptr(3),
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels:      map[string]string{"app": resourceName, "test": "test"},
							Annotations: map[string]string{"test": "test"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "nginx",
									Image: "nginx:latest",
								},
							},
						},
					},
					VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
						{
							Spec: corev1.PersistentVolumeClaimSpec{
								AccessModes: []corev1.PersistentVolumeAccessMode{
									corev1.ReadWriteOnce, // 访问模式
								},
								Resources: corev1.VolumeResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceStorage: *resource.NewQuantity(int64(100), resource.DecimalSI),
									},
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, mystatefulset)).To(Succeed())
		})

		It("should create the expected number of Pods", func() {
			By("Reconciling the created resource")
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking if the expected number of Pods are created")
			Eventually(func() int {
				podList := &corev1.PodList{}
				err = k8sClient.List(ctx, podList, client.InNamespace("default"), client.MatchingLabels{"mystatefulset-name": resourceName})
				Expect(err).NotTo(HaveOccurred())
				return len(podList.Items)
			}, time.Second*5, time.Millisecond*500).Should(Equal(3))
		})

		It("should update Pods when the MyStatefulSet is updated, images", func() {
			By("Updating the MyStatefulSet resource")
			mystatefulset := &appsv1.MyStatefulSet{}
			err := k8sClient.Get(ctx, typeNamespacedName, mystatefulset)
			Expect(err).NotTo(HaveOccurred())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// 假设更新的是镜像版本
			mystatefulset.Spec.Template.Spec.Containers[0].Image = "nginx:1.19"
			Expect(k8sClient.Update(ctx, mystatefulset)).To(Succeed())

			By("Reconciling the updated resource")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// 验证Pod是否被更新
			Eventually(func() string {
				podList := &corev1.PodList{}
				err = k8sClient.List(ctx, podList, client.InNamespace("default"), client.MatchingLabels{"mystatefulset-name": resourceName})
				Expect(err).NotTo(HaveOccurred())
				if len(podList.Items) > 0 {
					return podList.Items[0].Spec.Containers[0].Image
				}
				return ""
			}, time.Second*5, time.Millisecond*500).Should(Equal("nginx:1.19"))
		})

		It("should update Pods when the MyStatefulSet is updated, labels", func() {
			By("Updating the MyStatefulSet resource")
			mystatefulset := &appsv1.MyStatefulSet{}
			err := k8sClient.Get(ctx, typeNamespacedName, mystatefulset)
			Expect(err).NotTo(HaveOccurred())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// 假设更新的是标签
			mystatefulset.Spec.Template.ObjectMeta.Labels = map[string]string{"app": resourceName, "test": "change"}
			Expect(k8sClient.Update(ctx, mystatefulset)).To(Succeed())

			By("Reconciling the updated resource")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// 验证Pod是否被更新
			Eventually(func() string {
				podList := &corev1.PodList{}
				err = k8sClient.List(ctx, podList, client.InNamespace("default"), client.MatchingLabels{"mystatefulset-name": resourceName})
				Expect(err).NotTo(HaveOccurred())
				if len(podList.Items) > 0 {
					return podList.Items[0].ObjectMeta.Labels["test"]
				}
				return ""
			}, time.Second*5, time.Millisecond*500).Should(Equal("change"))
		})

		It("should update Pods when the MyStatefulSet is updated, annotations", func() {
			By("Updating the MyStatefulSet resource")
			mystatefulset := &appsv1.MyStatefulSet{}
			err := k8sClient.Get(ctx, typeNamespacedName, mystatefulset)
			Expect(err).NotTo(HaveOccurred())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// 假设更新的是注解
			mystatefulset.Spec.Template.ObjectMeta.Annotations = map[string]string{"test": "change"}
			Expect(k8sClient.Update(ctx, mystatefulset)).To(Succeed())

			By("Reconciling the updated resource")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// 验证Pod是否被更新
			Eventually(func() string {
				podList := &corev1.PodList{}
				err = k8sClient.List(ctx, podList, client.InNamespace("default"), client.MatchingLabels{"mystatefulset-name": resourceName})
				Expect(err).NotTo(HaveOccurred())
				if len(podList.Items) > 0 {
					return podList.Items[0].ObjectMeta.Annotations["test"]
				}
				return ""
			}, time.Second*5, time.Millisecond*500).Should(Equal("change"))
		})

		It("should clean up Pods and PVCs when the MyStatefulSet is deleted", func() {
			By("Deleting the MyStatefulSet resource")
			mystatefulset := &appsv1.MyStatefulSet{}
			err := k8sClient.Get(ctx, typeNamespacedName, mystatefulset)
			Expect(err).NotTo(HaveOccurred())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			mystatefulset.Finalizers = []string{"test.finalizer"}
			Expect(k8sClient.Update(ctx, mystatefulset)).To(Succeed())

			By("Reconciling the deleted resource")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// 验证Pod和PVC是否被删除
			Eventually(func() int {
				podList := &corev1.PodList{}
				err = k8sClient.List(ctx, podList, client.InNamespace("default"), client.MatchingLabels{"mystatefulset-name": resourceName})
				Expect(err).NotTo(HaveOccurred())
				return len(podList.Items)
			}, time.Second*5, time.Millisecond*500).Should(Equal(0))

			Eventually(func() int {
				pvcList := &corev1.PersistentVolumeClaimList{}
				err = k8sClient.List(ctx, pvcList, client.InNamespace("default"), client.MatchingLabels{"app": resourceName})
				Expect(err).NotTo(HaveOccurred())
				return len(pvcList.Items)
			}, time.Second*5, time.Millisecond*500).Should(Equal(0))
		})

		It("other", func() {
			By("Not Found")
			mystatefulset := &appsv1.MyStatefulSet{}
			err := k8sClient.Get(ctx, typeNamespacedName, mystatefulset)
			Expect(err).NotTo(HaveOccurred())
			mystatefulset.Spec.Replicas = int32Ptr(1)
			Expect(k8sClient.Update(ctx, mystatefulset)).To(Succeed())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			mystatefulset.Spec.Replicas = nil
			Expect(k8sClient.Update(ctx, mystatefulset)).To(Succeed())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Eventually(err.Error()).Should(Equal("配置缺少期望的副本数量"))

			Expect(k8sClient.Delete(ctx, mystatefulset)).To(Succeed())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func int32Ptr(i int32) *int32 {
	return &i
}
