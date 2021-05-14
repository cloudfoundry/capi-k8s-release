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
	"errors"
	"fmt"
	"strings"

	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/model"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/image_registry"
	"github.com/buildpacks/lifecycle"
	"github.com/buildpacks/lifecycle/launch"
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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

const BuildGUIDLabel = "cloudfoundry.org/build_guid"
const BuildReasonAnnotation = "image.kpack.io/reason"
const StackUpdateBuildReason = "STACK"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fake/controller_runtime_client.go --fake-name ControllerRuntimeClient sigs.k8s.io/controller-runtime/pkg/client.Client

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fake/cf_build_updater.go --fake-name CFBuildUpdater . CfBuildUpdater
type CfBuildUpdater interface {
	UpdateBuild(buildGUID string, build model.Build) error
}

// BuildReconciler reconciles a Build object
type BuildReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	CFClient CfBuildUpdater
	image_registry.ImageConfigFetcher
}

// +kubebuilder:rbac:groups=kpack.io,resources=builds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kpack.io,resources=builds/status,verbs=get;update;patch

func (r *BuildReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	var build buildv1alpha1.Build
	err := r.Get(ctx, req.NamespacedName, &build)
	if err != nil {
		r.Log.WithValues("request", req.NamespacedName).Error(err, "failed initial build fetch")
		if apierrors.IsNotFound(err) {
			r.Log.WithValues("request", req.NamespacedName).Info("Attempted to reconcile build which no longer exists")
			return ctrl.Result{Requeue: false}, nil
		}
		return ctrl.Result{}, err
	}

	logger := r.Log.WithValues(
		"buildName", types.NamespacedName{Name: build.Name, Namespace: build.Namespace},
		BuildGUIDLabel, build.GetLabels()[BuildGUIDLabel],
		"status", build.Status,
	)

	condition := build.Status.GetCondition(corev1alpha1.ConditionSucceeded)
	if condition.IsTrue() {
		return r.reconcileSuccessfulBuild(&build, logger)
	}

	failureMessage := fmt.Sprintf(
		"Kpack build unsuccessful: Build failure reason: '%s', message: '%s'.",
		condition.Reason,
		condition.Message,
	)

	failedContainerState := findAnyFailedContainerState(build.Status.StepStates)
	if failedContainerState != nil {
		failureMessage = fmt.Sprintf(
			"Kpack build failed during container execution: Step failure reason: '%s', message: '%s'.",
			failedContainerState.Terminated.Reason,
			failedContainerState.Terminated.Message,
		)
	}

	return r.reconcileFailedBuild(&build, failureMessage, logger)
}

func (r *BuildReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(new(buildv1alpha1.Build)).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				r.Log.WithValues("requestLink", e.Meta.GetSelfLink()).
					V(1).Info("Build create event received")
				return r.buildFilter(e.Object)
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				r.Log.WithValues("requestLink", e.MetaNew.GetSelfLink()).
					V(1).Info("Build update event received")
				return r.buildFilter(e.ObjectNew)
			},
			DeleteFunc:  func(_ event.DeleteEvent) bool { return false },
			GenericFunc: func(_ event.GenericEvent) bool { return false },
		}).
		Complete(r)
}

var BuildFilterError = errors.New("Received a build event with a non-build runtime.Object")

func (r *BuildReconciler) buildFilter(e runtime.Object) bool {
	newBuild, ok := e.(*buildv1alpha1.Build)
	if !ok {
		r.Log.WithValues("event", e).Error(BuildFilterError, "ignoring event")
		return false
	}

	if _, isGuidPresent := newBuild.ObjectMeta.Labels[BuildGUIDLabel]; !isGuidPresent {
		r.Log.WithValues("build", newBuild).V(1).Info("ignoring event: received update event for a non-CF Build resource")
		return false
	}
	buildReason, ok := newBuild.ObjectMeta.Annotations[BuildReasonAnnotation]
	if !ok {
		r.Log.WithValues("build", newBuild).V(1).Info("ignoring event: received update event that was missing the build reason")
		return false
	}

	// Stack updates are handled by the Image controller
	if buildReason == StackUpdateBuildReason {
		r.Log.WithValues("build", newBuild).V(1).Info("ignoring event: build triggered due to an automatic stack update")
		return false
	}

	// Wait until the 'Succeeded' condition is in a terminal 'False' or 'True' state
	if newBuild.Status.GetCondition(corev1alpha1.ConditionSucceeded).IsUnknown() {
		r.Log.WithValues("build", newBuild).V(1).Info("ignoring event: build 'Succeeded' condition status is Unknown")
		return false
	}

	r.Log.WithValues("build", newBuild).V(1).Info("event passed ignore filters, continuing with reconciliation")
	return true
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
		ret[process.Type] = extractFullCommand(process)
	}
	return ret, nil
}

func extractFullCommand(process launch.Process) string {
	commandWithArgs := append([]string{process.Command}, process.Args...)
	return strings.Join(commandWithArgs, " ")
}

func (r *BuildReconciler) reconcileSuccessfulBuild(build *buildv1alpha1.Build, logger logr.Logger) (ctrl.Result, error) {
	logger.V(1).Info("Build completed successfully, marking as staged")

	processTypes, err := r.extractProcessTypes(build)
	if err != nil {
		logger.Error(err, "Failed to fetch image config")
		return r.reconcileFailedBuild(
			build,
			fmt.Sprintf(
				"Failed to handle successful kpack build: %s",
				err,
			),
			logger,
		)
	}

	updateBuildRequest := model.NewBuildFromKpackBuild(build)
	updateBuildRequest.Lifecycle.Data.ProcessTypes = processTypes
	err = r.CFClient.UpdateBuild(build.GetLabels()[BuildGUIDLabel], updateBuildRequest)
	if err != nil {
		logger.Error(err, "Failed to send request to CF API")
		// TODO: should we limit number of requeues? [story: #173573889]
		return ctrl.Result{Requeue: true}, err
	}

	return ctrl.Result{}, nil
}

func (r *BuildReconciler) reconcileFailedBuild(build *buildv1alpha1.Build, errorMessage string, logger logr.Logger) (ctrl.Result, error) {
	logger.V(1).Info("Build failed, marking as failed staging")

	err := r.CFClient.UpdateBuild(build.GetLabels()[BuildGUIDLabel], model.Build{
		State: model.BuildFailedState,
		Error: errorMessage,
	})
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
