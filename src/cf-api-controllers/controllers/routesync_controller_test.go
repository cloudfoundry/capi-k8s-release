package controllers

import (
	"context"

	. "github.com/onsi/gomega/gstruct"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

	createRouteInK8s := func() {
		port := 56789

		Expect(
			k8sClient.Create(context.Background(), &networkingv1alpha1.Route{
				ObjectMeta: v1.ObjectMeta{
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
							Guid: "dest-guid",
							Port: &port,
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
				ObjectMeta: v1.ObjectMeta{
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
					ObjectMeta: v1.ObjectMeta{
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
					ObjectMeta: v1.ObjectMeta{
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
})
