package handlers_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/google/go-containerregistry/pkg/v1/remote"

	"code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/handlers"

	"github.com/google/go-containerregistry/pkg/authn"

	"code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/handlers/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeleteImageHandler", func() {
	var (
		handler             http.HandlerFunc
		imageDeleteFunc     *fakes.ImageDeleteFunc
		imageDescriptorFunc *fakes.ImageDescriptorFunc
		response            *httptest.ResponseRecorder
		authenticator       authn.Authenticator
		jsonBody            string
	)

	const (
		imageReference = "registry.example.com/cf-workloads/some-package@sha256:15e8a86a2cfce269dbc4321e741f729d60f41cafc7a8f7e11cd2892a4de3bf39"
	)

	BeforeEach(func() {
		imageDeleteFunc = new(fakes.ImageDeleteFunc)
		imageDescriptorFunc = new(fakes.ImageDescriptorFunc)
		logger := log.New(GinkgoWriter, "", 0)
		authenticator = authn.FromConfig(authn.AuthConfig{
			Username: "some-user",
			Password: "some-password",
		})
		handler = handlers.DeleteImageHandler(imageDeleteFunc.Spy, imageDescriptorFunc.Spy, logger, authenticator)
		response = httptest.NewRecorder()
	})

	Context("successfully deleting an image", func() {
		BeforeEach(func() {
			jsonBody = `{
              "image_reference": "` + imageReference + `"
            }`

			response = httptest.NewRecorder()

			descriptor := remote.Descriptor{
				Descriptor: v1.Descriptor{
					Digest: v1.Hash{
						Algorithm: "sha256",
						Hex:       "15e8a86a2cfce269dbc4321e741f729d60f41cafc7a8f7e11cd2892a4de3bf39",
					},
				},
			}
			imageDescriptorFunc.Returns(&descriptor, nil)
		})

		It("deletes the image from the registry", func() {
			req := httptest.NewRequest("DELETE", "/images", strings.NewReader(jsonBody))

			handler.ServeHTTP(response, req)

			Expect(imageDescriptorFunc.CallCount()).To(Equal(1))
			Expect(imageDeleteFunc.CallCount()).To(Equal(2))

			ref, _ := imageDeleteFunc.ArgsForCall(0)
			Expect(ref.Name()).To(Equal(imageReference))

			Expect(response.Code).To(Equal(http.StatusAccepted))

			parsedBody := handlers.DeleteImageResponseBody{}
			Expect(
				json.NewDecoder(response.Body).Decode(&parsedBody),
			).To(Succeed())
			Expect(parsedBody).To(Equal(handlers.DeleteImageResponseBody{
				ImageReference: imageReference,
			}))
		})
	})

	DescribeTable("required fields are missing/blank",
		func(jsonBody string) {
			req := httptest.NewRequest("DELETE", "/images", strings.NewReader(jsonBody))

			handler.ServeHTTP(response, req)

			Expect(imageDeleteFunc.CallCount()).To(Equal(0))

			body, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(body)).To(ContainSubstring("missing required parameter"))
			Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
		},
		Entry("image_reference is missing", `{"some_other_field": "some-value"}`),
		Entry("request body is empty", `{}`),
	)

	When("the request JSON is malformed", func() {
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
              "image_reference": "."
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
              "image_reference": "` + imageReference + `"
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
