package kubernetes

import (
	cfmodel "code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/model"
	"code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TranslateRoute(route cfmodel.Route, space cfmodel.Space, domain cfmodel.Domain, namespace string) v1alpha1.Route {
	var destinations []v1alpha1.RouteDestination

	for _, dest := range route.Destinations {
		// TODO: populate weight info once it is supported by the networking component(s)
		destinations = append(destinations, v1alpha1.RouteDestination{
			Guid: dest.GUID,
			Port: &dest.Port,
			App: v1alpha1.DestinationApp{
				Guid: dest.App.GUID,
				Process: v1alpha1.AppProcess{
					Type: dest.App.Process.Type,
				},
			},
			Selector: v1alpha1.DestinationSelector{
				MatchLabels: map[string]string{
					"cloudfoundry.org/app_guid":     dest.App.GUID,
					"cloudfoundry.org/process_type": dest.App.Process.Type,
				},
			},
		})
	}

	// TODO: create/re-use constants for the label keys
	routeCR := v1alpha1.Route{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			Name:      route.GUID,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       route.GUID,
				"cloudfoundry.org/org_guid":    space.Relationships["organization"].Data.GUID,
				"cloudfoundry.org/space_guid":  space.GUID,
				"cloudfoundry.org/domain_guid": domain.GUID,
				"cloudfoundry.org/route_guid":  route.GUID,
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
