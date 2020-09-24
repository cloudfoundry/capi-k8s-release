package kubernetes

import (
	cfmodel "code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/model"
	"code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	KubeNameLabel      = "app.kubernetes.io/name"
	KubeVersionLabel   = "app.kubernetes.io/version"
	KubeManagedByLabel = "app.kubernetes.io/managed-by"
	KubeComponentLabel = "app.kubernetes.io/component"
	KubePartOfLabel    = "app.kubernetes.io/part-of"
	CFAppGuidLabel     = "cloudfoundry.org/app_guid"
	CFProcessTypeLabel = "cloudfoundry.org/process_type"
	CFOrgGuidLabel     = "cloudfoundry.org/org_guid"
	CFSpaceGuidLabel   = "cloudfoundry.org/space_guid"
	CFDomainGuidLabel  = "cloudfoundry.org/domain_guid"
	CFRouteGuidLabel   = "cloudfoundry.org/route_guid"
)

func TranslateRoute(route *cfmodel.Route, space *cfmodel.Space, domain *cfmodel.Domain, namespace string) v1alpha1.Route {
	destinations := make([]v1alpha1.RouteDestination, 0)

	for _, dest := range route.Destinations {
		destinations = append(destinations, v1alpha1.RouteDestination{
			Guid:   dest.GUID,
			Port:   &dest.Port,
			Weight: dest.Weight,
			App: v1alpha1.DestinationApp{
				Guid: dest.App.GUID,
				Process: v1alpha1.AppProcess{
					Type: dest.App.Process.Type,
				},
			},
			Selector: v1alpha1.DestinationSelector{
				MatchLabels: map[string]string{
					CFAppGuidLabel:     dest.App.GUID,
					CFProcessTypeLabel: dest.App.Process.Type,
				},
			},
		})
	}

	routeCR := v1alpha1.Route{
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
			Destinations: destinations,
		},
		Status: v1alpha1.RouteStatus{},
	}

	return routeCR
}
