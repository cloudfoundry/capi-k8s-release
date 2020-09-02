package kubernetes_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"

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

	Context("given a CF API route object with all required fields", func() {
		const (
			namespace = "some-dope-space"
		)

		BeforeEach(func() {
			route = model.Route{
				GUID: "route-guid",
				Port: 1234,
				Host: "host",
				Path: "/path",
				URL:  "host.domain.com/path",
				Destinations: []model.Destination{
					{
						GUID: "destination-guid",
						Port: 8080,
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
			routeCR = TranslateRoute(route, space, domain, namespace)

			Expect(routeCR).NotTo(BeNil())

			Expect(routeCR.ObjectMeta.Name).To(Equal(route.GUID))
			Expect(routeCR.ObjectMeta.Namespace).To(Equal(namespace))
			// TODO: assertions for all the metadata (e.g. labels and annotations)
			Expect(routeCR.ObjectMeta.Labels).To(HaveKeyWithValue("app.kubernetes.io/name", route.GUID))
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
			// TODO: assert against weight info once it is supported by the networking component(s)
			Expect(routeCR.Spec.Destinations[0].App.Guid).To(Equal(route.Destinations[0].App.GUID))
			Expect(routeCR.Spec.Destinations[0].App.Process.Type).To(Equal(route.Destinations[0].App.Process.Type))

			Expect(routeCR.Spec.Destinations[0].Selector.MatchLabels).To(HaveLen(2))
			Expect(routeCR.Spec.Destinations[0].Selector.MatchLabels).To(HaveKeyWithValue("cloudfoundry.org/app_guid", route.Destinations[0].App.GUID))
			Expect(routeCR.Spec.Destinations[0].Selector.MatchLabels).To(HaveKeyWithValue("cloudfoundry.org/process_type", route.Destinations[0].App.Process.Type))
		})
	})
})
