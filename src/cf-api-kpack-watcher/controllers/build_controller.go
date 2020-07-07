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
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-kpack-watcher/cf"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-kpack-watcher/cf/api_model"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-kpack-watcher/image_registry"
	"github.com/buildpacks/lifecycle"
	"github.com/go-logr/logr"
	buildv1alpha1 "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const BuildGUIDLabel = "cloudfoundry.org/build_guid"

// BuildReconciler reconciles a Build object
type BuildReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	CFClient *cf.Client
	image_registry.ImageConfigFetcher
}

// +kubebuilder:rbac:groups=build.pivotal.io,resources=builds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=build.pivotal.io,resources=builds/status,verbs=get;update;patch

func (r *BuildReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("request", req.NamespacedName)

	// TODO: only process builds for a particular namespace

	// fetch build object
	var build buildv1alpha1.Build
	err := r.Get(ctx, req.NamespacedName, &build)
	if err != nil {
		// TODO: should we requeue here?
		// TODO: might need to do `client.IgnoreNotFound(err)` to deal with delete events
		return ctrl.Result{}, err
	}

	// handle update of a kpack build resource
	if build.Status.GetCondition(corev1alpha1.ConditionSucceeded).IsTrue() {
		// mark build as staged successfully
		return r.reconcileSuccessfulBuild(&build)
	}

	// if any steps have explicitly failed, then mark CCNG build as failed
	failedContainerState := findAnyFailedContainerState(build.Status.StepStates)
	if failedContainerState != nil {
		return r.reconcileFailedBuild(
			&build,
			fmt.Sprintf(
				"Kpack build failed. Build failure message: '%s'.",
				failedContainerState.Terminated.Message,
			),
		)
	}

	// the update event filter in `SetupWithManager` should prevent this from being reached
	logger.V(1).Info("[Should not have gotten here] Build is still progressing, requeueing")
	return ctrl.Result{Requeue: true}, nil
}

func (r *BuildReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&buildv1alpha1.Build{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				// TODO: log self-link or UID for debugging?
				r.Log.WithValues("requestLink", e.Meta.GetSelfLink()).V(1).Info("Build created, watching for updates...")
				return false
			},
			UpdateFunc:  r.updateEventFilter,
			DeleteFunc:  func(_ event.DeleteEvent) bool { return false },
			GenericFunc: func(_ event.GenericEvent) bool { return false },
		}).
		Complete(r)
}

func (r *BuildReconciler) updateEventFilter(e event.UpdateEvent) bool {
	// TODO: should we filter out builds that are not managed by CF? (i.e. don't have the
	// `cloudfoundry.org/*` labels on the objects)
	newBuild, ok := e.ObjectNew.(*buildv1alpha1.Build)
	if !ok {
		// TODO: log something? what log level?
		r.Log.WithValues("event", e).V(100).Info("Received a build update event that couldn't be deserialized")
		return false
	}
	return !newBuild.Status.GetCondition(corev1alpha1.ConditionSucceeded).IsUnknown()
}

func (r *BuildReconciler) extractProcessTypes(build *buildv1alpha1.Build) (map[string]string, error) {
	imageConfig, err := r.FetchImageConfig(build.Status.LatestImage, build.Spec.ServiceAccount, build.Namespace)
	if err != nil {
		return nil, err
	}

	var buildMetadata lifecycle.BuildMetadata
	if err = json.Unmarshal([]byte(imageConfig.Labels[lifecycle.BuildMetadataLabel]), &buildMetadata); err != nil {
		return nil, err
	}

	ret := make(map[string]string)
	for _, process := range buildMetadata.Processes {
		ret[process.Type] = process.Command
	}
	return ret, nil
}

func (r *BuildReconciler) reconcileSuccessfulBuild(build *buildv1alpha1.Build) (ctrl.Result, error) {
	logger := r.Log.WithValues("request", types.NamespacedName{Name: build.Name, Namespace: build.Namespace})

	buildGUID := build.GetLabels()[BuildGUIDLabel]
	logger.WithValues("guid", buildGUID).V(1).Info("Build completed successfully, marking as staged")

	processTypes, err := r.extractProcessTypes(build)
	if err != nil {
		logger.WithValues("error", err).V(1).Info("Failed to fetch image config")
		return r.reconcileFailedBuild(
			build,
			fmt.Sprintf(
				"Failed to handle successful kpack build: %s",
				err,
			),
		)
	}
	updateBuildRequest := api_model.NewBuild(build)
	updateBuildRequest.Lifecycle.Data.ProcessTypes = processTypes

	// TODO: do the things to determine `processTypes` stuff
	err = r.CFClient.UpdateBuild(buildGUID, updateBuildRequest)
	if err != nil {
		logger.Error(err, "Failed to send request to CF API")
		// TODO: should we limit number of requeues? [story: #173573889]
		return ctrl.Result{Requeue: true}, err
	}

	return ctrl.Result{}, nil
}

func (r *BuildReconciler) reconcileFailedBuild(build *buildv1alpha1.Build, errorMessage string) (ctrl.Result, error) {
	logger := r.Log.WithValues("request", types.NamespacedName{Name: build.Name, Namespace: build.Namespace})

	logger.V(1).Info("Build failed, marking as failed staging")

	buildGUID := build.GetLabels()[BuildGUIDLabel]
	cfAPIBuild := api_model.Build{
		State: api_model.BuildFailedState,
		Error: errorMessage,
	}
	err := r.CFClient.UpdateBuild(buildGUID, cfAPIBuild)
	if err != nil {
		logger.Error(err, "Failed to send request to CF API")
		return ctrl.Result{Requeue: true}, err
	}

	return ctrl.Result{}, nil
}

// returns true if any container has terminated with a non-zero exit code
func findAnyFailedContainerState(containerStates []corev1.ContainerState) *corev1.ContainerState {
	for _, container := range containerStates {
		if container.Terminated != nil && container.Terminated.ExitCode != 0 {
			return &container
		}
	}
	return nil
}
