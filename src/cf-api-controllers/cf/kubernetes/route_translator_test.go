package kubernetes_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/kubernetes"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/model"
	"code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"
)

var _ = Describe("RouteTranslator", func() {
	var (
		route   model.Route
		space   model.Space
		domain  model.Domain
		routeCR v1alpha1.Route
	)

	const (
		namespace = "some-dope-space"
	)

	Describe("TranslateRoute", func() {
		Context("given a CF API route object with all required fields", func() {

			intPtr := func(x int) *int {
				return &x
			}

			BeforeEach(func() {
				route = model.Route{
					GUID: "route-guid",
					Host: "host",
					Path: "/path",
					URL:  "host.domain.com/path",
					Destinations: []model.Destination{
						{
							GUID:   "destination-guid",
							Port:   8080,
							Weight: intPtr(100),
							App: model.DestinationApp{
								GUID: "app-guid",
								Process: model.DestinationProcess{
									Type: "web",
								},
							},
						},
					},
					Relationships: map[string]model.Relationship{
						"space": {
							Data: model.RelationshipData{GUID: "space-guid"},
						},
						"domain": {
							Data: model.RelationshipData{GUID: "domain-guid"},
						},
					},
				}
				space = model.Space{
					GUID: "space-guid",
					Relationships: map[string]model.Relationship{
						"organization": {
							Data: model.RelationshipData{GUID: "org-guid"},
						},
					},
				}
				domain = model.Domain{
					GUID:     "domain-guid",
					Name:     "domain.com",
					Internal: false,
				}
			})

			It("returns a valid Route CR", func() {
				routeCR = TranslateRoute(&route, &space, &domain, namespace)

				Expect(routeCR).NotTo(BeNil())

				Expect(routeCR.ObjectMeta.Name).To(Equal(route.GUID))
				Expect(routeCR.ObjectMeta.Namespace).To(Equal(namespace))
				Expect(routeCR.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/name", route.GUID))
				Expect(routeCR.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/version", "0.0.0"))
				Expect(routeCR.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", "cloudfoundry"))
				Expect(routeCR.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/component", "cf-networking"))
				Expect(routeCR.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/part-of", "cloudfoundry"))
				Expect(routeCR.ObjectMeta.Labels).To(HaveKeyWithValue("cloudfoundry.org/org_guid", space.Relationships["organization"].Data.GUID))
				Expect(routeCR.ObjectMeta.Labels).To(HaveKeyWithValue("cloudfoundry.org/space_guid", space.GUID))
				Expect(routeCR.ObjectMeta.Labels).To(HaveKeyWithValue("cloudfoundry.org/domain_guid", domain.GUID))
				Expect(routeCR.ObjectMeta.Labels).To(HaveKeyWithValue("cloudfoundry.org/route_guid", route.GUID))

				Expect(routeCR.Spec.Host).To(Equal(route.Host))
				Expect(routeCR.Spec.Path).To(Equal(route.Path))
				Expect(routeCR.Spec.Url).To(Equal(route.URL))

				Expect(routeCR.Spec.Domain.Name).To(Equal(domain.Name))
				Expect(routeCR.Spec.Domain.Internal).To(Equal(domain.Internal))

				Expect(routeCR.Spec.Destinations).To(HaveLen(1))
				Expect(routeCR.Spec.Destinations[0].Guid).To(Equal(route.Destinations[0].GUID))
				Expect(routeCR.Spec.Destinations[0].Port).To(gstruct.PointTo(Equal(8080)))
				Expect(routeCR.Spec.Destinations[0].Weight).To(gstruct.PointTo(Equal(100)))
				Expect(routeCR.Spec.Destinations[0].App.Guid).To(Equal(route.Destinations[0].App.GUID))
				Expect(routeCR.Spec.Destinations[0].App.Process.Type).To(Equal(route.Destinations[0].App.Process.Type))

				Expect(routeCR.Spec.Destinations[0].Selector.MatchLabels).To(HaveLen(2))
				Expect(routeCR.Spec.Destinations[0].Selector.MatchLabels).To(HaveKeyWithValue("cloudfoundry.org/app_guid", route.Destinations[0].App.GUID))
				Expect(routeCR.Spec.Destinations[0].Selector.MatchLabels).To(HaveKeyWithValue("cloudfoundry.org/process_type", route.Destinations[0].App.Process.Type))
			})
		})
	})

	Describe("CompareRoutes", func() {
		var (
			desiredRoute v1alpha1.Route
			actualRoute v1alpha1.Route
		)
		BeforeEach(func() {
			desiredRoute = v1alpha1.Route{
				TypeMeta: v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{
					Name:      route.GUID,
					Namespace: namespace,
					Labels: map[string]string{
						KubeNameLabel:      route.GUID,
						KubeVersionLabel:   "0.0.0",
						KubeManagedByLabel: "cloudfoundry",
						KubeComponentLabel: "cf-networking",
						KubePartOfLabel:    "cloudfoundry",
						CFOrgGuidLabel:     space.Relationships["organization"].Data.GUID,
						CFSpaceGuidLabel:   space.GUID,
						CFDomainGuidLabel:  domain.GUID,
						CFRouteGuidLabel:   route.GUID,
					},
				},
				Spec: v1alpha1.RouteSpec{
					Host: route.Host,
					Path: route.Path,
					Url:  route.URL,
					Domain: v1alpha1.RouteDomain{
						Name:     domain.Name,
						Internal: domain.Internal,
					},
					Destinations: []v1alpha1.RouteDestination{
						{
							Guid: "route-destination-guid-1",
							Weight:   intToPtr(80),
							Port:     intToPtr(9000),
							App:      v1alpha1.DestinationApp{
								Guid:    "app-guid-1",
								Process: v1alpha1.AppProcess{
									Type: "web",
								},
							},
							Selector: v1alpha1.DestinationSelector{
								MatchLabels: map[string]string{
									"cloudfoundry.org/app_guid": "app-guid-1",
									"cloudfoundry.org/process_type": "web",
								},
							},
						},
						{
							Guid: "route-destination-guid-2",
							Weight:   intToPtr(20),
							Port:     intToPtr(9000),
							App:      v1alpha1.DestinationApp{
								Guid:    "app-guid-2",
								Process: v1alpha1.AppProcess{
									Type: "web",
								},
							},
							Selector: v1alpha1.DestinationSelector{
								MatchLabels: map[string]string{
									"cloudfoundry.org/app_guid": "app-guid-2",
									"cloudfoundry.org/process_type": "web",
								},
							},
						},
					},
				},
				Status: v1alpha1.RouteStatus{},
			}
		})

		Context("when the Routes have the same destinations", func() {
			BeforeEach(func() {
				actualRoute = v1alpha1.Route{
					TypeMeta: v1.TypeMeta{},
					ObjectMeta: v1.ObjectMeta{
						Name:      route.GUID,
						Namespace: namespace,
						Labels: map[string]string{
							KubeNameLabel:      route.GUID,
							KubeVersionLabel:   "0.0.0",
							KubeManagedByLabel: "cloudfoundry",
							KubeComponentLabel: "cf-networking",
							KubePartOfLabel:    "cloudfoundry",
							CFOrgGuidLabel:     space.Relationships["organization"].Data.GUID,
							CFSpaceGuidLabel:   space.GUID,
							CFDomainGuidLabel:  domain.GUID,
							CFRouteGuidLabel:   route.GUID,
						},
						CreationTimestamp: v1.Now(),
						ResourceVersion: "fake-resource-version",
					},
					Spec: v1alpha1.RouteSpec{
						Host: route.Host,
						Path: route.Path,
						Url:  route.URL,
						Domain: v1alpha1.RouteDomain{
							Name:     domain.Name,
							Internal: domain.Internal,
						},
						Destinations: []v1alpha1.RouteDestination{
							{
								Guid: "route-destination-guid-1",
								Weight:   intToPtr(80),
								Port:     intToPtr(9000),
								App:      v1alpha1.DestinationApp{
									Guid:    "app-guid-1",
									Process: v1alpha1.AppProcess{
										Type: "web",
									},
								},
								Selector: v1alpha1.DestinationSelector{
									MatchLabels: map[string]string{
										"cloudfoundry.org/app_guid": "app-guid-1",
										"cloudfoundry.org/process_type": "web",
									},
								},
							},
							{
								Guid: "route-destination-guid-2",
								Weight:   intToPtr(20),
								Port:     intToPtr(9000),
								App:      v1alpha1.DestinationApp{
									Guid:    "app-guid-2",
									Process: v1alpha1.AppProcess{
										Type: "web",
									},
								},
								Selector: v1alpha1.DestinationSelector{
									MatchLabels: map[string]string{
										"cloudfoundry.org/app_guid": "app-guid-2",
										"cloudfoundry.org/process_type": "web",
									},
								},
							},
						},
					},
					Status: v1alpha1.RouteStatus{},
				}
			})

			It("returns true", func() {
				Expect(CompareRoutes(desiredRoute, actualRoute)).To(BeTrue())
			})
		})

		Context("when the actual Route is missing destinations", func() {
			BeforeEach(func() {
				actualRoute = v1alpha1.Route{
					TypeMeta: v1.TypeMeta{},
					ObjectMeta: v1.ObjectMeta{
						Name:      route.GUID,
						Namespace: namespace,
						Labels: map[string]string{
							KubeNameLabel:      route.GUID,
							KubeVersionLabel:   "0.0.0",
							KubeManagedByLabel: "cloudfoundry",
							KubeComponentLabel: "cf-networking",
							KubePartOfLabel:    "cloudfoundry",
							CFOrgGuidLabel:     space.Relationships["organization"].Data.GUID,
							CFSpaceGuidLabel:   space.GUID,
							CFDomainGuidLabel:  domain.GUID,
							CFRouteGuidLabel:   route.GUID,
						},
						CreationTimestamp: v1.Now(),
						ResourceVersion: "fake-resource-version",
					},
					Spec: v1alpha1.RouteSpec{
						Host: route.Host,
						Path: route.Path,
						Url:  route.URL,
						Domain: v1alpha1.RouteDomain{
							Name:     domain.Name,
							Internal: domain.Internal,
						},
						Destinations: []v1alpha1.RouteDestination{
							{
								Guid: "route-destination-guid-2",
								Weight:   intToPtr(20),
								Port:     intToPtr(9000),
								App:      v1alpha1.DestinationApp{
									Guid:    "app-guid-2",
									Process: v1alpha1.AppProcess{
										Type: "web",
									},
								},
								Selector: v1alpha1.DestinationSelector{
									MatchLabels: map[string]string{
										"cloudfoundry.org/app_guid": "app-guid-2",
										"cloudfoundry.org/process_type": "web",
									},
								},
							},
						},
					},
					Status: v1alpha1.RouteStatus{},
				}
			})
			It("returns false", func() {
				Expect(CompareRoutes(desiredRoute, actualRoute)).To(BeFalse())
			})
		})

		Context("when the actual Route has extra destinations", func() {
			BeforeEach(func() {
				actualRoute = v1alpha1.Route{
					TypeMeta: v1.TypeMeta{},
					ObjectMeta: v1.ObjectMeta{
						Name:      route.GUID,
						Namespace: namespace,
						Labels: map[string]string{
							KubeNameLabel:      route.GUID,
							KubeVersionLabel:   "0.0.0",
							KubeManagedByLabel: "cloudfoundry",
							KubeComponentLabel: "cf-networking",
							KubePartOfLabel:    "cloudfoundry",
							CFOrgGuidLabel:     space.Relationships["organization"].Data.GUID,
							CFSpaceGuidLabel:   space.GUID,
							CFDomainGuidLabel:  domain.GUID,
							CFRouteGuidLabel:   route.GUID,
						},
						CreationTimestamp: v1.Now(),
						ResourceVersion: "fake-resource-version",
					},
					Spec: v1alpha1.RouteSpec{
						Host: route.Host,
						Path: route.Path,
						Url:  route.URL,
						Domain: v1alpha1.RouteDomain{
							Name:     domain.Name,
							Internal: domain.Internal,
						},
						Destinations: []v1alpha1.RouteDestination{
							{
								Guid: "route-destination-guid-1",
								Weight:   intToPtr(80),
								Port:     intToPtr(9000),
								App:      v1alpha1.DestinationApp{
									Guid:    "app-guid-1",
									Process: v1alpha1.AppProcess{
										Type: "web",
									},
								},
								Selector: v1alpha1.DestinationSelector{
									MatchLabels: map[string]string{
										"cloudfoundry.org/app_guid": "app-guid-1",
										"cloudfoundry.org/process_type": "web",
									},
								},
							},
							{
								Guid: "route-destination-guid-2",
								Weight:   intToPtr(10),
								Port:     intToPtr(9000),
								App:      v1alpha1.DestinationApp{
									Guid:    "app-guid-2",
									Process: v1alpha1.AppProcess{
										Type: "web",
									},
								},
								Selector: v1alpha1.DestinationSelector{
									MatchLabels: map[string]string{
										"cloudfoundry.org/app_guid": "app-guid-2",
										"cloudfoundry.org/process_type": "web",
									},
								},
							},
							{
								Guid: "route-destination-guid-3",
								Weight:   intToPtr(10),
								Port:     intToPtr(9000),
								App:      v1alpha1.DestinationApp{
									Guid:    "app-guid-3",
									Process: v1alpha1.AppProcess{
										Type: "web",
									},
								},
								Selector: v1alpha1.DestinationSelector{
									MatchLabels: map[string]string{
										"cloudfoundry.org/app_guid": "app-guid-3",
										"cloudfoundry.org/process_type": "web",
									},
								},
							},
						},
					},
					Status: v1alpha1.RouteStatus{},
				}
			})

			It("returns false", func() {
				Expect(CompareRoutes(desiredRoute, actualRoute)).To(BeFalse())
			})
		})
	})
})

func intToPtr(i int) *int {
	return &i
}
