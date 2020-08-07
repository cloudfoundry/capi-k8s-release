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
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/model"
	"context"
	"errors"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	buildv1alpha1 "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

const AppGUIDLabel = "cloudfoundry.org/app_guid"
const DropletGUIDLabel = "cloudfoundry.org/droplet_guid"

var ImageFilterError = errors.New("Received an image event with a non-image runtime.Object")

// ImageReconciler reconciles a Image object
type ImageReconciler struct {
	client.Client
	Log                logr.Logger
	Scheme             *runtime.Scheme
	CFClient *cf.Client
	AppsClientSet      *appsv1.AppsV1Client
	WorkloadsNamespace string
}

// +kubebuilder:rbac:groups=build.pivotal.io,resources=images,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=build.pivotal.io,resources=images/status,verbs=get;update;patch

func (r *ImageReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	var image buildv1alpha1.Image
	err := r.Get(ctx, req.NamespacedName, &image)
	if err != nil {
		r.Log.WithValues("request", req.NamespacedName).Error(err, "failed to fetch Image resource from cache")
		return ctrl.Result{}, err
	}
	logger := r.Log.WithValues("image", req.NamespacedName)

	if image.Status.GetCondition(corev1alpha1.ConditionReady).IsTrue() && image.Status.LatestBuildReason == "STACK" {
		return r.handleRebasedImage(image, logger)
	}

	logger.Info("Image status indicates either a failure or non-stack related updated, took no action")
	return ctrl.Result{}, nil
}

func (r *ImageReconciler) handleRebasedImage(image buildv1alpha1.Image, logger logr.Logger) (ctrl.Result, error) {
	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{AppGUIDLabel: image.ObjectMeta.Labels[AppGUIDLabel]}}
	statefulsets, err := r.AppsClientSet.StatefulSets(r.WorkloadsNamespace).
		List(v1.ListOptions{LabelSelector: labels.Set(labelSelector.MatchLabels).String()})
	if err != nil {
		logger.Error(err, "Could not find statefulsets for an app")
		return ctrl.Result{}, err
	}
	if len(statefulsets.Items) == 0 {
		logger.WithValues("appGUID", image.ObjectMeta.Labels[AppGUIDLabel]).
			Info("No statefulsets found for the app")
		return ctrl.Result{}, nil
	}
	if len(statefulsets.Items) > 1 {
		logger.WithValues("NumberOfStatefulSets", len(statefulsets.Items)).
			Info("Multiple statefulsets found for the app, requeueing image update event.")
		return ctrl.Result{Requeue: true}, nil
	}
	statefulset := statefulsets.Items[0]

	containers := statefulset.Spec.Template.Spec.Containers
	for i, container := range containers {
		if container.Name != "opi" {
			continue
		}
		containers[i].Image = image.Status.LatestImage

		_, err := r.AppsClientSet.StatefulSets(r.WorkloadsNamespace).Update(&statefulset)
		if err != nil {
			logger.Error(err, "Could not update statefulset")
			return ctrl.Result{}, err
		}
		logger.WithValues("newImage", image.Status.LatestImage).Info("Successfully updated app's StatefulSet with new image based off new stack")
		updateDropletRequest := model.Droplet{Image: image.Status.LatestImage}
		err = r.CFClient.UpdateDroplet(image.GetLabels()[DropletGUIDLabel], updateDropletRequest)
		if err != nil {
			logger.Error(err, "Failed to send request to CF API")
			// TODO: should we limit number of requeues? [story: #173573889]
			return ctrl.Result{Requeue: true}, err
		}
	}
	// Update associated droplet with new image reference

	return ctrl.Result{}, nil
}

func (r *ImageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(new(buildv1alpha1.Image)).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				r.Log.WithValues("requestLink", e.Meta.GetSelfLink()).
					V(1).Info("Image created, reconciling")
				return r.imageFilter(e.Object)
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				r.Log.WithValues("requestLink", e.MetaNew.GetSelfLink()).
					V(1).Info("Image updated, reconciling")
				return r.imageFilter(e.ObjectNew)
			},
			DeleteFunc:  func(_ event.DeleteEvent) bool { return false },
			GenericFunc: func(_ event.GenericEvent) bool { return false },
		}).
		Complete(r)
}

func (r *ImageReconciler) imageFilter(e runtime.Object) bool {
	image, ok := e.(*buildv1alpha1.Image)
	if !ok {
		r.Log.WithValues("event", e).Error(ImageFilterError, "ignoring event")
		return false
	}
	if _, present := image.ObjectMeta.Labels[AppGUIDLabel]; !present {
		r.Log.WithValues("image", image).Info("received update event for a non-CF Image resource, ignoring event")
		return false
	}
	return !image.Status.GetCondition(corev1alpha1.ConditionReady).IsUnknown()
}
