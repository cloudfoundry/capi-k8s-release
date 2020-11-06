package integration_test

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/handlers"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/random"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deleting Images", func() {
	var (
		image     v1.Image
		imageRef  string
		ref       name.Reference
		digestRef name.Reference
		client    *http.Client
	)

	BeforeEach(func() {
		var err error
		image, err = random.Image(1024, 1)
		Expect(err).NotTo(HaveOccurred())

		hex, err := randomSuffix()
		Expect(err).NotTo(HaveOccurred())
		randImageName := fmt.Sprintf("test-image-%s", hex)

		imageRef = fmt.Sprintf("%s/%s", registryBasePath, randImageName)

		ref, err = name.ParseReference(imageRef)
		Expect(err).NotTo(HaveOccurred())

		err = remote.Write(ref, image, remote.WithAuth(authenticator))
		Expect(err).NotTo(HaveOccurred())

		descriptor, err := remote.Get(ref, remote.WithAuth(authenticator))
		Expect(err).NotTo(HaveOccurred())

		digestRefStr := fmt.Sprintf("%s@%s:%s", imageRef, descriptor.Digest.Algorithm, descriptor.Digest.Hex)
		digestRef, err = name.ParseReference(digestRefStr)
		Expect(err).NotTo(HaveOccurred())

		client = &http.Client{}
	})

	It("deletes image manifests for both the image digest and tag", func() {
		jsonBody := `{
              "image_reference": "` + ref.Name() + `"
            }`
		req, err := http.NewRequest("DELETE", testServer.URL+"/images", strings.NewReader(jsonBody))
		Expect(err).NotTo(HaveOccurred())
		res, err := client.Do(req)
		Expect(err).NotTo(HaveOccurred())

		Expect(res.StatusCode).To(Equal(http.StatusAccepted))
		parsedBody := handlers.DeleteImageResponseBody{}
		Expect(
			json.NewDecoder(res.Body).Decode(&parsedBody),
		).To(Succeed())
		Expect(parsedBody).To(Equal(handlers.DeleteImageResponseBody{
			ImageReference: ref.Name(),
		}))

		// The manifest for the image tag is deleted
		Eventually(func() int {
			// calling Get returns image metadata if the image manifest exists
			// it returns an error if the image manifest has been deleted
			_, err = remote.Get(ref, remote.WithAuth(authenticator))
			if e, ok := err.(*transport.Error); ok && e.StatusCode == 404 {
				return e.StatusCode
			}
			return 0
		}, "5s", defaultPollingInterval).Should(Equal(404))

		// The manifest for the image digest is deleted
		Eventually(func() int {
			// calling Get returns image metadata if the image manifest exists
			// it returns an error if the image manifest has been deleted
			_, err = remote.Get(digestRef, remote.WithAuth(authenticator))
			if e, ok := err.(*transport.Error); ok && e.StatusCode == 404 {
				return e.StatusCode
			}
			return 0
		}, "5s", defaultPollingInterval).Should(Equal(404))
	})
})

func randomSuffix() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
