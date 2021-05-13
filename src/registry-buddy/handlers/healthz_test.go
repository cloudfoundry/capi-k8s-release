package handlers_test

import (
	"code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/handlers/fakes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"

	. "code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/handlers"
	"github.com/google/go-containerregistry/pkg/authn"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HealthzHandler", func() {
	var (
		healthyFunc  *fakes.HealthyFunc
		handler       http.HandlerFunc
		response      *httptest.ResponseRecorder
		authenticator authn.Authenticator
	)

	const (
		registryBasePath = "registry.example.com/example-registry"
	)

	BeforeEach(func() {
		healthyFunc = new(fakes.HealthyFunc)
		logger := log.New(GinkgoWriter, "", 0)
		authenticator = authn.FromConfig(authn.AuthConfig{
			Username: "some-user",
			Password: "some-password",
		})
		handler = HealthzHandler(registryBasePath, healthyFunc.Spy, logger, authenticator)
		response = httptest.NewRecorder()
	})

	When("happy path", func() {
		BeforeEach(func() {
			healthyFunc.Returns(nil)
		})

		It("is able to reach the registry", func() {
			req := httptest.NewRequest("GET", "/healthz", nil)

			handler.ServeHTTP(response, req)

			Expect(healthyFunc.CallCount()).To(Equal(1))

			actualRegistryPath, actualAuthenticator := healthyFunc.ArgsForCall(0)
			Expect(actualRegistryPath).To(Equal(registryBasePath))
			Expect(actualAuthenticator).To(Equal(authenticator))

			Expect(response.Code).To(Equal(200))
		})
	})

	When("registry is unreachable", func() {
		BeforeEach(func() {
			healthyFunc.Returns(errors.New("cannot reach registry"))
		})

		It("returns an error", func() {
			req := httptest.NewRequest("GET", "/healthz", nil)

			handler.ServeHTTP(response, req)

			Expect(healthyFunc.CallCount()).To(Equal(1))
			Expect(response.Code).To(Equal(401))
			body, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("unable to reach registry"))
		})
	})
})
