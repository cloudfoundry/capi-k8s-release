package integration_test

import (
	"log"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/handlers"
	"code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/package_upload"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/gorilla/mux"

	"github.com/google/go-containerregistry/pkg/authn"

	"github.com/matt-royal/biloba"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Integration Suite", biloba.GoLandReporter())
}

var (
	testServer       *httptest.Server
	registryBasePath string
	authenticator    authn.Authenticator
	notConfigured    bool
)

const (
	defaultTimeout         = 30 * time.Second
	defaultPollingInterval = 1 * time.Second
)

var _ = BeforeSuite(func() {
	if os.Getenv("REGISTRY_USERNAME") == "" || os.Getenv("REGISTRY_PASSWORD") == "" || os.Getenv("REGISTRY_BASE_PATH") == "" {
		notConfigured = true
		return
	}

	registryBasePath = os.Getenv("REGISTRY_BASE_PATH")
	authenticator = authn.FromConfig(authn.AuthConfig{
		Username: os.Getenv("REGISTRY_USERNAME"),
		Password: os.Getenv("REGISTRY_PASSWORD"),
	})

	testServer = startServer(authenticator)

	SetDefaultEventuallyTimeout(defaultTimeout)
	SetDefaultEventuallyPollingInterval(defaultPollingInterval)
	SetDefaultConsistentlyDuration(defaultTimeout)
	SetDefaultConsistentlyPollingInterval(defaultPollingInterval)
})

var _ = AfterSuite(func() {
	if testServer != nil {
		testServer.Close()
	}
})

var _ = BeforeEach(func() {
	if notConfigured {
		Skip("Missing required config, skipping registry integration tests")
	}
})

func startServer(authenticator authn.Authenticator) *httptest.Server {
	logger := log.New(GinkgoWriter, "", 0)
	r := mux.NewRouter()

	r.HandleFunc("/packages", handlers.PostPackageHandler(package_upload.Upload, logger, authenticator)).Methods("POST")
	r.HandleFunc("/images", handlers.DeleteImageHandler(remote.Delete, remote.Get, logger, authenticator)).Methods("DELETE")

	return httptest.NewServer(r)
}
