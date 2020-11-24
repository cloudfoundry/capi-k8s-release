package controllers

import (
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/kubernetes"
	"context"
	. "github.com/onsi/gomega/gstruct"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1alpha1 "code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/apis/apps.cloudfoundry.org/v1alpha1"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/model"
	networkingv1alpha1 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("PeriodicSyncController", func() {
	var (
		port = 56789
	)

	const (
		routeGUID       = "route-guid"
		secondRouteGUID = "route-guid-2"
		appGUID         = "app-guid"
		destGUID        = "dest-guid"
		spaceGUID       = "space-guid"
		domainGUID      = "domain-guid"
		orgGUID         = "org-guid"

		host        = "a-host"
		path        = "/path"
		url         = "a-host.domain.com/path"
		processType = "web"
		domainName  = "domain.com"
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
		Expect(
			k8sClient.Create(context.Background(), &networkingv1alpha1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Name:      routeGUID,
					Namespace: workloadsNamespace,
					Labels: map[string]string{
						kubernetes.KubeManagedByLabel: "cloudfoundry",
					},
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
		}, "5s", "100ms").Should(Succeed())
	}

	When("there are routes in the CF API but not in Kubernetes", func() {
		BeforeEach(func() {
			fakeCFClient.ListRoutesReturns(model.RouteList{
				Resources: []model.Route{
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
					{
						GUID: secondRouteGUID,
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
				},
				Included: model.RouteListIncluded{
					Spaces: []model.Space{
						{
							GUID: spaceGUID,
							Relationships: map[string]model.Relationship{
								"organization": {
									Data: model.RelationshipData{
										GUID: orgGUID,
									},
								},
							},
						},
					},
					Domains: []model.Domain{
						{
							GUID:     domainGUID,
							Name:     domainName,
							Internal: false,
						},
					},
				},
			}, nil)

			Expect(k8sClient.Create(context.Background(), &appsv1alpha1.PeriodicSync{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "whatever",
					Namespace: workloadsNamespace,
				},
				Spec: appsv1alpha1.PeriodicSyncSpec{
					PeriodSeconds: 1,
				},
			})).To(Succeed())
		})

		It("creates the missing route resources in Kubernetes", func() {
			var createdRouteResource networkingv1alpha1.Route
			Eventually(func() error {
				return k8sClient.Get(context.Background(), types.NamespacedName{Name: routeGUID, Namespace: workloadsNamespace}, &createdRouteResource)
			}, "5s", "1s").Should(Succeed())

			Expect(createdRouteResource.ObjectMeta.Name).To(Equal(routeGUID))
			Expect(createdRouteResource.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/name", routeGUID))
			Expect(createdRouteResource.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", "cloudfoundry"))
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

			var secondRouteResource networkingv1alpha1.Route
			Eventually(func() error {
				return k8sClient.Get(context.Background(), types.NamespacedName{Name: secondRouteGUID, Namespace: workloadsNamespace}, &secondRouteResource)
			}, "5s", "1s").Should(Succeed())
			Expect(secondRouteResource.ObjectMeta.Name).To(Equal(secondRouteGUID))

		})
	})

	When("there are routes in Kubernetes that aren't in the CF API", func() {

		BeforeEach(func() {
			fakeCFClient.ListRoutesReturns(model.RouteList{}, nil)

			createRouteInK8s()

			Expect(
				k8sClient.Create(context.Background(), &appsv1alpha1.PeriodicSync{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "whatever",
						Namespace: workloadsNamespace,
					},
					Spec: appsv1alpha1.PeriodicSyncSpec{
						PeriodSeconds: 1,
					},
				}),
			).To(Succeed())
		})

		It("deletes the extra Route from kubernetes", func() {
			Eventually(func() error {
				return k8sClient.Get(context.Background(), types.NamespacedName{Name: routeGUID, Namespace: workloadsNamespace}, new(networkingv1alpha1.Route))
			}, "10s", "1s").Should(MatchError(ContainSubstring("not found")))
		})
	})

	When("the destinations in the Kubernetes Routes don't match those in the CF API", func() {

		BeforeEach(func() {
			fakeCFClient.ListRoutesReturns(model.RouteList{
				Resources: []model.Route{
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
					{
						GUID: secondRouteGUID,
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
				},
				Included: model.RouteListIncluded{
					Spaces: []model.Space{
						{
							GUID: spaceGUID,
							Relationships: map[string]model.Relationship{
								"organization": {
									Data: model.RelationshipData{
										GUID: orgGUID,
									},
								},
							},
						},
					},
					Domains: []model.Domain{
						{
							GUID:     domainGUID,
							Name:     domainName,
							Internal: false,
						},
					},
				},
			}, nil)

			Expect(
				k8sClient.Create(context.Background(), &networkingv1alpha1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Name:      routeGUID,
						Namespace: workloadsNamespace,
						Labels: map[string]string{
							kubernetes.KubeManagedByLabel: "cloudfoundry",
						},
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
									Guid: "wrong-app-guid",
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

			Expect(
				k8sClient.Create(context.Background(), &appsv1alpha1.PeriodicSync{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "whatever",
						Namespace: workloadsNamespace,
					},
					Spec: appsv1alpha1.PeriodicSyncSpec{
						PeriodSeconds: 1,
					},
				}),
			).To(Succeed())
		})

		It("updates the destination on Kubernetes", func() {
			var updatedRouteResource networkingv1alpha1.Route
			Eventually(func() string {
				k8sClient.Get(context.Background(), types.NamespacedName{Name: routeGUID, Namespace: workloadsNamespace}, &updatedRouteResource)
				return updatedRouteResource.Spec.Destinations[0].App.Guid
			}, "5s", "1s").Should(Equal(appGUID))

			Expect(updatedRouteResource.Spec.Destinations).To(HaveLen(1))
			Expect(updatedRouteResource.Spec.Destinations[0].Guid).To(Equal(destGUID))
			Expect(updatedRouteResource.Spec.Destinations[0].Port).To(PointTo(Equal(port)))
			Expect(updatedRouteResource.Spec.Destinations[0].App.Guid).To(Equal(appGUID))
			Expect(updatedRouteResource.Spec.Destinations[0].App.Process.Type).To(Equal(processType))

			Expect(updatedRouteResource.Spec.Destinations[0].Selector.MatchLabels).To(HaveLen(2))
			Expect(updatedRouteResource.Spec.Destinations[0].Selector.MatchLabels).To(HaveKeyWithValue("cloudfoundry.org/app_guid", appGUID))
			Expect(updatedRouteResource.Spec.Destinations[0].Selector.MatchLabels).To(HaveKeyWithValue("cloudfoundry.org/process_type", processType))
		})
	})

	When("there are routes not managed by Cloud Foundry", func() {
		BeforeEach(func() {
			fakeCFClient.ListRoutesReturns(model.RouteList{}, nil)

			Expect(
				k8sClient.Create(context.Background(), &networkingv1alpha1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Name:      routeGUID,
						Namespace: workloadsNamespace,
						Labels: map[string]string{
							kubernetes.KubeManagedByLabel: "some-other-platform",
						},
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
									Guid: "wrong-app-guid",
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

			Expect(
				k8sClient.Create(context.Background(), &appsv1alpha1.PeriodicSync{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "whatever",
						Namespace: workloadsNamespace,
					},
					Spec: appsv1alpha1.PeriodicSyncSpec{
						PeriodSeconds: 1,
					},
				}),
			).To(Succeed())
		})

		It("does NOT delete the extra Route from kubernetes", func() {
			var routeResource networkingv1alpha1.Route
			Consistently(func() error {
				return k8sClient.Get(context.Background(), types.NamespacedName{Name: routeGUID, Namespace: workloadsNamespace}, &routeResource)
			}, "5s", "1s").Should(Succeed())
		})
	})

	Describe("PeriodicSync Status", func() {
		Context("when the sync succeeds", func() {
			const (
				periodicSyncName = "my-route-sync"
			)
			var (
				testStartTime metav1.Time
			)

			BeforeEach(func() {
				testStartTime = metav1.Now()
				fakeCFClient.ListRoutesReturns(model.RouteList{}, nil)

				createRouteInK8s()

				Expect(
					k8sClient.Create(context.Background(), &appsv1alpha1.PeriodicSync{
						ObjectMeta: metav1.ObjectMeta{
							Name:      periodicSyncName,
							Namespace: workloadsNamespace,
						},
						Spec: appsv1alpha1.PeriodicSyncSpec{
							PeriodSeconds: 1,
						},
					}),
				).To(Succeed())
			})

			It("updates the Synced condition on the PeriodicSync's Status", func() {
				periodicSync := appsv1alpha1.PeriodicSync{}
				Eventually(func() error {
					return k8sClient.Get(context.Background(), types.NamespacedName{Name: periodicSyncName, Namespace: workloadsNamespace}, &periodicSync)
				}, "5s", "1s").Should(Succeed())

				var syncCondition *appsv1alpha1.Condition
				Eventually(func() *appsv1alpha1.Condition {
					err := k8sClient.Get(context.Background(), types.NamespacedName{Name: periodicSyncName, Namespace: workloadsNamespace}, &periodicSync)
					Expect(err).NotTo(HaveOccurred())
					syncCondition = findSyncCondition(periodicSync.Status.Conditions)
					return syncCondition
				}, "5s", "1s").ShouldNot(BeNil())

				Expect(syncCondition.Status).To(Equal(appsv1alpha1.TrueConditionStatus))
				Expect(syncCondition.Reason).To(Equal(appsv1alpha1.CompletedConditionReason))
				Expect(testStartTime.Unix()).To(BeNumerically("~", syncCondition.LastTransitionTime.Unix(), 5))
			})
		})
	})
})
