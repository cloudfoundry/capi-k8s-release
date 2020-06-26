package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-kpack-watcher/capi/capifakes"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-kpack-watcher/capi_model"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kpackv1alpha1 "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Controllers/BuildController", func() {
	Context("When a Build is completed", func() {
		var subject *kpackv1alpha1.Build

		BeforeEach(func() {
			mockRestClient.PatchReturns(&http.Response{
				StatusCode: 200,
			}, nil)
		})

		AfterEach(func() {
			deleteBuild(subject)
			// clean up mocks
			*mockRestClient = capifakes.FakeRest{}
			*mockUAAClient = capifakes.FakeTokenFetcher{}
		})

		It("successfully marks the CC v3 build as staged", func() {
			buildGUID := "here-be-a-guid"
			subject = createBuildAndUpdateStatus(buildGUID, kpackv1alpha1.BuildStatus{
				Status: corev1alpha1.Status{
					Conditions: []corev1alpha1.Condition{
						corev1alpha1.Condition{
							Type:   corev1alpha1.ConditionSucceeded,
							Status: corev1.ConditionTrue,
						},
					},
				},
				StepStates: []corev1.ContainerState{
					corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							ExitCode: 0,
						},
					},
				},
				LatestImage: "foo.bar/here/be/an/image",
			})

			// eventually expect CF API/CCNG to receive request to update its "v3 build" object
			Eventually(func() int {
				// TODO: figure out how to get rid of this horrible sleep
				// need an initial sleep because of some suspected weirdness about how `counterfeiter`
				// takes some time to release some mutexes it uses for counting stub usages
				time.Sleep(1 * time.Second)
				return mockRestClient.PatchCallCount()
			}).Should(Equal(1))
			url, _, body := mockRestClient.PatchArgsForCall(0)
			Expect(url).To(Equal(fmt.Sprintf("https://cf.api/v3/builds/%s", buildGUID)))

			raw, err := ioutil.ReadAll(body)
			Expect(err).ToNot(HaveOccurred())

			var updateBuildRequest capi_model.Build
			err = json.Unmarshal(raw, &updateBuildRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(updateBuildRequest.State).To(Equal(capi_model.BuildStagedState))
			Expect(updateBuildRequest.Lifecycle.Data.Image).To(Equal("foo.bar/here/be/an/image"))
		})

		XContext("and the cloud controller responds with an error", func() {
			It("requeues the Build resource and eventually reconciles again", func() {
				// TODO
			})
		})
	})

	Context("When a Build has failed", func() {
		var subject *kpackv1alpha1.Build

		BeforeEach(func() {
			mockRestClient.PatchReturns(&http.Response{
				StatusCode: 200,
			}, nil)
		})

		AfterEach(func() {
			deleteBuild(subject)
			// clean up mocks
			*mockRestClient = capifakes.FakeRest{}
			*mockUAAClient = capifakes.FakeTokenFetcher{}
		})

		It("successfully marks the CC v3 build as failed to have staged", func() {
			buildGUID := "here-be-a-guid"
			subject = createBuildAndUpdateStatus(buildGUID, kpackv1alpha1.BuildStatus{
				Status: corev1alpha1.Status{
					Conditions: []corev1alpha1.Condition{
						corev1alpha1.Condition{
							Type:   corev1alpha1.ConditionSucceeded,
							Status: corev1.ConditionFalse,
						},
					},
				},
				StepStates: []corev1.ContainerState{
					corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							ExitCode: 1,
							Message:  "what do we say to passing tests? not today",
						},
					},
				},
			})

			// eventually expect CF API/CCNG to receive request to update its "v3 build" object
			Eventually(func() int {
				// TODO: figure out how to get rid of this horrible sleep
				// need an initial sleep because of some suspected weirdness about how `counterfeiter`
				// takes some time to release some mutexes it uses for counting stub usages
				time.Sleep(1 * time.Second)
				return mockRestClient.PatchCallCount()
			}).Should(Equal(1))
			url, _, body := mockRestClient.PatchArgsForCall(0)
			Expect(url).To(Equal(fmt.Sprintf("https://cf.api/v3/builds/%s", buildGUID)))

			raw, err := ioutil.ReadAll(body)
			Expect(err).ToNot(HaveOccurred())

			var updateBuildRequest capi_model.Build
			err = json.Unmarshal(raw, &updateBuildRequest)
			Expect(err).ToNot(HaveOccurred())
			Expect(updateBuildRequest.State).To(Equal(capi_model.BuildFailedState))
			Expect(updateBuildRequest.Error).To(Equal(subject.Status.StepStates[0].Terminated.Message))
		})
	})
})

func createBuildAndUpdateStatus(desiredBuildGUID string, desiredBuildStatus kpackv1alpha1.BuildStatus) *kpackv1alpha1.Build {
	key := types.NamespacedName{Name: desiredBuildGUID, Namespace: "default"}
	completedBuild := &kpackv1alpha1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
			Labels:    map[string]string{BuildGUIDLabel: desiredBuildGUID},
		},
		Spec: kpackv1alpha1.BuildSpec{},
		Status: kpackv1alpha1.BuildStatus{
			Status: corev1alpha1.Status{
				Conditions: []corev1alpha1.Condition{
					corev1alpha1.Condition{
						Type:   corev1alpha1.ConditionSucceeded,
						Status: corev1.ConditionUnknown,
					},
				},
			},
			StepStates: []corev1.ContainerState{
				corev1.ContainerState{},
			},
		},
	}

	// create build (status info isn't persisted to API) and wait for it to propagate
	Expect(k8sClient.Create(context.Background(), completedBuild)).Should(Succeed())
	Eventually(func() error {
		obj := &kpackv1alpha1.Build{}
		return k8sClient.Get(context.Background(), key, obj)
	}, "5s", "100ms").Should(Succeed())

	// update build to update its status and wait for it to propagate
	var updatedBuild *kpackv1alpha1.Build
	completedBuild.Status = desiredBuildStatus
	Expect(k8sClient.Status().Update(context.Background(), completedBuild)).Should(Succeed())
	Eventually(func() bool {
		updatedBuild = &kpackv1alpha1.Build{}
		err := k8sClient.Get(context.Background(), key, updatedBuild)
		if err != nil {
			panic(err)
		}
		return !updatedBuild.Status.GetCondition(corev1alpha1.ConditionSucceeded).IsUnknown()
	}, "5s", "100ms").Should(BeTrue())

	return updatedBuild
}

func deleteBuild(subject *kpackv1alpha1.Build) {
	Expect(k8sClient.Delete(context.Background(), subject)).To(BeNil())
	Eventually(func() error {
		obj := &kpackv1alpha1.Build{}
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
