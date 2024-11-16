package v1

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "my.com/devops-golang-test/api/v1"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

type MockObject struct{}

func (m *MockObject) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

func (m *MockObject) DeepCopyObject() runtime.Object {
	return &MockObject{}
}

var _ = Describe("MyStatefulSet Webhook", func() {
	var (
		obj       *appsv1.MyStatefulSet
		oldObj    *appsv1.MyStatefulSet
		validator MyStatefulSetCustomValidator
		defaulter MyStatefulSetCustomDefaulter
		ctx       context.Context
	)

	BeforeEach(func() {
		obj = &appsv1.MyStatefulSet{}
		oldObj = &appsv1.MyStatefulSet{}
		validator = MyStatefulSetCustomValidator{}
		defaulter = MyStatefulSetCustomDefaulter{}
		ctx = context.TODO()
	})

	Context("When creating MyStatefulSet under Defaulting Webhook", func() {
		It("Should apply defaults when replicas is not set", func() {
			By("simulating a scenario where replicas is not set")
			obj.Spec.Replicas = nil
			By("calling the Default method to apply defaults")
			err := defaulter.Default(ctx, obj)
			Expect(err).ToNot(HaveOccurred())
			By("checking that the default value for replicas is set to 1")
			Expect(obj.Spec.Replicas).ToNot(BeNil())
			Expect(*obj.Spec.Replicas).To(Equal(int32(1)))
		})
	})

	Context("When creating or updating MyStatefulSet under Validating Webhook", func() {
		It("Should deny creation if replicas is less than 1", func() {
			By("simulating an invalid creation scenario with replicas less than 1")
			replicas := int32(0)
			obj.Spec.Replicas = &replicas
			_, err := validator.ValidateCreate(ctx, obj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("replicas must be greater than or equal to 1"))

			incorrectObj := &MockObject{} // Using an empty struct to simulate incorrect type
			_, err = validator.ValidateCreate(ctx, incorrectObj)
			Expect(err.Error()).To(ContainSubstring("expected a MyStatefulSet object but got"))
		})

		It("Should admit creation if replicas is 1 or more", func() {
			By("simulating a valid creation scenario with replicas set to 1")
			replicas := int32(1)
			obj.Spec.Replicas = &replicas
			_, err := validator.ValidateCreate(ctx, obj)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should validate updates correctly", func() {
			By("simulating a valid update scenario with replicas set to 2")
			oldReplicas := int32(1)
			newReplicas := int32(2)
			oldObj.Spec.Replicas = &oldReplicas
			obj.Spec.Replicas = &newReplicas
			_, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("When deleting MyStatefulSet under Validating Webhook", func() {
		It("Should validate deletion correctly", func() {
			By("simulating a deletion scenario")
			_, err := validator.ValidateDelete(ctx, obj)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
