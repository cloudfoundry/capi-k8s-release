package controllers_test

import (
	"context"

	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/model"

	. "github.com/onsi/gomega"

	"github.com/buildpacks/lifecycle"

	v1 "github.com/google/go-containerregistry/pkg/v1"

	logrTesting "github.com/go-logr/logr/testing"

	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/controllers"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/controllers/fake"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/image_registry/image_registryfakes"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	buildv1alpha1 "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("BuildReconciler", func() {
	Describe("Reconcile", func() {
		When("a build succeeds", func() {
			var (
				reconciler         *BuildReconciler
				client             *fake.ControllerRuntimeClient
				logger             logr.Logger
				cfBuildUpdater     *fake.CFBuildUpdater
				imageConfigFetcher *image_registryfakes.FakeImageConfigFetcher
				request            ctrl.Request
				build              buildv1alpha1.Build
			)

			const (
				buildNamespace = "cf-workloads-staging"
				buildName      = "some-build"
				buildGUID      = "build-guid"
				latestImage    = "theLatestImage"
			)

			BeforeEach(func() {
				client = new(fake.ControllerRuntimeClient)
				cfBuildUpdater = new(fake.CFBuildUpdater)
				imageConfigFetcher = new(image_registryfakes.FakeImageConfigFetcher)

				logger = logrTesting.NullLogger{}

				reconciler = &BuildReconciler{
					Client:             client,
					CFClient:           cfBuildUpdater,
					ImageConfigFetcher: imageConfigFetcher,
					Log:                logger,
					Scheme:             nil,
				}
				request = ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: buildNamespace,
						Name:      buildName,
					},
				}
				build = buildv1alpha1.Build{
					ObjectMeta: metav1.ObjectMeta{
						Name:      buildName,
						Namespace: buildNamespace,
						Labels: map[string]string{
							BuildGUIDLabel: buildGUID,
						},
					},
					TypeMeta: metav1.TypeMeta{
						Kind: "Build",
					},
					Spec: buildv1alpha1.BuildSpec{
						ServiceAccount: "serviceAccount",
					},
					Status: buildv1alpha1.BuildStatus{
						LatestImage: latestImage,
						Status: corev1alpha1.Status{
							Conditions: corev1alpha1.Conditions{
								{Type: "Succeeded", Status: "True"},
							},
						},
					},
				}

				client.GetCalls(func(ctx context.Context, name types.NamespacedName, object runtime.Object) error {
					ptr := object.(*buildv1alpha1.Build)
					*ptr = build
					return nil
				})

				imageConfig := v1.Config{
					Labels: map[string]string{
						lifecycle.BuildMetadataLabel: `{"processes": [{"type": "web", "command": "rackup"}]}`,
					},
				}
				imageConfigFetcher.FetchImageConfigReturns(&imageConfig, nil)
			})

			It("updates the build in CC", func() {
				result, err := reconciler.Reconcile(request)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))

				Expect(cfBuildUpdater.UpdateBuildCallCount()).To(Equal(1))

				actualBuildGUID, updateBuildRequest := cfBuildUpdater.UpdateBuildArgsForCall(0)
				Expect(actualBuildGUID).To(Equal(buildGUID))
				Expect(updateBuildRequest).To(Equal(model.Build{
					State: "STAGED",
					Lifecycle: model.Lifecycle{
						Type: "kpack",
						Data: model.LifecycleData{
							Image: latestImage,
							ProcessTypes: map[string]string{
								"web": "rackup",
							},
						},
					},
				}))
			})
		})
	})
})
