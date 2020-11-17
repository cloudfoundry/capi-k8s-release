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

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fake/controller_runtime_client.go --fake-name ControllerRuntimeClient sigs.k8s.io/controller-runtime/pkg/client.Client

// PeriodicSyncReconciler reconciles a PeriodicSync object
type PeriodicSyncReconciler struct {
	client.Client
	Log                logr.Logger
	Scheme             *runtime.Scheme
	CFClient           cf.ClientInterface
	WorkloadsNamespace string
}

// +kubebuilder:rbac:groups=apps.cloudfoundry.org,resources=periodicsyncs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.cloudfoundry.org,resources=periodicsyncs/status,verbs=get;update;patch

func (r *PeriodicSyncReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("periodicsync", req.NamespacedName)

	var periodicSync appsv1alpha1.PeriodicSync
	err := r.Get(ctx, req.NamespacedName, &periodicSync)
	if err != nil {
		if apierrors.IsNotFound(err) { // untested
			r.Log.WithValues("request", req.NamespacedName).Error(err, "PeriodicSync resource not found")
		} else {
			r.updateSyncStatusFailure(ctx, &periodicSync, err.Error())
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	ccRouteList, err := r.CFClient.ListRoutes()
	if err != nil {
		r.updateSyncStatusFailure(ctx, &periodicSync, err.Error())
		return ctrl.Result{}, fmt.Errorf("error listing routes from CF API: %w", err)
	}

	var routesInK8s networkingv1alpha1.RouteList
	err = r.List(ctx, &routesInK8s, &client.ListOptions{Namespace: r.WorkloadsNamespace})
	if err != nil {
		r.updateSyncStatusFailure(ctx, &periodicSync, err.Error())
		return ctrl.Result{}, fmt.Errorf("error listing routes from kubernetes API: %w", err)
	}

	ccRouteMap := make(map[string]*model.Route)
	for i, ccRoute := range ccRouteList.Resources {
		ccRouteMap[ccRoute.GUID] = &ccRouteList.Resources[i]
	}

	ccSpaceMap := make(map[string]*model.Space)
	for i, ccSpace := range ccRouteList.Included.Spaces {
		ccSpaceMap[ccSpace.GUID] = &ccRouteList.Included.Spaces[i]
	}

	ccDomainMap := make(map[string]*model.Domain)
	for i, ccDomain := range ccRouteList.Included.Domains {
		ccDomainMap[ccDomain.GUID] = &ccRouteList.Included.Domains[i]
	}

	k8sRouteMap := make(map[string]networkingv1alpha1.Route)
	for _, k8sRoute := range routesInK8s.Items {
		k8sRouteMap[k8sRoute.Name] = k8sRoute
	}

	reconciledSuccessfully := true

	var missingInK8s []*model.Route
	for ccRouteGuid, ccRoute := range ccRouteMap {
		if k8sRoute, ok := k8sRouteMap[ccRouteGuid]; ok {
			spaceGUID := ccRoute.Relationships["space"].Data.GUID
			domainGUID := ccRoute.Relationships["domain"].Data.GUID

			desiredRoute := kubernetes.TranslateRoute(ccRoute, ccSpaceMap[spaceGUID], ccDomainMap[domainGUID], r.WorkloadsNamespace)

			if kubernetes.CompareRoutes(desiredRoute, k8sRoute) {
				continue
			}

			k8sRoute.Spec = desiredRoute.Spec
			err = r.Update(ctx, &k8sRoute)

			if err != nil {
				reconciledSuccessfully = false
				r.Log.WithValues("request", req.NamespacedName, "route_guid", k8sRoute.Name).Error(err, "errored updating Route resource in k8s")
				continue
			}

			r.Log.WithValues("request", req.NamespacedName, "route_guid", k8sRoute.Name).Info("successfully updated Route resource")
		} else {
			// track the set of route GUIDs which need to be created in k8s
			missingInK8s = append(missingInK8s, ccRoute)
		}
	}

	// iterate over all routes to be created and create them
	for _, ccRoute := range missingInK8s {
		// fetch additional required information from CC (domain, space)
		spaceGUID := ccRoute.Relationships["space"].Data.GUID
		domainGUID := ccRoute.Relationships["domain"].Data.GUID

		newRouteCR := kubernetes.TranslateRoute(ccRoute, ccSpaceMap[spaceGUID], ccDomainMap[domainGUID], r.WorkloadsNamespace)

		err = r.Create(ctx, &newRouteCR)
		if err != nil {
			reconciledSuccessfully = false
			r.Log.WithValues("request", req.NamespacedName, "route_guid", ccRoute.GUID).Error(err, "errored creating Route resource in k8s")
			continue
		}

		r.Log.WithValues("request", req.NamespacedName, "route_guid", ccRoute.GUID).Info("successfully created Route resource")
	}

	// calculate the set of route GUIDs which need to be deleted in k8s
	var extraInK8s []string
	for k8sRouteGuid, _ := range k8sRouteMap {
		if _, ok := ccRouteMap[k8sRouteGuid]; !ok {
			extraInK8s = append(extraInK8s, k8sRouteGuid)
		}
	}

	// iterate over all routes to be deleted and delete them
	for _, extraRouteGUID := range extraInK8s {
		err = r.Delete(ctx, &networkingv1alpha1.Route{
			ObjectMeta: metav1.ObjectMeta{
				Name:      extraRouteGUID,
				Namespace: r.WorkloadsNamespace,
			},
		})

		// ignoring "not found" errors because the Route is already gone from k8s
		if err != nil && !apierrors.IsNotFound(err) {
			reconciledSuccessfully = false
			r.Log.WithValues("request", req.NamespacedName, "route_guid", extraRouteGUID).Error(err, "errored deleting Route resource in k8s")
			continue
		}

		r.Log.WithValues("request", req.NamespacedName, "route_guid", extraRouteGUID).Info("successfully deleted Route resource")
	}

	if !reconciledSuccessfully {
		err := errors.New("failed to reconcile at least one route")
		r.updateSyncStatusFailure(ctx, &periodicSync, err.Error())
		return ctrl.Result{}, err
	}

	if err := r.updateSyncStatusSuccess(ctx, &periodicSync); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: time.Duration(periodicSync.Spec.PeriodSeconds) * time.Second}, nil
}

func (r *PeriodicSyncReconciler) updateSyncStatusSuccess(ctx context.Context, periodicSync *appsv1alpha1.PeriodicSync) error {
	setPeriodicSyncStatus(periodicSync, appsv1alpha1.TrueConditionStatus, appsv1alpha1.CompletedConditionReason, "")

	if err := r.Status().Update(ctx, periodicSync); err != nil {
		return err
	}

	return nil
}

func (r *PeriodicSyncReconciler) updateSyncStatusFailure(ctx context.Context, periodicSync *appsv1alpha1.PeriodicSync, failureMessage string) error {
	setPeriodicSyncStatus(periodicSync, appsv1alpha1.FalseConditionStatus, appsv1alpha1.FailedConditionReason, failureMessage)

	if err := r.Status().Update(ctx, periodicSync); err != nil {
		return err
	}

	return nil
}

func setPeriodicSyncStatus(periodicSync *appsv1alpha1.PeriodicSync, status appsv1alpha1.ConditionStatus, reason, message string) {
	periodicSync.Status.Conditions = []appsv1alpha1.Condition{
		{
			Type:               appsv1alpha1.SyncedConditionType,
			Status:             status,
			LastTransitionTime: metav1.Now(),
			Reason:             reason,
			Message:            message,
		},
	}
}

func (r *PeriodicSyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.PeriodicSync{}).
		Complete(r)
}
