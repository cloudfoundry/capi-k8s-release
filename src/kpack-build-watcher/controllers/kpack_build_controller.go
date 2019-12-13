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
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kpack "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
)

// KpackBuildReconciler reconciles a KpackBuild object
type KpackBuildReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=webapp.my.domain,resources=kpackbuilds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=webapp.my.domain,resources=kpackbuilds/status,verbs=get;update;patch

func (r *KpackBuildReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("kpackbuild", req.NamespacedName)

	// your logic here
	r.Log.Info("wat\n")
	var kpackImage kpack.Image
	if err := r.Get(ctx, req.NamespacedName, &kpackImage); err != nil {
		r.Log.Info("getting kpack image\n")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, err
	}

	fmt.Println()
	spew.Dump(kpackImage)
	fmt.Println()

	return ctrl.Result{}, nil
}

func (r *KpackBuildReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kpack.Image{}).
		Complete(r)
}
