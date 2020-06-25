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

	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-kpack-watcher/capi"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-kpack-watcher/capi_model"
	"github.com/go-logr/logr"
	kpackbuild "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const BuildGUIDLabel = "cloudfoundry.org/build_guid"

// BuildReconciler reconciles a Build object
type BuildReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	CFAPIClient *capi.Client
}

// +kubebuilder:rbac:groups=build.pivotal.io.build.pivotal.io,resources=buildren,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=build.pivotal.io.build.pivotal.io,resources=buildren/status,verbs=get;update;patch

func (r *BuildReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("request", req.NamespacedName)

	// fetch build object
	var build kpackbuild.Build
	err := r.Get(ctx, req.NamespacedName, &build)
	if err != nil {
		// TODO: should we requeue here?
		return ctrl.Result{}, err
	}
	// TODO: this amount of logging too much?
	logger = logger.WithValues("build", build)

	// handle deletion of kpack Build resource
	if isBeingDeleted(&build.ObjectMeta) {
		logger.V(1).Info("TODO: do we need to do anything if a kpack build is being deleted?")
		return ctrl.Result{}, nil
	}

	// handle create/update of a kpack Build resource
	// TODO: do we need to distinguish between create and update?
	if areAllContainersCompletedSuccessfully(build.Status.StepStates) {
		// mark build as staged successfully
		buildGUID := build.GetLabels()[BuildGUIDLabel]
		logger.WithValues("guid", buildGUID).V(1).Info("build completed successfully, marking as staged")

		err := r.CFAPIClient.UpdateBuild(buildGUID, capi_model.NewBuild(&build))
		if err != nil {
			logger.Error(err, "failed to send request to CF API")
			// TODO: is it safe to always requeue here?
			return ctrl.Result{Requeue: true}, nil
		}

		return ctrl.Result{}, nil
	} else {
		// if any steps have explicitly failed, then mark build as failed
		if hasAnyContainerFailed(build.Status.StepStates) {
			logger.V(1).Info("received a kpack build which failed; uh-oh")
			return ctrl.Result{}, nil
		} else {
			// else requeue
			return ctrl.Result{Requeue: true}, nil
		}
	}
}

func (r *BuildReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kpackbuild.Build{}).
		Complete(r)
}

func isBeingDeleted(objectMeta *metav1.ObjectMeta) bool {
	return !objectMeta.DeletionTimestamp.IsZero()
}

// returns true if all containers have terminated with exit code of zero
func areAllContainersCompletedSuccessfully(containerStates []corev1.ContainerState) bool {
	for _, container := range containerStates {
		if container.Terminated == nil || container.Terminated.ExitCode != 0 {
			return false
		}
	}
	return true
}

// returns true if any container has terminated with a non-zero exit code
func hasAnyContainerFailed(containerStates []corev1.ContainerState) bool {
	for _, container := range containerStates {
		if container.Terminated != nil && container.Terminated.ExitCode != 0 {
			return true
		}
	}
	return false
}
