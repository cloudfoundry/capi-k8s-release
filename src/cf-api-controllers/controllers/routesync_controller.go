/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"errors"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/kubernetes"

	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/model"

	networkingv1alpha1 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/apis/apps.cloudfoundry.org/v1alpha1"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf"
)

// RouteSyncReconciler reconciles a RouteSync object
type RouteSyncReconciler struct {
	CtrClient          client.Client
	Log                logr.Logger
	Scheme             *runtime.Scheme
	CFClient           cf.ClientInterface
	WorkloadsNamespace string
}

// +kubebuilder:rbac:groups=apps.cloudfoundry.org,resources=routesyncs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.cloudfoundry.org,resources=routesyncs/status,verbs=get;update;patch

func (r *RouteSyncReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("routesync", req.NamespacedName)

	var routeSync appsv1alpha1.RouteSync
	err := r.CtrClient.Get(ctx, req.NamespacedName, &routeSync)
	if err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.WithValues("request", req.NamespacedName).Error(err, "RouteSync resource not found")
		} else {
			r.updateSyncStatusFailure(ctx, &routeSync, err.Error())
		}
		return ctrl.Result{}, client.IgnoreNotFound(err) // untested
	}

	// TODO: confirm that kubebuilder replays all events on startup (i.e. its bookmark of last event doesn't persist somewhere)

	// 1. fetch all routes from CC (done)
	routesInCC, err := r.CFClient.ListRoutes()
	if err != nil {
		r.updateSyncStatusFailure(ctx, &routeSync, err.Error())
		return ctrl.Result{}, fmt.Errorf("error listing routes from CF API: %w", err) // untested
	}

	// 2. fetch all route CRs from kubernetes
	var routesInK8s networkingv1alpha1.RouteList
	err = r.CtrClient.List(ctx, &routesInK8s, &client.ListOptions{Namespace: r.WorkloadsNamespace})
	if err != nil {
		r.updateSyncStatusFailure(ctx, &routeSync, err.Error())
		return ctrl.Result{}, err // untested
	}

	// 3. map each slice of routes into a slice of strings
	ccRouteMap := make(map[string]*model.Route)
	for _, ccRoute := range routesInCC {
		ccRouteMap[ccRoute.GUID] = &ccRoute
	}

	k8sRouteGUIDs := make(map[string]struct{})
	for _, k8sRoute := range routesInK8s.Items {
		k8sRouteGUIDs[k8sRoute.Name] = struct{}{}
	}

	// 4. calculate the set of route GUIDs which need to be created in k8s (mostly done)
	var missingInK8s []*model.Route
	for ccRouteGuid, ccRoute := range ccRouteMap {
		if _, ok := k8sRouteGUIDs[ccRouteGuid]; !ok {
			missingInK8s = append(missingInK8s, ccRoute)
		}
	}

	reconciledSuccessfully := true

	// 5. iterate over all routes to be created and create them (started)
	for _, ccRoute := range missingInK8s {
		// 5a. fetch additional information from CC (domain, space)
		spaceGUID := ccRoute.Relationships["space"].Data.GUID
		space, err := r.CFClient.GetSpace(spaceGUID)
		if err != nil {
			reconciledSuccessfully = false
			r.Log.WithValues("request", req.NamespacedName, "space_guid", spaceGUID).Error(err, "errored fetching CF API space")
		}

		domainGUID := ccRoute.Relationships["domain"].Data.GUID
		domain, err := r.CFClient.GetDomain(domainGUID)
		if err != nil {
			reconciledSuccessfully = false
			r.Log.WithValues("request", req.NamespacedName, "domain_guid", domainGUID).Error(err, "errored fetching CF API domain")
		}

		// 5b. translate CC objects into k8s CR
		newRouteCR := kubernetes.TranslateRoute(*ccRoute, space, domain, r.WorkloadsNamespace)

		// 5c. send CR to k8s for creation
		err = r.CtrClient.Create(ctx, &newRouteCR)
		if err != nil {
			reconciledSuccessfully = false
			r.Log.WithValues("request", req.NamespacedName, "route_guid", ccRoute.GUID).Error(err, "errored creating Route resource in k8s")
		}
	}

	// 6. calculate the set of route GUIDs which need to be deleted in k8s (mostly done)
	var extraInK8s []string
	for k8sRouteGuid, _ := range k8sRouteGUIDs {
		if _, ok := ccRouteMap[k8sRouteGuid]; !ok {
			extraInK8s = append(extraInK8s, k8sRouteGuid)
		}
	}

	// 7. iterate over all routes to be deleted and delete them
	for _, extraRouteGUID := range extraInK8s {
		err = r.CtrClient.Delete(ctx, &networkingv1alpha1.Route{
			ObjectMeta: metav1.ObjectMeta{
				Name:      extraRouteGUID,
				Namespace: r.WorkloadsNamespace,
			},
		})
		if err != nil {
			reconciledSuccessfully = false
			r.Log.WithValues("request", req.NamespacedName, "route_guid", extraRouteGUID).Error(err, "errored deleting Route resource in k8s")
		}
	}

	if !reconciledSuccessfully {
		err := errors.New("failed to reconcile at least one route")
		r.updateSyncStatusFailure(ctx, &routeSync, err.Error())
		return ctrl.Result{}, err // untested
	}

	if err := r.updateSyncStatusSuccess(ctx, &routeSync); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: time.Duration(routeSync.Spec.PeriodSeconds) * time.Second}, nil
}

func (r *RouteSyncReconciler) updateSyncStatusSuccess(ctx context.Context, routeSync *appsv1alpha1.RouteSync) error {
	setRouteSyncStatus(routeSync, appsv1alpha1.TrueConditionStatus, appsv1alpha1.CompletedConditionReason, "")

	if err := r.CtrClient.Status().Update(ctx, routeSync); err != nil {
		return err
	}

	return nil
}

func (r *RouteSyncReconciler) updateSyncStatusFailure(ctx context.Context, routeSync *appsv1alpha1.RouteSync, failureMessage string) error {
	setRouteSyncStatus(routeSync, appsv1alpha1.FalseConditionStatus, appsv1alpha1.FailedConditionReason, failureMessage)

	if err := r.CtrClient.Status().Update(ctx, routeSync); err != nil {
		return err
	}

	return nil
}

func setRouteSyncStatus(routeSync *appsv1alpha1.RouteSync, status appsv1alpha1.ConditionStatus, reason, message string) {
	routeSync.Status.Conditions = []appsv1alpha1.Condition{
		{
			Type:               appsv1alpha1.SyncedConditionType,
			Status:             status,
			LastTransitionTime: metav1.Now(),
			Reason:             reason,
			Message:            message,
		},
	}
}

func (r *RouteSyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.RouteSync{}).
		Complete(r)
}
