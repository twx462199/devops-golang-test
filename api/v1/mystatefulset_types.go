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

package v1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MyStatefulSetSpec defines the desired state of MyStatefulSet.
type MyStatefulSetSpec struct {
	// Replicas 是期望的副本数量
	Replicas *int32 `json:"replicas,omitempty"`

	// Selector 是标签选择器，用于选择管理的 Pod
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// Template 是 Pod 的模板
	Template corev1.PodTemplateSpec `json:"template"`

	// VolumeClaimTemplates 是用于创建 PVC 的模板
	VolumeClaimTemplates []corev1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`

	// ServiceName 是用于管理 Pod 的服务名称
	ServiceName string `json:"serviceName"`

	// PodManagementPolicy 控制 Pod 的管理策略
	PodManagementPolicy appsv1.PodManagementPolicyType `json:"podManagementPolicy,omitempty"`

	// UpdateStrategy 控制 StatefulSet 的更新策略
	UpdateStrategy appsv1.StatefulSetUpdateStrategy `json:"updateStrategy,omitempty"`

	// RevisionHistoryLimit 是保留的历史修订版本的数量
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty"`
}

// MyStatefulSetStatus defines the observed state of MyStatefulSet.
type MyStatefulSetStatus struct {
	// ObservedGeneration 是观察到的最新生成
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Replicas 是当前的副本数量
	Replicas int32 `json:"replicas,omitempty"`

	// ReadyReplicas 是当前就绪的副本数量
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// CurrentReplicas 是当前版本的副本数量
	CurrentReplicas int32 `json:"currentReplicas,omitempty"`

	// UpdatedReplicas 是更新后的副本数量
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty"`

	// CurrentRevision 是当前版本的修订
	CurrentRevision string `json:"currentRevision,omitempty"`

	// UpdateRevision 是更新后的版本修订
	UpdateRevision string `json:"updateRevision,omitempty"`

	// CollisionCount 是检测到的版本冲突次数
	CollisionCount *int32 `json:"collisionCount,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// MyStatefulSet is the Schema for the mystatefulsets API.
type MyStatefulSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MyStatefulSetSpec   `json:"spec,omitempty"`
	Status MyStatefulSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MyStatefulSetList contains a list of MyStatefulSet.
type MyStatefulSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MyStatefulSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MyStatefulSet{}, &MyStatefulSetList{})
}
