/*
Copyright 2024.

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

package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	appsv1 "my.com/devops-golang-test/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// MyStatefulSetReconciler reconciles a MyStatefulSet object
type MyStatefulSetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *MyStatefulSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// 获取当前的 MyStatefulSet 实例
	myStatefulSet, err := r.getMyStatefulSet(ctx, req)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("MyStatefulSet 资源未找到，可能已经被删除")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// 处理 MyStatefulSet 的删除逻辑
	if myStatefulSet.DeletionTimestamp != nil || myStatefulSet.Finalizers != nil {
		return r.cleanupMyStatefulSet(ctx, myStatefulSet, req)
	}

	// 获取期望的副本数量
	if myStatefulSet.Spec.Replicas == nil {
		err := errors.New("配置缺少期望的副本数量")
		logger.Error(err, "MyStatefulSet 配置错误")
		return ctrl.Result{}, err
	}
	desiredReplicas := *myStatefulSet.Spec.Replicas

	// 列出与 MyStatefulSet 关联的 Pod
	podList, err := r.listPods(ctx, req, myStatefulSet)
	if err != nil {
		return ctrl.Result{}, err
	}

	// 创建缺失的 Pod 和 PVC
	if err := r.createMissingPodsAndPVCs(ctx, req, myStatefulSet, podList, desiredReplicas); err != nil {
		return ctrl.Result{}, err
	}

	// 更新需要更新的 Pod
	if err := r.updatePods(ctx, myStatefulSet, podList, req); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *MyStatefulSetReconciler) getMyStatefulSet(ctx context.Context, req ctrl.Request) (*appsv1.MyStatefulSet, error) {
	var myStatefulSet appsv1.MyStatefulSet
	err := r.Get(ctx, req.NamespacedName, &myStatefulSet)
	return &myStatefulSet, err
}

func (r *MyStatefulSetReconciler) listPods(ctx context.Context, req ctrl.Request, myStatefulSet *appsv1.MyStatefulSet) (*corev1.PodList, error) {
	podList := &corev1.PodList{}
	// 构建标签选择器
	labels := map[string]string{
		"mystatefulset-name": myStatefulSet.Name,
	}
	listOpts := []client.ListOption{
		client.InNamespace(req.Namespace),
		client.MatchingLabels(labels),
	}
	err := r.List(ctx, podList, listOpts...)
	return podList, err
}

func (r *MyStatefulSetReconciler) createMissingPodsAndPVCs(ctx context.Context, req ctrl.Request, myStatefulSet *appsv1.MyStatefulSet, podList *corev1.PodList, desiredReplicas int32) error {
	for i := int32(0); i < desiredReplicas; i++ {
		podName := fmt.Sprintf("%s-%d", myStatefulSet.Name, i)
		if !podExists(podName, podList) {
			if err := r.createPVCs(ctx, req, myStatefulSet, i); err != nil {
				return err
			}
			if err := r.createPod(ctx, req, myStatefulSet, podName); err != nil {
				return err
			}
		}
	}
	return nil
}

func podExists(podName string, podList *corev1.PodList) bool {
	for _, pod := range podList.Items {
		if pod.Name == podName {
			return true
		}
	}
	return false
}

func (r *MyStatefulSetReconciler) createPVCs(ctx context.Context, req ctrl.Request, myStatefulSet *appsv1.MyStatefulSet, ordinal int32) error {
	for _, pvcTemplate := range myStatefulSet.Spec.VolumeClaimTemplates {
		pvcName := fmt.Sprintf("%s-%s-%d", pvcTemplate.Name, myStatefulSet.Name, ordinal)
		newPVC := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:        pvcName,
				Namespace:   req.Namespace,
				Labels:      createLabels(pvcTemplate.Labels, myStatefulSet.Name),
				Annotations: pvcTemplate.Annotations,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(myStatefulSet, appsv1.GroupVersion.WithKind("MyStatefulSet")),
				},
			},
			Spec: pvcTemplate.Spec,
		}
		if err := r.Create(ctx, newPVC); err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}

func (r *MyStatefulSetReconciler) createPod(ctx context.Context, req ctrl.Request, myStatefulSet *appsv1.MyStatefulSet, podName string) error {
	newPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        podName,
			Namespace:   req.Namespace,
			Labels:      createLabels(myStatefulSet.Spec.Template.Labels, myStatefulSet.Name),
			Annotations: myStatefulSet.Spec.Template.Annotations,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(myStatefulSet, appsv1.GroupVersion.WithKind("MyStatefulSet")),
			},
		},
		Spec: myStatefulSet.Spec.Template.Spec,
	}
	if err := r.Create(ctx, newPod); err != nil {
		return err
	}

	return nil
}

func (r *MyStatefulSetReconciler) updatePods(ctx context.Context, myStatefulSet *appsv1.MyStatefulSet, podList *corev1.PodList, req ctrl.Request) error {
	for _, pod := range podList.Items {
		if podNeedsUpdate(&pod, myStatefulSet.Spec.Template) {
			if err := r.Delete(ctx, &pod); err != nil {
				return err
			}
			time.Sleep(3 * time.Second)
			// 在删除后重新创建 Pod
			if err := r.createPod(ctx, req, myStatefulSet, pod.Name); err != nil {
				return err
			}
			break // 一次只更新一个 Pod，确保有序性
		}
	}
	return nil
}

func (r *MyStatefulSetReconciler) cleanupMyStatefulSet(ctx context.Context, myStatefulSet *appsv1.MyStatefulSet, req ctrl.Request) (ctrl.Result, error) {
	podList, err := r.listPods(ctx, req, myStatefulSet)
	if err != nil {
		return ctrl.Result{}, err
	}

	for _, pod := range podList.Items {
		if err := r.Delete(ctx, &pod); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.deletePVCs(ctx, myStatefulSet, pod.Name); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *MyStatefulSetReconciler) deletePVCs(ctx context.Context, myStatefulSet *appsv1.MyStatefulSet, podName string) error {
	for _, pvcTemplate := range myStatefulSet.Spec.VolumeClaimTemplates {
		pvcName := fmt.Sprintf("%s-%s", pvcTemplate.Name, podName)
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pvcName,
				Namespace: myStatefulSet.Namespace,
			},
		}
		if err := r.Delete(ctx, pvc); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

// podNeedsUpdate 判断 Pod 是否需要更新
func podNeedsUpdate(pod *corev1.Pod, desiredPodTemplate corev1.PodTemplateSpec) bool {
	// 检查容器数量是否一致
	if len(pod.Spec.Containers) != len(desiredPodTemplate.Spec.Containers) {
		return true
	}

	// 检查每个容器的镜像版本
	for i, container := range pod.Spec.Containers {
		desiredContainer := desiredPodTemplate.Spec.Containers[i]

		// 如果镜像名称或版本不一致，则需要更新
		if container.Image != desiredContainer.Image {
			return true
		}
	}

	// 检查标签是否一致
	for key, value := range desiredPodTemplate.Labels {
		if pod.Labels[key] != value {
			return true
		}
	}

	// 检查注解是否一致
	for key, value := range desiredPodTemplate.Annotations {
		if pod.Annotations[key] != value {
			return true
		}
	}

	// 如果没有需要更新的地方，则返回 false
	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyStatefulSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.MyStatefulSet{}).
		Complete(r)
}

// createLabels 创建一个包含 MyStatefulSet 的模板标签和名称的标签集合
func createLabels(oldMap map[string]string, name string) map[string]string {
	// 创建一个新的 map 来存储合并后的标签
	labels := make(map[string]string)

	// 将模板中的标签复制到新的 map 中
	for key, value := range oldMap {
		labels[key] = value
	}

	// 添加 MyStatefulSet 的名称作为一个新的标签
	labels["mystatefulset-name"] = name

	return labels
}
