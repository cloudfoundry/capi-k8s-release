package units_test

import (
	"errors"

	"context"
	"time"

	appsv1alpha1 "code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/apis/apps.cloudfoundry.org/v1alpha1"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/cffakes"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/model"
	networkingv1alpha1 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"

	logrTesting "github.com/go-logr/logr/testing"

	. "code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/controllers"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/controllers/fake"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("PeriodicSyncController", func() {
	const (
		workloadsNamespace = "cf-workloads-fake"
		periodicSyncName   = "some-periodic-sync"
		syncPeriodSeconds  = 5
	)

	Describe("Reconcile", func() {
		var (
			reconciler *PeriodicSyncReconciler
			client     *fake.ControllerRuntimeClient
			logger     logr.Logger
			cfClient   *cffakes.FakeClientInterface
			request    ctrl.Request
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

			client.StatusReturns(client)

			client.GetCalls(func(ctx context.Context, name types.NamespacedName, object runtime.Object) error {
				ptr := object.(*appsv1alpha1.PeriodicSync)
				*ptr = periodicSync
				return nil
			})
		})

		Context("the sync succeeds", func() {
			It("sync routes and requeues on the specified duration", func() {
				result, err := reconciler.Reconcile(request)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{RequeueAfter: syncPeriodSeconds * time.Second}))
			})
		})

		Context("when it fails to fetch the period sync", func() {
			var (
				errMsg = "error fetching periodic sync"
			)

			BeforeEach(func() {
				client.GetReturns(errors.New(errMsg))
			})

			It("updates the Synced condition on the PeriodicSync's Status", func() {
				_, err := reconciler.Reconcile(request)
				Expect(err).To(MatchError(ContainSubstring(errMsg)))

				_, syncObject, _ := client.UpdateArgsForCall(0)
				conditions := syncObject.(*appsv1alpha1.PeriodicSync).Status.Conditions
				Expect(conditions).To(ConsistOf(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Status":  Equal(appsv1alpha1.FalseConditionStatus),
					"Reason":  Equal(appsv1alpha1.FailedConditionReason),
					"Message": Equal(errMsg),
				})))
			})
		})

		Context("when it fails to fetch cf routes", func() {
			var (
				errMsg = "error fetching cf routes o no"
			)

			BeforeEach(func() {
				cfClient.ListRoutesReturns(model.RouteList{}, errors.New(errMsg))
			})

			It("updates the Synced condition on the PeriodicSync's Status", func() {
				_, err := reconciler.Reconcile(request)
				Expect(err).To(MatchError(ContainSubstring(errMsg)))

				_, syncObject, _ := client.UpdateArgsForCall(0)
				conditions := syncObject.(*appsv1alpha1.PeriodicSync).Status.Conditions
				Expect(conditions).To(ConsistOf(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Status":  Equal(appsv1alpha1.FalseConditionStatus),
					"Reason":  Equal(appsv1alpha1.FailedConditionReason),
					"Message": Equal(errMsg),
				})))
			})
		})

		Context("when it fails fetch k8s routes", func() {
			var (
				errMsg = "error fetching k8s routes o no"
			)

			BeforeEach(func() {
				client.ListReturns(errors.New(errMsg))
			})

			It("updates the Synced condition on the PeriodicSync's Status", func() {
				_, err := reconciler.Reconcile(request)
				Expect(err).To(MatchError(ContainSubstring(errMsg)))

				_, syncObject, _ := client.UpdateArgsForCall(0)
				conditions := syncObject.(*appsv1alpha1.PeriodicSync).Status.Conditions
				Expect(conditions).To(ConsistOf(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Status":  Equal(appsv1alpha1.FalseConditionStatus),
					"Reason":  Equal(appsv1alpha1.FailedConditionReason),
					"Message": Equal(errMsg),
				})))
			})
		})

		Context("when it fails to create routes", func() {
			var (
				errMsg = "error creating k8s route o no"
			)
			BeforeEach(func() {
				cfClient.ListRoutesReturns(model.RouteList{
					Resources: []model.Route{{
						Destinations: []model.Destination{},
						Relationships: map[string]model.Relationship{
							"space": {
								Data: model.RelationshipData{},
							},
							"domain": {
								Data: model.RelationshipData{},
							},
						},
					}},
					Included: model.RouteListIncluded{
						Spaces: []model.Space{
							{
								Relationships: map[string]model.Relationship{
									"organization": {
										Data: model.RelationshipData{},
									},
								},
							},
						},
						Domains: []model.Domain{{}},
					},
				}, nil)

				client.CreateReturns(errors.New(errMsg))
			})

			It("updates the Synced condition on the PeriodicSync's Status", func() {
				_, err := reconciler.Reconcile(request)
				Expect(err).To(MatchError("failed to reconcile at least one route"))

				_, syncObject, _ := client.UpdateArgsForCall(0)
				conditions := syncObject.(*appsv1alpha1.PeriodicSync).Status.Conditions
				Expect(conditions).To(ConsistOf(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Status":  Equal(appsv1alpha1.FalseConditionStatus),
					"Reason":  Equal(appsv1alpha1.FailedConditionReason),
					"Message": Equal("failed to reconcile at least one route"),
				})))
			})
		})

		Context("when it fails to delete routes", func() {
			var (
				errMsg = "error deleting k8s route o no"
			)
			BeforeEach(func() {
				client.ListCalls(func(_ context.Context, object runtime.Object, _ ...ctrlClient.ListOption) error {
					ptr := object.(*networkingv1alpha1.RouteList)
					*ptr = networkingv1alpha1.RouteList{
						Items: []networkingv1alpha1.Route{{}},
					}
					return nil
				})

				client.DeleteReturns(errors.New(errMsg))
			})

			It("updates the Synced condition on the PeriodicSync's Status", func() {
				_, err := reconciler.Reconcile(request)
				Expect(err).To(MatchError("failed to reconcile at least one route"))

				_, syncObject, _ := client.UpdateArgsForCall(0)
				conditions := syncObject.(*appsv1alpha1.PeriodicSync).Status.Conditions
				Expect(conditions).To(ConsistOf(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Status":  Equal(appsv1alpha1.FalseConditionStatus),
					"Reason":  Equal(appsv1alpha1.FailedConditionReason),
					"Message": Equal("failed to reconcile at least one route"),
				})))
			})
		})
	})
})
