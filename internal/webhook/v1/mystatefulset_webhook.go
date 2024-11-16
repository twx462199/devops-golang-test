package v1

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	appsv1 "my.com/devops-golang-test/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var mystatefulsetlog = logf.Log.WithName("mystatefulset-resource")

// SetupMyStatefulSetWebhookWithManager registers the webhook for MyStatefulSet in the manager.
func SetupMyStatefulSetWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&appsv1.MyStatefulSet{}).
		WithValidator(&MyStatefulSetCustomValidator{}).
		WithDefaulter(&MyStatefulSetCustomDefaulter{}).
		Complete()
}

// MyStatefulSetCustomDefaulter sets default values for MyStatefulSet
type MyStatefulSetCustomDefaulter struct{}

var _ webhook.CustomDefaulter = &MyStatefulSetCustomDefaulter{}

// Default sets default values for MyStatefulSet
func (d *MyStatefulSetCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	mystatefulset, ok := obj.(*appsv1.MyStatefulSet)
	if !ok {
		return fmt.Errorf("expected a MyStatefulSet object but got %T", obj)
	}
	mystatefulsetlog.Info("Defaulting MyStatefulSet", "name", mystatefulset.GetName())

	// Example defaulting logic
	if mystatefulset.Spec.Replicas == nil {
		defaultReplicas := int32(1)
		mystatefulset.Spec.Replicas = &defaultReplicas
		mystatefulsetlog.Info("Setting default replicas", "replicas", defaultReplicas)
	}

	return nil
}

// MyStatefulSetCustomValidator validates MyStatefulSet
type MyStatefulSetCustomValidator struct{}

var _ webhook.CustomValidator = &MyStatefulSetCustomValidator{}

// ValidateCreate validates MyStatefulSet upon creation
func (v *MyStatefulSetCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	mystatefulset, ok := obj.(*appsv1.MyStatefulSet)
	if !ok {
		return nil, fmt.Errorf("expected a MyStatefulSet object but got %T", obj)
	}
	mystatefulsetlog.Info("Validating MyStatefulSet creation", "name", mystatefulset.GetName())

	// Example validation logic
	if mystatefulset.Spec.Replicas != nil && *mystatefulset.Spec.Replicas < 1 {
		return nil, fmt.Errorf("replicas must be greater than or equal to 1")
	}

	return nil, nil
}

// ValidateUpdate validates MyStatefulSet upon update
func (v *MyStatefulSetCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	newStatefulset, ok := newObj.(*appsv1.MyStatefulSet)
	if !ok {
		return nil, fmt.Errorf("expected a MyStatefulSet object for the newObj but got %T", newObj)
	}
	mystatefulsetlog.Info("Validating MyStatefulSet update", "name", newStatefulset.GetName())

	// Example validation logic
	if newStatefulset.Spec.Replicas != nil && *newStatefulset.Spec.Replicas < 1 {
		return nil, fmt.Errorf("replicas must be greater than or equal to 1")
	}

	return nil, nil
}

// ValidateDelete validates MyStatefulSet upon deletion
func (v *MyStatefulSetCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	mystatefulset, ok := obj.(*appsv1.MyStatefulSet)
	if !ok {
		return nil, fmt.Errorf("expected a MyStatefulSet object but got %T", obj)
	}
	mystatefulsetlog.Info("Validating MyStatefulSet deletion", "name", mystatefulset.GetName())

	// No specific validation logic for deletion in this example
	return nil, nil
}
