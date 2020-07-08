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
	"crypto/tls"
	"net/http"
	"path/filepath"
	"testing"

	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/cffakes"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/image_registry/image_registryfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	buildpivotaliov1alpha1 "github.com/pivotal/kpack/pkg/client/clientset/versioned/scheme"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.
var (
	config          *rest.Config
	k8sManager      ctrl.Manager
	managerStopChan chan struct{}
	k8sClient       client.Client
	testEnv         *envtest.Environment

	fakeCFAPIServer        *ghttp.Server
	cfClient               cf.Client
	mockUAAClient          cffakes.FakeTokenFetcher
	mockImageConfigFetcher image_registryfakes.FakeImageConfigFetcher
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	By("bootstrapping test environment")
	// TODO: we need to correctly version the kpack go module and the kpack CRD file we bring in
	testEnv = &envtest.Environment{
		ErrorIfCRDPathMissing: true,
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd")},
	}

	var err error
	config, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(config).ToNot(BeNil())

	// TODO: is this generic scheme needed?
	err = scheme.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = buildpivotaliov1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sManager, err = ctrl.NewManager(config, ctrl.Options{
		Scheme:                 scheme.Scheme,
		HealthProbeBindAddress: "0",
		MetricsBindAddress:     "0",
	})
	Expect(err).NotTo(HaveOccurred())

	// initialize mocks defensively in case cache sync touches them
	mockUAAClient = cffakes.FakeTokenFetcher{}
	mockImageConfigFetcher = image_registryfakes.FakeImageConfigFetcher{}

	fakeCFAPIServer = ghttp.NewServer()
	cfClient = *cf.NewClient(fakeCFAPIServer.URL(), &cf.RestClient{
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}, &mockUAAClient)

	// start controller with manager
	// TODO: refactor to remove mocks since this is an integration test
	err = (&BuildReconciler{
		Client:             k8sManager.GetClient(),
		Log:                ctrl.Log.WithName("controllers").WithName("Build"),
		Scheme:             k8sManager.GetScheme(),
		CFClient:           &cfClient,
		ImageConfigFetcher: &mockImageConfigFetcher,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	managerStopChan = make(chan struct{})
	go func() {
		err = k8sManager.Start(managerStopChan)
		Expect(err).NotTo(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

var _ = BeforeEach(func() {
	mockUAAClient = cffakes.FakeTokenFetcher{}
	mockImageConfigFetcher = image_registryfakes.FakeImageConfigFetcher{}
})

var _ = AfterSuite(func() {
	By("tearing down the fake CF API Server")
	fakeCFAPIServer.Close()

	By("tearing down the controller")
	if managerStopChan != nil {
		close(managerStopChan)
	}

	By("tearing down the test environment")
	// gexec.KillAndWait(5 * time.Second)
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
