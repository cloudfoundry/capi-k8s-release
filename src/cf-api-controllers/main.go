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

package main

import (
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/auth"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/image_registry"
	"crypto/tls"
	"flag"
	"github.com/pivotal/kpack/pkg/dockercreds/k8sdockercreds"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"net/http"
	"os"

	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	appsv1alpha1 "code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/apis/apps.cloudfoundry.org/v1alpha1"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/controllers"
	networkingv1alpha1 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"
	buildpivotaliov1alpha1 "github.com/pivotal/kpack/pkg/client/clientset/versioned/scheme"
	// +kubebuilder:scaffold:imports
)

var (
	// TODO: revisit whether we should assemble our own Scheme
	scheme   = buildpivotaliov1alpha1.Scheme
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = buildpivotaliov1alpha1.AddToScheme(scheme)
	_ = networkingv1alpha1.AddToScheme(scheme)
	_ = appsv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	if os.Getenv("CF_API_HOST") == "" {
		panic("`CF_API_HOST` environment variable must be set")
	}
	if os.Getenv("STAGING_NAMESPACE") == "" {
		panic("`STAGING_NAMESPACE` environment variable must be set")
	}
	if os.Getenv("WORKLOADS_NAMESPACE") == "" {
		panic("`WORKLOADS_NAMESPACE` environment variable must be set")
	}

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "7cba68d7.build.pivotal.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	client := kubernetes.NewForConfigOrDie(mgr.GetConfig())
	keychainFactory, err := k8sdockercreds.NewSecretKeychainFactory(client)
	if err != nil {
		panic(err.Error())
	}

	uaaClient := auth.NewUAAClient()
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	if err = (&controllers.BuildReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Build"),
		Scheme: mgr.GetScheme(),
		CFClient: cf.NewClient(os.Getenv("CF_API_HOST"), &cf.RestClient{
			Client: httpClient,
		}, uaaClient),
		ImageConfigFetcher: image_registry.NewImageConfigFetcher(keychainFactory),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Build")
		os.Exit(1)
	}

	clientset, err := appsv1.NewForConfig(mgr.GetConfig())
	if err != nil {
		// TODO: something better?
		panic(err.Error())
	}
	if err = (&controllers.ImageReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Image"),
		Scheme: mgr.GetScheme(),
		CFClient: cf.NewClient(os.Getenv("CF_API_HOST"), &cf.RestClient{
			Client: httpClient,
		}, uaaClient),
		AppsClientSet:      clientset,
		WorkloadsNamespace: os.Getenv("WORKLOADS_NAMESPACE"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Image")
		os.Exit(1)
	}
	if err = (&controllers.RouteSyncReconciler{
		CtrClient:          mgr.GetClient(),
		Log:                ctrl.Log.WithName("controllers").WithName("RouteSync"),
		Scheme:             mgr.GetScheme(),
		WorkloadsNamespace: os.Getenv("WORKLOADS_NAMESPACE"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RouteSync")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
