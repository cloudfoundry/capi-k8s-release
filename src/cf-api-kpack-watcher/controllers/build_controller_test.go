package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-kpack-watcher/capi_model"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	kpackv1alpha1 "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Controllers/BuildController", func() {
	var (
		// err error
		// controller     BuildReconciler
		fakeCCNGServer *ghttp.Server
	)

	BeforeEach(func() {
		// controller = BuildReconciler{Client: k8sClient}
		fakeCCNGServer = ghttp.NewServer()
	})

	AfterEach(func() {
		fakeCCNGServer.Close()
	})

	Context("When a Build is completed", func() {
		BeforeEach(func() {
			mockRestClient.PatchReturns(&http.Response{
				StatusCode: 200,
			}, nil)
		})

		It("successfully marks the CC v3 build as staged", func() {
			key := types.NamespacedName{Name: "completed-build", Namespace: "default"}
			buildGUID := "here-be-a-guid"
			completedBuild := &kpackv1alpha1.Build{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
					Labels:    map[string]string{"cloudfoundry.org/build_guid": buildGUID},
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
					LatestImage: "foo.bar/here/be/an/image",
				},
			}

			// create build (status info isn't persisted to API) and wait for it to propagate
			Expect(k8sClient.Create(context.Background(), completedBuild)).Should(Succeed())
			Eventually(func() error {
				obj := &kpackv1alpha1.Build{}
				return k8sClient.Get(context.Background(), key, obj)
			}, "5s", "100ms").Should(Succeed())

			// update build to update its status and wait for it to propagate
			completedBuild.Status.Conditions[0].Status = corev1.ConditionTrue
			completedBuild.Status.StepStates[0].Terminated = &corev1.ContainerStateTerminated{ExitCode: 0}
			Expect(k8sClient.Status().Update(context.Background(), completedBuild)).Should(Succeed())
			Eventually(func() bool {
				build := &kpackv1alpha1.Build{}
				err := k8sClient.Get(context.Background(), key, build)
				if err != nil {
					return false
				}
				return build.Status.GetCondition(corev1alpha1.ConditionSucceeded).IsTrue()
			}, "5s", "100ms").Should(BeTrue())

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

	XContext("When a Build has failed", func() {
		It("successfully marks the CC v3 build as failed to have staged", func() {
			// TODO
		})
	})
})
