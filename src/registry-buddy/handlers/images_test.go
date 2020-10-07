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

	"github.com/google/go-containerregistry/pkg/v1/remote/transport"

	"code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/handlers"

	"github.com/google/go-containerregistry/pkg/authn"

	"code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/handlers/fakes"

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
		digestImageRef = "registry.example.com/cf-workloads/some-package@sha256:15e8a86a2cfce269dbc4321e741f729d60f41cafc7a8f7e11cd2892a4de3bf39"
		tagImageRef    = "registry.example.com/cf-workloads/some-package:some-tag"
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
		descriptor := remote.Descriptor{
			Descriptor: v1.Descriptor{
				Digest: v1.Hash{
					Algorithm: "sha256",
					Hex:       "15e8a86a2cfce269dbc4321e741f729d60f41cafc7a8f7e11cd2892a4de3bf39",
				},
			},
		}
		imageDescriptorFunc.Returns(&descriptor, nil)

		response = httptest.NewRecorder()
	})

	Context("successfully deleting an image", func() {
		BeforeEach(func() {
			response = httptest.NewRecorder()
		})

		When("image reference is a digest", func() {
			BeforeEach(func() {
				jsonBody = `{
              		"image_reference": "` + digestImageRef + `"
            	}`
			})

			It("deletes the image manifest from the registry by digest", func() {
				req := httptest.NewRequest("DELETE", "/images", strings.NewReader(jsonBody))

				handler.ServeHTTP(response, req)

				Expect(imageDescriptorFunc.CallCount()).To(Equal(1))
				Expect(imageDeleteFunc.CallCount()).To(Equal(1))

				ref, _ := imageDeleteFunc.ArgsForCall(0)
				Expect(ref.Name()).To(Equal(digestImageRef))

				Expect(response.Code).To(Equal(http.StatusAccepted))

				parsedBody := handlers.DeleteImageResponseBody{}
				Expect(
					json.NewDecoder(response.Body).Decode(&parsedBody),
				).To(Succeed())
				Expect(parsedBody).To(Equal(handlers.DeleteImageResponseBody{
					ImageReference: digestImageRef,
				}))
			})
		})

		When("image reference is a tag", func() {
			BeforeEach(func() {
				jsonBody = `{
              		"image_reference": "` + tagImageRef + `"
            	}`
			})

			It("deletes the image manifest from the registry by digest and tag", func() {
				req := httptest.NewRequest("DELETE", "/images", strings.NewReader(jsonBody))

				handler.ServeHTTP(response, req)

				Expect(imageDescriptorFunc.CallCount()).To(Equal(1))
				Expect(imageDeleteFunc.CallCount()).To(Equal(2))

				ref, _ := imageDeleteFunc.ArgsForCall(0)
				Expect(ref.Name()).To(Equal(tagImageRef))

				ref, _ = imageDeleteFunc.ArgsForCall(1)
				Expect(ref.Name()).To(Equal(digestImageRef))

				Expect(response.Code).To(Equal(http.StatusAccepted))

				parsedBody := handlers.DeleteImageResponseBody{}
				Expect(
					json.NewDecoder(response.Body).Decode(&parsedBody),
				).To(Succeed())
				Expect(parsedBody).To(Equal(handlers.DeleteImageResponseBody{
					ImageReference: digestImageRef,
				}))
			})
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

	When("fetching image metadata fails", func() {
		When("the registry returns a 404", func() {
			BeforeEach(func() {
				jsonBody = `{
              "image_reference": "` + digestImageRef + `"
            }`
				imageDescriptorFunc.Returns(nil, &transport.Error{
					StatusCode: 404,
				})
			})

			It("assumes the image has been deleted and returns a successful response", func() {
				req := httptest.NewRequest("DELETE", "/images", strings.NewReader(jsonBody))

				handler.ServeHTTP(response, req)

				Expect(response.Code).To(Equal(http.StatusAccepted))

				parsedBody := handlers.DeleteImageResponseBody{}
				Expect(
					json.NewDecoder(response.Body).Decode(&parsedBody),
				).To(Succeed())
				Expect(parsedBody).To(Equal(handlers.DeleteImageResponseBody{
					ImageReference: digestImageRef,
				}))
			})
		})

		When("the registry returns another error", func() {
			BeforeEach(func() {
				jsonBody = `{
              "image_reference": "` + digestImageRef + `"
            }`
				imageDescriptorFunc.Returns(nil, errors.New("delete failed o no"))
			})

			It("returns a 500 error", func() {
				req := httptest.NewRequest("DELETE", "/images", strings.NewReader(jsonBody))

				handler.ServeHTTP(response, req)

				Expect(imageDescriptorFunc.CallCount()).To(Equal(1))
				Expect(response.Code).To(Equal(http.StatusInternalServerError))
				body, err := ioutil.ReadAll(response.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(body).To(ContainSubstring("unable to fetch image metadata"))
				Expect(body).To(ContainSubstring(digestImageRef))
			})
		})
	})

	When("the delete errors", func() {
		When("the registry returns a 404", func() {
			BeforeEach(func() {
				jsonBody = `{
              "image_reference": "` + digestImageRef + `"
            }`

				imageDeleteFunc.Returns(&transport.Error{
					StatusCode: 404,
				})
			})

			It("assumes the image has been deleted and returns a successful response", func() {
				req := httptest.NewRequest("DELETE", "/images", strings.NewReader(jsonBody))

				handler.ServeHTTP(response, req)

				Expect(response.Code).To(Equal(http.StatusAccepted))

				parsedBody := handlers.DeleteImageResponseBody{}
				Expect(
					json.NewDecoder(response.Body).Decode(&parsedBody),
				).To(Succeed())
				Expect(parsedBody).To(Equal(handlers.DeleteImageResponseBody{
					ImageReference: digestImageRef,
				}))
			})
		})

		When("the registry returns another error", func() {
			BeforeEach(func() {
				jsonBody = `{
              "image_reference": "` + digestImageRef + `"
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
				Expect(body).To(ContainSubstring(digestImageRef))
			})
		})
	})
})
