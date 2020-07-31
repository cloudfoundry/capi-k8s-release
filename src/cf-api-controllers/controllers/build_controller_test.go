package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/model"
	"github.com/buildpacks/lifecycle"
	"github.com/buildpacks/lifecycle/launch"
	ociv1 "github.com/google/go-containerregistry/pkg/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	buildv1alpha1 "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("BuildController", func() {
	var (
		subject               *buildv1alpha1.Build
		buildGUID             string
		receivedApiBuildPatch chan model.Build
		updatedBuildStatus    buildv1alpha1.BuildStatus
	)
	BeforeEach(func() {
		raw, err := json.Marshal(lifecycle.BuildMetadata{
			Processes: []launch.Process{{
				Type:    "baz",
				Command: "some-start-command",
			}},
		})
		Expect(err).To(BeNil())
		mockImageConfigFetcher.FetchImageConfigReturns(&ociv1.Config{
			Labels: map[string]string{lifecycle.BuildMetadataLabel: string(raw)},
		}, nil)

		buildGUID = fmt.Sprintf("build-guid-%d", GinkgoRandomSeed())
		receivedApiBuildPatch = make(chan model.Build)

		fakeCFAPIServer.Reset()
		fakeCFAPIServer.AppendHandlers(ghttp.CombineHandlers(
			ghttp.VerifyRequest("PATCH", "/v3/builds/"+buildGUID),
			ghttp.VerifyHeaderKV("Authorization", "Bearer"),
			func(_ http.ResponseWriter, r *http.Request) {
				bytes, err := ioutil.ReadAll(r.Body)
				Expect(err).NotTo(HaveOccurred())

				var apiBuildPatch model.Build
				err = json.Unmarshal(bytes, &apiBuildPatch)
				Expect(err).NotTo(HaveOccurred())

				// send the build patch back to the test thread
				// so we can assert against it without sharing memory
				receivedApiBuildPatch <- apiBuildPatch
			},
		))

		updatedBuildStatus = buildv1alpha1.BuildStatus{
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
		}
	})

	AfterEach(func() {
		deleteBuild(subject)
	})

	Context("when the kpack build triggered by app update is valid", func() {
		BeforeEach(func() {
			subject = createBuild(&buildv1alpha1.Build{
				ObjectMeta: metav1.ObjectMeta{
					Name:        buildGUID,
					Namespace:   "default",
					Labels:      map[string]string{BuildGUIDLabel: buildGUID},
					Annotations: map[string]string{BuildReasonAnnotation: "src"},
				},
				Spec: buildv1alpha1.BuildSpec{},
				Status: buildv1alpha1.BuildStatus{
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
			})
		})

		It("marks successful builds as successful", func() {
			subject = updateBuildStatus(subject, &updatedBuildStatus)
			Eventually(fakeCFAPIServer.ReceivedRequests, time.Second*15).Should(HaveLen(1))

			var actualBuildPatch model.Build
			Eventually(receivedApiBuildPatch).Should(Receive(&actualBuildPatch))

			Expect(actualBuildPatch.State).To(Equal(model.BuildStagedState))
			Expect(actualBuildPatch.Lifecycle.Data.Image).To(Equal("foo.bar/here/be/an/image"))
			Expect(actualBuildPatch.Lifecycle.Data.ProcessTypes).To(HaveLen(1))
			Expect(actualBuildPatch.Lifecycle.Data.ProcessTypes).To(HaveKeyWithValue("baz", "some-start-command"))
		})

		Context("when the build fails with a failed container", func() {
			BeforeEach(func() {
				updatedBuildStatus.Status.Conditions[0].Status = corev1.ConditionFalse
				updatedBuildStatus.StepStates[0].Terminated = &corev1.ContainerStateTerminated{
					ExitCode: 1,
					Message:  "what do we say to passing tests? not today",
				}
			})

			It("marks the CC build as failed", func() {
				subject = updateBuildStatus(subject, &updatedBuildStatus)
				Eventually(fakeCFAPIServer.ReceivedRequests, time.Second*15).Should(HaveLen(1))

				var actualBuildPatch model.Build
				Eventually(receivedApiBuildPatch).Should(Receive(&actualBuildPatch))

				Expect(actualBuildPatch.State).To(Equal(model.BuildFailedState))
				Expect(actualBuildPatch.Error).To(ContainSubstring(subject.Status.StepStates[0].Terminated.Message))
			})
		})

		Context("and there is an error fetching process types from image config", func() {
			BeforeEach(func() {
				mockImageConfigFetcher.FetchImageConfigReturns(nil, errors.New("fake error: couldn't fetch image config"))
			})

			It("successfully marks the CC v3 build as failed", func() {
				subject = updateBuildStatus(subject, &updatedBuildStatus)

				var actualBuildPatch model.Build
				Eventually(receivedApiBuildPatch).Should(Receive(&actualBuildPatch))

				Expect(actualBuildPatch.State).To(Equal(model.BuildFailedState))
				Expect(actualBuildPatch.Error).To(Equal("Failed to handle successful kpack build: fake error: couldn't fetch image config"))
			})
		})

		Context("and the cloud controller responds with an error", func() {
			BeforeEach(func() {
				fakeCFAPIServer.Reset()
				fakeCFAPIServer.AppendHandlers(
					ghttp.RespondWith(500, ""),
					// fail for the first request, unmarshal the 2nd request to our buildpack channel
					func(_ http.ResponseWriter, r *http.Request) {
						bytes, err := ioutil.ReadAll(r.Body)
						Expect(err).NotTo(HaveOccurred())

						var buildPatch model.Build
						json.Unmarshal(bytes, &buildPatch)
						receivedApiBuildPatch <- buildPatch
					},
				)
			})

			It("requeues the Build resource and eventually reconciles again", func() {
				subject = updateBuildStatus(subject, &updatedBuildStatus)
				Eventually(fakeCFAPIServer.ReceivedRequests, time.Second*30).Should(HaveLen(2))

				var actualBuildPatch model.Build
				Eventually(receivedApiBuildPatch).Should(Receive(&actualBuildPatch))

				Expect(actualBuildPatch.State).To(Equal(model.BuildStagedState))
				Expect(actualBuildPatch.Lifecycle.Data.Image).To(Equal("foo.bar/here/be/an/image"))
			})
		})
	})

	Context("when there is a kpack build without a CF build guid", func() {
		BeforeEach(func() {
			subject = createBuild(&buildv1alpha1.Build{
				ObjectMeta: metav1.ObjectMeta{
					Name:        buildGUID,
					Namespace:   "default",
					Labels:      map[string]string{},
					Annotations: map[string]string{BuildReasonAnnotation: "src"},
				},
				Spec: buildv1alpha1.BuildSpec{},
				Status: buildv1alpha1.BuildStatus{
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
			})
		})

		It("ignores the build", func() {
			subject = updateBuildStatus(subject, &updatedBuildStatus)
			Consistently(fakeCFAPIServer.ReceivedRequests, time.Second*15).Should(HaveLen(0))
		})
	})

	Context("when there is a kpack build without a build reason", func() {
		BeforeEach(func() {
			subject = createBuild(&buildv1alpha1.Build{
				ObjectMeta: metav1.ObjectMeta{
					Name:      buildGUID,
					Namespace: "default",
					Labels:    map[string]string{BuildGUIDLabel: buildGUID},
				},
				Spec: buildv1alpha1.BuildSpec{},
				Status: buildv1alpha1.BuildStatus{
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
			})
		})

		It("ignores the build", func() {
			subject = updateBuildStatus(subject, &updatedBuildStatus)
			Consistently(fakeCFAPIServer.ReceivedRequests, time.Second*15).Should(HaveLen(0))
		})
	})

	// this is so the image controller picks this change up instead
	Context("when there is a kpack build with a build reason indicating a stack update", func() {
		BeforeEach(func() {
			subject = createBuild(&buildv1alpha1.Build{
				ObjectMeta: metav1.ObjectMeta{
					Name:        buildGUID,
					Namespace:   "default",
					Labels:      map[string]string{BuildGUIDLabel: buildGUID},
					Annotations: map[string]string{BuildReasonAnnotation: "STACK"},
				},
				Spec: buildv1alpha1.BuildSpec{},
				Status: buildv1alpha1.BuildStatus{
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
			})
		})

		It("ignores the build", func() {
			subject = updateBuildStatus(subject, &updatedBuildStatus)
			Consistently(fakeCFAPIServer.ReceivedRequests, time.Second*15).Should(HaveLen(0))
		})
	})
})

// can be used in top level beforeach
func createBuild(build *buildv1alpha1.Build) *buildv1alpha1.Build {
	Expect(k8sClient.Create(context.Background(), build)).Should(Succeed())
	var createdBuild buildv1alpha1.Build
	Eventually(func() error {
		return k8sClient.Get(context.Background(), namespacedName(build), &createdBuild)
	}, "5s", "100ms").Should(Succeed())
	return &createdBuild
}

// can be used in Its as subject
func updateBuildStatus(existingBuild *buildv1alpha1.Build, desiredBuildStatus *buildv1alpha1.BuildStatus) *buildv1alpha1.Build {
	// update build to update its status and wait for it to propagate
	existingBuild.Status = *desiredBuildStatus
	Expect(k8sClient.Status().Update(context.Background(), existingBuild)).Should(Succeed())

	var updatedBuild buildv1alpha1.Build
	Eventually(func() bool {
		err := k8sClient.Get(context.Background(), namespacedName(existingBuild), &updatedBuild)
		if err != nil {
			panic(err)
		}
		return !updatedBuild.Status.GetCondition(corev1alpha1.ConditionSucceeded).IsUnknown()
	}, "5s", "100ms").Should(BeTrue())
	Expect(updatedBuild).ToNot(BeNil())

	return &updatedBuild
}

func namespacedName(build *buildv1alpha1.Build) types.NamespacedName {
	return types.NamespacedName{
		Name:      build.ObjectMeta.Name,
		Namespace: build.ObjectMeta.Namespace,
	}
}

func deleteBuild(subject *buildv1alpha1.Build) {
	Expect(k8sClient.Delete(context.Background(), subject)).To(BeNil())
	Eventually(func() error {
		obj := &buildv1alpha1.Build{}
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
