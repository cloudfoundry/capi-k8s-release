package handlers_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	"code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/handlers"

	"github.com/google/go-containerregistry/pkg/authn"

	"code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/handlers/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeleteImageHandler", func() {
	var (
		handler         http.HandlerFunc
		imageDeleteFunc *fakes.ImageDeleteFunc
		response        *httptest.ResponseRecorder
		authenticator   authn.Authenticator
		jsonBody        string
	)

	const (
		registryBasePath = "registry.example.com/cf-workloads"
		imageReference   = "some-package@sha256:15e8a86a2cfce269dbc4321e741f729d60f41cafc7a8f7e11cd2892a4de3bf39"
	)

	BeforeEach(func() {
		imageDeleteFunc = new(fakes.ImageDeleteFunc)
		logger := log.New(GinkgoWriter, "", 0)
		authenticator = authn.FromConfig(authn.AuthConfig{
			Username: "some-user",
			Password: "some-password",
		})
		handler = handlers.DeleteImageHandler(imageDeleteFunc.Spy, logger, authenticator)
		response = httptest.NewRecorder()
	})

	Context("successfully deleting an image", func() {
		BeforeEach(func() {
			jsonBody = `{
              "image_reference": "` + imageReference + `",
              "registry_base_path": "` + registryBasePath + `"
            }`

			response = httptest.NewRecorder()
		})

		It("deletes the image from the registry", func() {
			req := httptest.NewRequest("DELETE", "/images", strings.NewReader(jsonBody))

			handler.ServeHTTP(response, req)

			Expect(imageDeleteFunc.CallCount()).To(Equal(1))

			ref, _ := imageDeleteFunc.ArgsForCall(0)
			expectedRef := fmt.Sprintf("%s/%s", registryBasePath, imageReference)
			Expect(ref.Name()).To(Equal(expectedRef))

			Expect(response.Code).To(Equal(http.StatusAccepted))
		})
	})

	DescribeTable("required fields are missing/blank",
		func(jsonBody string) {
			req := httptest.NewRequest("DELETE", "/images", strings.NewReader(jsonBody))

			handler.ServeHTTP(response, req)

			Expect(imageDeleteFunc.CallCount()).To(Equal(0))

			body, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("missing required parameter"))
			Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		},
		Entry("image_reference is missing", `{"registry_base_path": "`+registryBasePath+`"}`),
		Entry("registry_base_path is missing", `{"iamge_reference": "`+imageReference+`"}`),
	)

	When("the JSON is malformed", func() {
		BeforeEach(func() {
			jsonBody = `{`
		})

		It("returns a 400", func() {
			req := httptest.NewRequest("DELETE", "/images", strings.NewReader(jsonBody))

			handler.ServeHTTP(response, req)

			Expect(imageDeleteFunc.CallCount()).To(Equal(0))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			body, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("unable to parse request body"))
		})
	})

	When("the image reference is invalid", func() {
		BeforeEach(func() {
			jsonBody = `{
              "image_reference": ".",
              "registry_base_path": "."
            }`
		})

		It("returns a 422", func() {
			req := httptest.NewRequest("DELETE", "/images", strings.NewReader(jsonBody))

			handler.ServeHTTP(response, req)

			Expect(imageDeleteFunc.CallCount()).To(Equal(0))
			Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
			body, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("unable to parse image reference"))
		})
	})

	When("the upload errors", func() {
		BeforeEach(func() {
			jsonBody = `{
              "image_reference": "` + imageReference + `",
              "registry_base_path": "` + registryBasePath + `"
            }`
			imageDeleteFunc.Returns(errors.New("delete failed o no"))
		})

		It("returns a 500 error", func() {
			req := httptest.NewRequest("DELETE", "/images", strings.NewReader(jsonBody))

			handler.ServeHTTP(response, req)

			Expect(imageDeleteFunc.CallCount()).To(Equal(1))
			Expect(response.Code).To(Equal(http.StatusInternalServerError))
			body, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("unable to delete image"))
			Expect(body).To(ContainSubstring(imageReference))
		})
	})
})
