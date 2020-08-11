package controllers

import (
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/model"
	"context"
	"encoding/json"
	"fmt"
	"github.com/onsi/gomega/ghttp"
	"io/ioutil"
	"net/http"
	"time"

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
		dropletGUID string
		receivedApiDropletPatch chan model.Droplet
		updatedImageStatus buildv1alpha1.ImageStatus
	)
	const (
		postStackUpdateImageReference = "post-stack-update-image-reference"
		preStackUpdateImageReference  = "pre-stack-update-image-reference"
	)

	BeforeEach(func() {
		appStatefulSet = createStatefulSet(&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "statefulset-name",
				Namespace: workloadsNamespace,
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
							{
								Name:  "opi",
								Image: preStackUpdateImageReference,
							},
						},
					},
				},
			},
		})
		dropletGUID = fmt.Sprintf("droplet-guid-%d", GinkgoRandomSeed())
		receivedApiDropletPatch = make(chan model.Droplet)

		fakeCFAPIServer.Reset()
		fakeCFAPIServer.AppendHandlers(ghttp.CombineHandlers(
			ghttp.VerifyRequest("PATCH", "/v3/droplets/"+dropletGUID),
			ghttp.VerifyHeaderKV("Authorization", "Bearer"),
			func(_ http.ResponseWriter, r *http.Request) {
				bytes, err := ioutil.ReadAll(r.Body)
				Expect(err).NotTo(HaveOccurred())

				var apiDropletPatch model.Droplet
				err = json.Unmarshal(bytes, &apiDropletPatch)
				Expect(err).NotTo(HaveOccurred())

				// send the droplet patch back to the test thread
				// so we can assert against it without sharing memory
				receivedApiDropletPatch <- apiDropletPatch
			},
		))
		updatedImageStatus = buildv1alpha1.ImageStatus{
			Status: corev1alpha1.Status{
				Conditions: []corev1alpha1.Condition{
					{
						Type:   corev1alpha1.ConditionReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
			LatestBuildReason: "STACK",
			LatestImage:       postStackUpdateImageReference,
		}
	})

	Context("when the kpack build triggered by a stack update is valid", func() {
		BeforeEach(func() {
			subject = createImage(&buildv1alpha1.Image{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "image",
					Namespace: stagingNamespace,
					Labels:    map[string]string{AppGUIDLabel: "some-app-guid-123", DropletGUIDLabel: dropletGUID},
				},
				Spec: buildv1alpha1.ImageSpec{},
				Status: buildv1alpha1.ImageStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							{
								Type:   corev1alpha1.ConditionReady,
								Status: corev1.ConditionUnknown,
							},
						},
					},
				},
			})
		})
		When("there is only one statefulset for the app", func() {
			It("updates the associated statefulset with the new image reference", func() {
				subject = updateImageStatus(subject, &updatedImageStatus)

				Eventually(func() bool {
					statefulset := appsv1.StatefulSet{}
					statefulsetNamespacedName := types.NamespacedName{
						Name:      appStatefulSet.ObjectMeta.Name,
						Namespace: appStatefulSet.ObjectMeta.Namespace,
					}
					Expect(k8sClient.Get(context.Background(), statefulsetNamespacedName, &statefulset)).To(Succeed())
					return statefulset.Spec.Template.Spec.Containers[0].Image == postStackUpdateImageReference
				}, "5s", "100ms").Should(BeTrue())

				Eventually(fakeCFAPIServer.ReceivedRequests, time.Second*15).Should(HaveLen(1))

				var actualDropletPatch model.Droplet
				Eventually(receivedApiDropletPatch).Should(Receive(&actualDropletPatch))

				Expect(actualDropletPatch.Image).To(Equal(postStackUpdateImageReference))
			})
		})

		When("and the cloud controller responds with an error", func() {
			BeforeEach(func() {
				fakeCFAPIServer.Reset()
				fakeCFAPIServer.AppendHandlers(
					ghttp.RespondWith(500, ""),
					func(_ http.ResponseWriter, r *http.Request) {
						bytes, err := ioutil.ReadAll(r.Body)
						Expect(err).NotTo(HaveOccurred())

						var dropletPatch model.Droplet
						json.Unmarshal(bytes, &dropletPatch)
						receivedApiDropletPatch <- dropletPatch
					},
				)
			})

			It("requeues the Image resource and eventually reconciles again", func() {
				subject = updateImageStatus(subject, &updatedImageStatus)
				Eventually(fakeCFAPIServer.ReceivedRequests, time.Second*30).Should(HaveLen(2))

				var actualDropletPatch model.Droplet
				Eventually(receivedApiDropletPatch).Should(Receive(&actualDropletPatch))

				Expect(actualDropletPatch.Image).To(Equal(postStackUpdateImageReference))
			})
		})

		When("there are multiple statefulsets for the app (i.e. during a rolling deployment)", func() {
			var (
				anotherAppStatefulSet        *appsv1.StatefulSet
				updatedSourceCodeImageStatus buildv1alpha1.ImageStatus
			)
			const oldStackNewSrc = "old-stack-new-src"

			BeforeEach(func() {
				anotherAppStatefulSet = createStatefulSet(&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "statefulset-2-name",
						Namespace: workloadsNamespace,
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
									{
										Name:  "opi",
										Image: oldStackNewSrc,
									},
								},
							},
						},
					},
				})
				updatedSourceCodeImageStatus = buildv1alpha1.ImageStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							{
								Type:   corev1alpha1.ConditionReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
					LatestBuildReason: "SRC",
					LatestImage:       oldStackNewSrc,
				}
			})

			It("does not update the statefulsets and requeues the update event", func() {
				subject = updateImageStatus(subject, &updatedSourceCodeImageStatus)
				subject = updateImageStatus(subject, &updatedImageStatus)
				Consistently(func() []string {
					statefulset1 := appsv1.StatefulSet{}
					statefulset1NamespacedName := types.NamespacedName{
						Name:      appStatefulSet.ObjectMeta.Name,
						Namespace: appStatefulSet.ObjectMeta.Namespace,
					}
					Expect(k8sClient.Get(context.Background(), statefulset1NamespacedName, &statefulset1)).To(Succeed())
					statefulset2 := appsv1.StatefulSet{}
					statefulset2NamespacedName := types.NamespacedName{
						Name:      anotherAppStatefulSet.ObjectMeta.Name,
						Namespace: anotherAppStatefulSet.ObjectMeta.Namespace,
					}
					Expect(k8sClient.Get(context.Background(), statefulset2NamespacedName, &statefulset2)).To(Succeed())
					return []string{statefulset1.Spec.Template.Spec.Containers[0].Image, statefulset2.Spec.Template.Spec.Containers[0].Image}
				}, "5s", "100ms").Should(Equal([]string{preStackUpdateImageReference, oldStackNewSrc}))

				deleteStatefulSet(appStatefulSet)

				Eventually(func() string {
					statefulset2 := appsv1.StatefulSet{}
					statefulset2NamespacedName := types.NamespacedName{
						Name:      anotherAppStatefulSet.ObjectMeta.Name,
						Namespace: anotherAppStatefulSet.ObjectMeta.Namespace,
					}
					Expect(k8sClient.Get(context.Background(), statefulset2NamespacedName, &statefulset2)).To(Succeed())
					return statefulset2.Spec.Template.Spec.Containers[0].Image
				}, "5s", "100ms").Should(Equal(postStackUpdateImageReference))

				Eventually(fakeCFAPIServer.ReceivedRequests, time.Second*15).Should(HaveLen(1))

				var actualDropletPatch model.Droplet
				Eventually(receivedApiDropletPatch).Should(Receive(&actualDropletPatch))

				Expect(actualDropletPatch.Image).To(Equal(postStackUpdateImageReference))
			})
		})

	})

	Context("when there is a kpack image without a CF image guid", func() {
		BeforeEach(func() {
			subject = createImage(&buildv1alpha1.Image{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "image",
					Namespace: stagingNamespace,
				},
				Spec: buildv1alpha1.ImageSpec{},
				Status: buildv1alpha1.ImageStatus{
					Status: corev1alpha1.Status{
						Conditions: []corev1alpha1.Condition{
							{
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
			Expect(statefulset.Spec.Template.Spec.Containers[0].Image).To(Equal(preStackUpdateImageReference))
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

func deleteStatefulSet(subject *appsv1.StatefulSet) {
	Expect(k8sClient.Delete(context.Background(), subject)).To(Succeed())
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
