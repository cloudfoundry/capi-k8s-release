package controllers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	buildv1alpha1 "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("ImageController", func() {
	var (
		subject            *buildv1alpha1.Image
		appStatefulSet     *appsv1.StatefulSet
		updatedImageStatus buildv1alpha1.ImageStatus
	)
	BeforeEach(func() {
		appStatefulSet = createStatefulSet(&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "statefulset-name",
				Namespace: "cf-workloads",
				Labels:    map[string]string{"cloudfoundry.org/app_guid": "some-app-guid-123"},
			},
			Spec: appsv1.StatefulSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"cloudfoundry.org/app_guid": "some-app-guid-123"},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"cloudfoundry.org/app_guid": "some-app-guid-123"},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							corev1.Container{
								Name:  "opi",
								Image: "pre-stack-update-image-reference",
							},
						},
					},
				},
			},
		})
		updatedImageStatus = buildv1alpha1.ImageStatus{
			Status: corev1alpha1.Status{
				Conditions: []corev1alpha1.Condition{
					corev1alpha1.Condition{
						Type:   corev1alpha1.ConditionReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
			LatestBuildReason: "STACK",
			LatestImage:       "post-stack-update-image-reference",
		}
	})

	AfterEach(func() {
		deleteImage(subject)
		deleteStatefulSet(appStatefulSet)
	})

	Context("when the kpack build triggered by a stack update is valid", func() {
		BeforeEach(func() {
			subject = createImage(&buildv1alpha1.Image{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "image",
					Namespace: "default",
					Labels:    map[string]string{"cloudfoundry.org/app_guid": "some-app-guid-123"},
				},
				Spec: buildv1alpha1.ImageSpec{},
				Status: buildv1alpha1.ImageStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							corev1alpha1.Condition{
								Type:   corev1alpha1.ConditionReady,
								Status: corev1.ConditionUnknown,
							},
						},
					},
				},
			})
		})

		It("updates the associated statefulset with the new image reference", func() {
			subject = updateImageStatus(subject, &updatedImageStatus)

			Eventually(func() bool {
				statefulset := appsv1.StatefulSet{}
				statefulsetNamespacedName := types.NamespacedName{
					Name:      appStatefulSet.ObjectMeta.Name,
					Namespace: appStatefulSet.ObjectMeta.Namespace,
				}
				Expect(k8sClient.Get(context.Background(), statefulsetNamespacedName, &statefulset)).To(Succeed())
				return statefulset.Spec.Template.Spec.Containers[0].Image == "post-stack-update-image-reference"
			}, "5s", "100ms").Should(BeTrue())
		})
	})

	Context("when there is a kpack image without a CF image guid", func() {
		BeforeEach(func() {
			subject = createImage(&buildv1alpha1.Image{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "image",
					Namespace: "default",
				},
				Spec: buildv1alpha1.ImageSpec{},
				Status: buildv1alpha1.ImageStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							corev1alpha1.Condition{
								Type:   corev1alpha1.ConditionReady,
								Status: corev1.ConditionUnknown,
							},
						},
					},
				},
			})
		})

		It("ignores the image", func() {
			subject = updateImageStatus(subject, &updatedImageStatus)

			// TODO: actually test somehow that the statefulset remains untouched...(eventual consistency is hard to deal with)
			statefulset := appsv1.StatefulSet{}
			statefulsetNamespacedName := types.NamespacedName{
				Name:      appStatefulSet.ObjectMeta.Name,
				Namespace: appStatefulSet.ObjectMeta.Namespace,
			}
			Expect(k8sClient.Get(context.Background(), statefulsetNamespacedName, &statefulset)).To(Succeed())
			Expect(statefulset.Spec.Template.Spec.Containers[0].Image).To(Equal("pre-stack-update-image-reference"))
		})
	})
})

// can be used in top level beforeach
func createImage(image *buildv1alpha1.Image) *buildv1alpha1.Image {
	Expect(k8sClient.Create(context.Background(), image)).Should(Succeed())
	var createdImage buildv1alpha1.Image
	Eventually(func() error {
		return k8sClient.Get(context.Background(), namespacedImageName(image), &createdImage)
	}, "5s", "100ms").Should(Succeed())
	return &createdImage
}

func createStatefulSet(statefulset *appsv1.StatefulSet) *appsv1.StatefulSet {
	Expect(k8sClient.Create(context.Background(), statefulset)).Should(Succeed())
	var createdStatefulset appsv1.StatefulSet
	statefulsetNamespacedName := types.NamespacedName{
		Name:      statefulset.ObjectMeta.Name,
		Namespace: statefulset.ObjectMeta.Namespace,
	}
	Eventually(func() error {
		return k8sClient.Get(context.Background(), statefulsetNamespacedName, &createdStatefulset)
	}, "5s", "100ms").Should(Succeed())
	return &createdStatefulset
}

// can be used in Its as subject
func updateImageStatus(existingImage *buildv1alpha1.Image, desiredImageStatus *buildv1alpha1.ImageStatus) *buildv1alpha1.Image {
	// update build to update its status and wait for it to propagate
	existingImage.Status = *desiredImageStatus
	Expect(k8sClient.Status().Update(context.Background(), existingImage)).Should(Succeed())

	var updatedImage buildv1alpha1.Image
	Eventually(func() bool {
		err := k8sClient.Get(context.Background(), namespacedImageName(existingImage), &updatedImage)
		if err != nil {
			panic(err)
		}
		return !updatedImage.Status.GetCondition(corev1alpha1.ConditionReady).IsUnknown()
	}, "5s", "100ms").Should(BeTrue())
	Expect(updatedImage).ToNot(BeNil())

	return &updatedImage
}

func namespacedImageName(image *buildv1alpha1.Image) types.NamespacedName {
	return types.NamespacedName{
		Name:      image.ObjectMeta.Name,
		Namespace: image.ObjectMeta.Namespace,
	}
}

func deleteImage(subject *buildv1alpha1.Image) {
	Expect(k8sClient.Delete(context.Background(), subject)).To(BeNil())
	Eventually(func() error {
		obj := &buildv1alpha1.Image{}
		return k8sClient.Get(
			context.Background(),
			types.NamespacedName{
				Name:      subject.ObjectMeta.Name,
				Namespace: subject.ObjectMeta.Namespace,
			},
			obj,
		)
	}, "5s", "100ms").ShouldNot(Succeed())
}

func deleteStatefulSet(subject *appsv1.StatefulSet) {
	Expect(k8sClient.Delete(context.Background(), subject)).To(BeNil())
	Eventually(func() error {
		obj := &appsv1.StatefulSet{}
		return k8sClient.Get(
			context.Background(),
			types.NamespacedName{
				Name:      subject.ObjectMeta.Name,
				Namespace: subject.ObjectMeta.Namespace,
			},
			obj,
		)
	}, "5s", "100ms").ShouldNot(Succeed())
}
