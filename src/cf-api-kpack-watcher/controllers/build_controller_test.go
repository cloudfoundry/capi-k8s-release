package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-kpack-watcher/cf/api_model"
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
		subject            *buildv1alpha1.Build
		buildGUID          string
		receivedBuildPatch chan api_model.Build
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
		receivedBuildPatch = make(chan api_model.Build)

		fakeCFAPIServer.Reset()
		fakeCFAPIServer.AppendHandlers(ghttp.CombineHandlers(
			ghttp.VerifyRequest("PATCH", "/v3/builds/"+buildGUID),
			ghttp.VerifyHeaderKV("Authorization", "Bearer"),
			func(_ http.ResponseWriter, r *http.Request) {
				bytes, err := ioutil.ReadAll(r.Body)
				Expect(err).NotTo(HaveOccurred())
				var buildPatch api_model.Build
				json.Unmarshal(bytes, &buildPatch)
				receivedBuildPatch <- buildPatch
			},
		))
	})

	AfterEach(func() {
		deleteBuild(subject)
	})

	It("marks successful builds as successful", func() {
		subject = createBuildAndUpdateStatus(buildGUID, buildv1alpha1.BuildStatus{
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

		Eventually(fakeCFAPIServer.ReceivedRequests, time.Second*15).Should(HaveLen(1))

		var actualBuildPatch api_model.Build
		Eventually(receivedBuildPatch).Should(Receive(&actualBuildPatch))

		Expect(actualBuildPatch.State).To(Equal(api_model.BuildStagedState))
		Expect(actualBuildPatch.Lifecycle.Data.Image).To(Equal("foo.bar/here/be/an/image"))
		Expect(actualBuildPatch.Lifecycle.Data.ProcessTypes).To(HaveLen(1))
		Expect(actualBuildPatch.Lifecycle.Data.ProcessTypes).To(HaveKeyWithValue("baz", "some-start-command"))
	})

	It("marks failed builds as failed", func() {
		subject = createBuildAndUpdateStatus(buildGUID, buildv1alpha1.BuildStatus{
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
		Eventually(fakeCFAPIServer.ReceivedRequests, time.Second*15).Should(HaveLen(1))

		var actualBuildPatch api_model.Build
		Eventually(receivedBuildPatch).Should(Receive(&actualBuildPatch))

		Expect(actualBuildPatch.State).To(Equal(api_model.BuildFailedState))
		Expect(actualBuildPatch.Error).To(ContainSubstring(subject.Status.StepStates[0].Terminated.Message))
	})

	Context("and there is an error fetching process types from image config", func() {
		BeforeEach(func() {
			mockImageConfigFetcher.FetchImageConfigReturns(nil, errors.New("fake error: couldn't fetch image config"))
		})

		It("successfully marks the CC v3 build as failed", func() {
			subject = createBuildAndUpdateStatus(buildGUID, buildv1alpha1.BuildStatus{
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

			var actualBuildPatch api_model.Build
			Eventually(receivedBuildPatch).Should(Receive(&actualBuildPatch))

			Expect(actualBuildPatch.State).To(Equal(api_model.BuildFailedState))
			Expect(actualBuildPatch.Error).To(Equal("Failed to handle successful kpack build: fake error: couldn't fetch image config"))
		})
	})

	Context("and the cloud controller responds with an error", func() {
		BeforeEach(func() {
			fakeCFAPIServer.Reset()
			// fail for the first request, unmarshal the 2nd request to actualBuildPatch
			fakeCFAPIServer.AppendHandlers(
				ghttp.RespondWith(500, ""),
				func(_ http.ResponseWriter, r *http.Request) {
					bytes, err := ioutil.ReadAll(r.Body)
					Expect(err).NotTo(HaveOccurred())

					var buildPatch api_model.Build
					json.Unmarshal(bytes, &buildPatch)
					receivedBuildPatch <- buildPatch
				},
			)
		})

		It("requeues the Build resource and eventually reconciles again", func() {
			subject = createBuildAndUpdateStatus(buildGUID, buildv1alpha1.BuildStatus{
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

			Eventually(fakeCFAPIServer.ReceivedRequests, time.Second*30).Should(HaveLen(2))

			var actualBuildPatch api_model.Build
			Eventually(receivedBuildPatch).Should(Receive(&actualBuildPatch))

			Expect(actualBuildPatch.State).To(Equal(api_model.BuildStagedState))
			Expect(actualBuildPatch.Lifecycle.Data.Image).To(Equal("foo.bar/here/be/an/image"))
		})
	})
})

func createBuildAndUpdateStatus(desiredBuildGUID string, desiredBuildStatus buildv1alpha1.BuildStatus) *buildv1alpha1.Build {
	key := types.NamespacedName{Name: desiredBuildGUID, Namespace: "default"}
	completedBuild := &buildv1alpha1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
			Labels:    map[string]string{BuildGUIDLabel: desiredBuildGUID},
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
	}

	// create build (status info isn't persisted to API) and wait for it to propagate
	Expect(k8sClient.Create(context.Background(), completedBuild)).Should(Succeed())
	Eventually(func() error {
		obj := &buildv1alpha1.Build{}
		return k8sClient.Get(context.Background(), key, obj)
	}, "5s", "100ms").Should(Succeed())

	// update build to update its status and wait for it to propagate
	var updatedBuild *buildv1alpha1.Build
	completedBuild.Status = desiredBuildStatus
	Expect(k8sClient.Status().Update(context.Background(), completedBuild)).Should(Succeed())
	Eventually(func() bool {
		updatedBuild = &buildv1alpha1.Build{}
		err := k8sClient.Get(context.Background(), key, updatedBuild)
		if err != nil {
			panic(err)
		}
		return !updatedBuild.Status.GetCondition(corev1alpha1.ConditionSucceeded).IsUnknown()
	}, "5s", "100ms").Should(BeTrue())
	Expect(updatedBuild).ToNot(BeNil())

	return updatedBuild
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
