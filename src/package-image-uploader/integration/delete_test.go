package integration_test

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/handlers"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/random"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deleting Images", func() {
	var (
		image      v1.Image
		imageRef   string
		ref        name.Reference
		client     *http.Client
		descriptor *remote.Descriptor
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

		descriptor, err = remote.Get(ref, remote.WithAuth(authenticator))
		Expect(err).NotTo(HaveOccurred())

		client = &http.Client{}
	})

	It("deletes an uploaded images by tag", func() {
		jsonBody := `{
              "image_reference": "` + imageRef + `"
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

		Eventually(func() error {
			_, err = remote.Get(ref, remote.WithAuth(authenticator))
			return err
		}, "5s", defaultPollingInterval).Should(MatchError(ContainSubstring("MANIFEST_UNKNOWN")))

		digestRefStr := fmt.Sprintf("%s@%s:%s", imageRef, descriptor.Digest.Algorithm, descriptor.Digest.Hex)
		digestRef, err := name.ParseReference(digestRefStr)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			_, err = remote.Get(digestRef, remote.WithAuth(authenticator))
			return err
		}, "5s", defaultPollingInterval).Should(MatchError(ContainSubstring("MANIFEST_UNKNOWN")))
	})
})

func randomSuffix() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
