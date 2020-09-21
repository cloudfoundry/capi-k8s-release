package units_test

import (
	appsv1alpha1 "code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/apis/apps.cloudfoundry.org/v1alpha1"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/cffakes"
	"context"
	"time"

	. "github.com/onsi/gomega"

	logrTesting "github.com/go-logr/logr/testing"

	. "code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/controllers"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/controllers/fake"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("PeriodicSyncController", func() {
	Describe("Reconcile", func() {
		When("a the sync succeeds", func() {
			var (
				reconciler *PeriodicSyncReconciler
				client     *fake.ControllerRuntimeClient
				logger     logr.Logger
				cfClient   *cffakes.FakeClientInterface
				request    ctrl.Request
			)

			const (
				workloadsNamespace = "cf-workloads-fake"
				periodicSyncName   = "some-periodic-sync"
				syncPeriodSeconds  = 5
			)

			BeforeEach(func() {
				client = new(fake.ControllerRuntimeClient)
				cfClient = new(cffakes.FakeClientInterface)

				logger = logrTesting.NullLogger{}

				reconciler = &PeriodicSyncReconciler{
					Client:             client,
					Log:                logger,
					Scheme:             nil,
					CFClient:           cfClient,
					WorkloadsNamespace: workloadsNamespace,
				}
				request = ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: workloadsNamespace,
						Name:      periodicSyncName,
					},
				}
				periodicSync := appsv1alpha1.PeriodicSync{
					Spec: appsv1alpha1.PeriodicSyncSpec{
						PeriodSeconds: syncPeriodSeconds,
					},
					Status: appsv1alpha1.PeriodicSyncStatus{},
				}

				client.GetCalls(func(ctx context.Context, name types.NamespacedName, object runtime.Object) error {
					ptr := object.(*appsv1alpha1.PeriodicSync)
					*ptr = periodicSync
					return nil
				})

				client.StatusReturns(client)
			})

			It("sync routes and requeues on the specified duration", func() {
				result, err := reconciler.Reconcile(request)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{RequeueAfter: syncPeriodSeconds * time.Second}))
			})
		})
	})
})
