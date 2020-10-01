package handlers_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	"code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/handlers/fakes"

	. "code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/handlers"
	"code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/upload"
	"github.com/google/go-containerregistry/pkg/authn"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("PostPackageHandler", func() {
	var (
		uploaderFunc  *fakes.UploaderFunc
		handler       http.HandlerFunc
		response      *httptest.ResponseRecorder
		authenticator authn.Authenticator
		jsonBody      string
	)

	const (
		packageGuid      = "package-guid"
		packageZipPath   = "/path/to/some_appbits.zip"
		registryBasePath = "registry.example.com/example-registry"
	)

	BeforeEach(func() {
		uploaderFunc = new(fakes.UploaderFunc)
		logger := log.New(GinkgoWriter, "", 0)
		authenticator = authn.FromConfig(authn.AuthConfig{
			Username: "some-user",
			Password: "some-password",
		})
		handler = PostPackageHandler(uploaderFunc.Spy, logger, authenticator)
		response = httptest.NewRecorder()
	})

	When("happy path", func() {
		const (
			algorithm = "sha256"
			hex       = "my-awesome-sha"
		)

		BeforeEach(func() {
			jsonBody = `{
              "package_zip_path": "` + packageZipPath + `",
              "package_guid": "` + packageGuid + `",
              "registry_base_path": "` + registryBasePath + `"
            }`

			uploaderFunc.Returns(upload.Hash{Algorithm: algorithm, Hex: hex}, nil)
		})

		It("uploads the package to the registry", func() {
			req := httptest.NewRequest("POST", "/packages", strings.NewReader(jsonBody))

			handler.ServeHTTP(response, req)

			Expect(uploaderFunc.CallCount()).To(Equal(1))

			zipPath, registryPath, actualAuthenticator := uploaderFunc.ArgsForCall(0)
			Expect(zipPath).To(Equal(packageZipPath))
			Expect(registryPath).To(Equal(registryBasePath + "/" + packageGuid))
			Expect(actualAuthenticator).To(Equal(authenticator))

			Expect(response.Code).To(Equal(200))

			parsedBody := PostPackageResponse{}
			Expect(
				json.NewDecoder(response.Body).Decode(&parsedBody),
			).To(Succeed())
			Expect(parsedBody).To(Equal(PostPackageResponse{
				Hash: HashResponse{
					Algorithm: algorithm,
					Hex:       hex,
				},
			}))
		})
	})

	DescribeTable("required fields are missing/blank",
		func(jsonBody string) {
			req := httptest.NewRequest("POST", "/packages", strings.NewReader(jsonBody))

			handler.ServeHTTP(response, req)

			Expect(uploaderFunc.CallCount()).To(Equal(0))

			body, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("missing required parameter"))
			Expect(response.Code).To(Equal(422))
		},
		Entry("package_zip_path is missing", `{"package_guid": "`+packageGuid+`", "registry_base_path": "`+registryBasePath+`"}`),
		Entry("package_guid is missing", `{"package_zip_path": "`+packageZipPath+`", "registry_base_path": "`+registryBasePath+`"}`),
		Entry("registry_base_path is missing", `{"package_guid": "`+packageGuid+`", "package_zip_path": "`+packageZipPath+`"}`),
	)

	When("the JSON is malformed", func() {
		BeforeEach(func() {
			jsonBody = `{`
		})

		It("returns a 400", func() {
			req := httptest.NewRequest("POST", "/packages", strings.NewReader(jsonBody))

			handler.ServeHTTP(response, req)

			Expect(uploaderFunc.CallCount()).To(Equal(0))
			Expect(response.Code).To(Equal(400))
			body, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("unable to parse request body"))
		})
	})

	When("the upload errors", func() {
		BeforeEach(func() {
			jsonBody = `{
              "package_zip_path": "` + packageZipPath + `",
              "package_guid": "` + packageGuid + `",
              "registry_base_path": "` + registryBasePath + `"
            }`
			uploaderFunc.Returns(upload.Hash{}, errors.New("upload failed o no"))
		})

		It("returns a 500 error", func() {
			req := httptest.NewRequest("POST", "/packages", strings.NewReader(jsonBody))

			handler.ServeHTTP(response, req)

			Expect(uploaderFunc.CallCount()).To(Equal(1))
			Expect(response.Code).To(Equal(500))
			body, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("unable to convert/upload package"))
		})
	})
})
