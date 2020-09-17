package controllers

import (
	"context"
	"errors"
	. "github.com/onsi/gomega/gstruct"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1alpha1 "code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/apis/apps.cloudfoundry.org/v1alpha1"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/model"
	networkingv1alpha1 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("RouteSyncController", func() {
	const (
		routeGUID = "route-guid"
	)

	findSyncCondition := func(conditions []appsv1alpha1.Condition) *appsv1alpha1.Condition {
		for _, c := range conditions {
			if c.Type == appsv1alpha1.SyncedConditionType {
				return &c
			}
		}
		return nil
	}
	createRouteInK8s := func() {
		port := 56789

		Expect(
			k8sClient.Create(context.Background(), &networkingv1alpha1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Name:      routeGUID,
					Namespace: workloadsNamespace,
				},
				Spec: networkingv1alpha1.RouteSpec{
					Host: "a-host",
					Path: "/path",
					Url:  "a-host.domain.com/path",
					Domain: networkingv1alpha1.RouteDomain{
						Name:     "domain.com",
						Internal: false,
					},
					Destinations: []networkingv1alpha1.RouteDestination{
						{
							Guid:   "dest-guid",
							Port:   &port,
							Weight: nil,
							App: networkingv1alpha1.DestinationApp{
								Guid: "app-guid",
								Process: networkingv1alpha1.AppProcess{
									Type: "web",
								},
							},
							Selector: networkingv1alpha1.DestinationSelector{
								MatchLabels: map[string]string{},
							},
						},
					},
				},
			}),
		).To(Succeed())

		var createdRouteResource networkingv1alpha1.Route
		Eventually(func() error {
			return k8sClient.Get(context.Background(), types.NamespacedName{Name: routeGUID, Namespace: workloadsNamespace}, &createdRouteResource)
		}, "5s", "1s").Should(Succeed())
	}

	When("there are routes in the CF API but not in Kubernetes", func() {
		const (
			appGUID    = "app-guid"
			destGUID   = "dest-guid"
			spaceGUID  = "space-guid"
			domainGUID = "domain-guid"
			orgGUID    = "org-guid"

			host        = "a-host"
			path        = "/path"
			url         = "a-host.domain.com/path"
			port        = 56789
			processType = "web"
			domainName  = "domain.com"
		)

		BeforeEach(func() {
			fakeCFClient.ListRoutesReturns([]model.Route{
				{
					GUID: routeGUID,
					Host: host,
					Path: path,
					URL:  url,
					Destinations: []model.Destination{
						{
							GUID: destGUID,
							Port: port,
							App: model.DestinationApp{
								GUID: appGUID,
								Process: model.DestinationProcess{
									Type: processType,
								},
							},
						},
					},
					Relationships: map[string]model.Relationship{
						"space": {
							Data: model.RelationshipData{
								GUID: spaceGUID,
							},
						},
						"domain": {
							Data: model.RelationshipData{
								GUID: domainGUID,
							},
						},
					},
				},
			}, nil)

			fakeCFClient.GetSpaceReturns(model.Space{
				GUID: spaceGUID,
				Relationships: map[string]model.Relationship{
					"organization": {
						Data: model.RelationshipData{
							GUID: orgGUID,
						},
					},
				},
			}, nil)

			fakeCFClient.GetDomainReturns(model.Domain{
				GUID:     domainGUID,
				Name:     domainName,
				Internal: false,
			}, nil)

			Expect(k8sClient.Create(context.Background(), &appsv1alpha1.RouteSync{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "whatever",
					Namespace: workloadsNamespace,
				},
				Spec: appsv1alpha1.RouteSyncSpec{
					PeriodSeconds: 1,
				},
			})).To(Succeed())
		})

		It("creates the missing route resources in Kubernetes", func() {
			var createdRouteResource networkingv1alpha1.Route
			Eventually(func() error {
				return k8sClient.Get(context.Background(), types.NamespacedName{Name: routeGUID, Namespace: workloadsNamespace}, &createdRouteResource)
			}, "5s", "1s").Should(Succeed())

			Expect(createdRouteResource.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/name", routeGUID))
			Expect(createdRouteResource.ObjectMeta.Labels).To(HaveKeyWithValue("cloudfoundry.org/org_guid", orgGUID))
			Expect(createdRouteResource.ObjectMeta.Labels).To(HaveKeyWithValue("cloudfoundry.org/space_guid", spaceGUID))
			Expect(createdRouteResource.ObjectMeta.Labels).To(HaveKeyWithValue("cloudfoundry.org/domain_guid", domainGUID))
			Expect(createdRouteResource.ObjectMeta.Labels).To(HaveKeyWithValue("cloudfoundry.org/route_guid", routeGUID))

			Expect(createdRouteResource.Spec.Host).To(Equal(host))
			Expect(createdRouteResource.Spec.Path).To(Equal(path))
			Expect(createdRouteResource.Spec.Url).To(Equal(url))

			Expect(createdRouteResource.Spec.Domain.Name).To(Equal(domainName))
			Expect(createdRouteResource.Spec.Domain.Internal).To(BeFalse())

			Expect(createdRouteResource.Spec.Destinations).To(HaveLen(1))
			Expect(createdRouteResource.Spec.Destinations[0].Guid).To(Equal(destGUID))
			Expect(createdRouteResource.Spec.Destinations[0].Port).To(PointTo(Equal(port)))
			Expect(createdRouteResource.Spec.Destinations[0].App.Guid).To(Equal(appGUID))
			Expect(createdRouteResource.Spec.Destinations[0].App.Process.Type).To(Equal(processType))

			Expect(createdRouteResource.Spec.Destinations[0].Selector.MatchLabels).To(HaveLen(2))
			Expect(createdRouteResource.Spec.Destinations[0].Selector.MatchLabels).To(HaveKeyWithValue("cloudfoundry.org/app_guid", appGUID))
			Expect(createdRouteResource.Spec.Destinations[0].Selector.MatchLabels).To(HaveKeyWithValue("cloudfoundry.org/process_type", processType))
		})
	})

	When("there are routes in Kubernetes that aren't in the CF API", func() {

		BeforeEach(func() {
			fakeCFClient.ListRoutesReturns([]model.Route{}, nil)

			createRouteInK8s()

			Expect(
				k8sClient.Create(context.Background(), &appsv1alpha1.RouteSync{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "whatever",
						Namespace: workloadsNamespace,
					},
					Spec: appsv1alpha1.RouteSyncSpec{
						PeriodSeconds: 1,
					},
				}),
			).To(Succeed())
		})

		It("deletes the extra Route from kubernetes", func() {
			Eventually(func() error {
				return k8sClient.Get(context.Background(), types.NamespacedName{Name: routeGUID, Namespace: workloadsNamespace}, new(networkingv1alpha1.Route))
			}, "5s", "1s").Should(MatchError(ContainSubstring("not found")))
		})
	})

	Describe("resync interval", func() {
		BeforeEach(func() {
			fakeCFClient.ListRoutesReturns([]model.Route{}, nil)

			createRouteInK8s()

			Expect(
				k8sClient.Create(context.Background(), &appsv1alpha1.RouteSync{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "whatever",
						Namespace: workloadsNamespace,
					},
					Spec: appsv1alpha1.RouteSyncSpec{
						PeriodSeconds: 1,
					},
				}),
			).To(Succeed())
		})

		It("synchronizes at the frequency set on the RouteSync resource", func() {
			Eventually(func() error {
				return k8sClient.Get(context.Background(), types.NamespacedName{Name: routeGUID, Namespace: workloadsNamespace}, new(networkingv1alpha1.Route))
			}, "5s", "1s").Should(MatchError(ContainSubstring("not found")))

			createRouteInK8s()

			Eventually(func() error {
				return k8sClient.Get(context.Background(), types.NamespacedName{Name: routeGUID, Namespace: workloadsNamespace}, new(networkingv1alpha1.Route))
			}, "5s", "1s").Should(MatchError(ContainSubstring("not found")))
		})
	})

	Describe("RouteSync Status", func() {
		Context("when the sync succeeds", func() {
			const (
				routeSyncName = "my-route-sync"
			)
			var (
				testStartTime metav1.Time
			)

			BeforeEach(func() {
				testStartTime = metav1.Now()
				fakeCFClient.ListRoutesReturns([]model.Route{}, nil)

				createRouteInK8s()

				Expect(
					k8sClient.Create(context.Background(), &appsv1alpha1.RouteSync{
						ObjectMeta: metav1.ObjectMeta{
							Name:      routeSyncName,
							Namespace: workloadsNamespace,
						},
						Spec: appsv1alpha1.RouteSyncSpec{
							PeriodSeconds: 1,
						},
					}),
				).To(Succeed())
			})

			It("updates the Synced condition on the RouteSync's Status", func() {
				routeSync := appsv1alpha1.RouteSync{}
				Eventually(func() error {
					return k8sClient.Get(context.Background(), types.NamespacedName{Name: routeSyncName, Namespace: workloadsNamespace}, &routeSync)
				}, "5s", "1s").Should(Succeed())

				var syncCondition *appsv1alpha1.Condition
				Eventually(func() *appsv1alpha1.Condition {
					err := k8sClient.Get(context.Background(), types.NamespacedName{Name: routeSyncName, Namespace: workloadsNamespace}, &routeSync)
					Expect(err).NotTo(HaveOccurred())
					syncCondition = findSyncCondition(routeSync.Status.Conditions)
					return syncCondition
				}, "5s", "1s").ShouldNot(BeNil())

				Expect(syncCondition.Status).To(Equal(appsv1alpha1.TrueConditionStatus))
				Expect(syncCondition.Reason).To(Equal(appsv1alpha1.CompletedConditionReason))
				Expect(testStartTime.Unix()).To(BeNumerically("~", syncCondition.LastTransitionTime.Unix(), 1))
			})
		})

		Context("when the sync fails", func() {
			const (
				routeSyncName = "my-route-sync"
			)
			var (
				testStartTime metav1.Time
				errMsg        = "error fetching routes o no"
			)

			BeforeEach(func() {
				testStartTime = metav1.Now()
				fakeCFClient.ListRoutesReturns([]model.Route{}, errors.New(errMsg))

				createRouteInK8s()

				Expect(
					k8sClient.Create(context.Background(), &appsv1alpha1.RouteSync{
						ObjectMeta: metav1.ObjectMeta{
							Name:      routeSyncName,
							Namespace: workloadsNamespace,
						},
						Spec: appsv1alpha1.RouteSyncSpec{
							PeriodSeconds: 1,
						},
					}),
				).To(Succeed())
			})

			It("updates the Synced condition on the RouteSync's Status", func() {
				routeSync := appsv1alpha1.RouteSync{}
				Eventually(func() error {
					return k8sClient.Get(context.Background(), types.NamespacedName{Name: routeSyncName, Namespace: workloadsNamespace}, &routeSync)
				}, "5s", "1s").Should(Succeed())

				var syncCondition *appsv1alpha1.Condition
				Eventually(func() *appsv1alpha1.Condition {
					err := k8sClient.Get(context.Background(), types.NamespacedName{Name: routeSyncName, Namespace: workloadsNamespace}, &routeSync)
					Expect(err).NotTo(HaveOccurred())
					syncCondition = findSyncCondition(routeSync.Status.Conditions)
					return syncCondition
				}, "5s", "1s").ShouldNot(BeNil())

				Expect(syncCondition.Status).To(Equal(appsv1alpha1.FalseConditionStatus))
				Expect(syncCondition.Reason).To(Equal(appsv1alpha1.FailedConditionReason))
				Expect(syncCondition.Message).To(Equal(errMsg))
				Expect(testStartTime.Unix()).To(BeNumerically("~", syncCondition.LastTransitionTime.Unix(), 5))
			})
		})
	})
})
