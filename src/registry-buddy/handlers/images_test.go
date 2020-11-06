package handlers_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	"code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/handlers"

	"github.com/google/go-containerregistry/pkg/authn"

	imageFakes "code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/image/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeleteImageHandler", func() {
	var (
		handler       http.HandlerFunc
		imageDeleter  *imageFakes.Deleter
		response      *httptest.ResponseRecorder
		authenticator authn.Authenticator
		jsonBody      string
		logger        *log.Logger
	)

	const (
		tagImageRef    = "registry.example.com/cf-workloads/some-package:some-tag"
	)

	BeforeEach(func() {
		logger = log.New(GinkgoWriter, "", 0)
		authenticator = authn.FromConfig(authn.AuthConfig{
			Username: "some-user",
			Password: "some-password",
		})
		imageDeleter = new(imageFakes.Deleter)
		handler = handlers.DeleteImageHandler(imageDeleter.Spy, logger, authenticator)

		response = httptest.NewRecorder()
		jsonBody = `{"image_reference": "` + tagImageRef + `"}`
	})

	Context("successfully deleting an image", func() {
		BeforeEach(func() {
			response = httptest.NewRecorder()
		})

		It("deletes the image manifest from the registry by digest", func() {
			req := httptest.NewRequest("DELETE", "/images", strings.NewReader(jsonBody))

			handler.ServeHTTP(response, req)
			Expect(response.Code).To(Equal(http.StatusAccepted))

			parsedBody := handlers.DeleteImageResponseBody{}
			Expect(
				json.NewDecoder(response.Body).Decode(&parsedBody),
			).To(Succeed())
			Expect(parsedBody).To(Equal(handlers.DeleteImageResponseBody{
				ImageReference: tagImageRef,
			}))

			Expect(imageDeleter.CallCount()).To(Equal(1))

			ref, actualAuth, actualLogger := imageDeleter.ArgsForCall(0)
			Expect(ref.Name()).To(Equal(tagImageRef))
			Expect(actualAuth).To(Equal(authenticator))
			Expect(actualLogger).To(Equal(logger))
		})
	})

	DescribeTable("required fields are missing/blank",
		func(jsonBody string) {
			req := httptest.NewRequest("DELETE", "/images", strings.NewReader(jsonBody))

			handler.ServeHTTP(response, req)

			Expect(imageDeleter.CallCount()).To(Equal(0))

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

			Expect(imageDeleter.CallCount()).To(Equal(0))
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

			Expect(imageDeleter.CallCount()).To(Equal(0))
			Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
			body, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("unable to parse image reference"))
		})
	})

	When("the delete errors", func() {
		BeforeEach(func() {
			jsonBody = `{
              "image_reference": "` + tagImageRef + `"
            }`
			imageDeleter.Returns(errors.New("delete failed o no"))
		})

		It("returns a 500 error", func() {
			req := httptest.NewRequest("DELETE", "/images", strings.NewReader(jsonBody))

			handler.ServeHTTP(response, req)

			Expect(imageDeleter.CallCount()).To(Equal(1))
			Expect(response.Code).To(Equal(http.StatusInternalServerError))
			body, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("unable to delete image"))
			Expect(body).To(ContainSubstring(tagImageRef))
		})
	})
})
