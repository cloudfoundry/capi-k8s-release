package image_registry

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func TestManifestFetcher(t *testing.T) {
	spec.Run(t, "TestManifestFetcher", func(t *testing.T, when spec.G, it spec.S) {
		it.Before(func() {
			RegisterTestingT(t)
		})

		when("fetching a Docker image manifest", func() {
			when("supplying a valid image reference", func() {
				var (
					fetcher     ManifestFetcher
					rawManifest []byte
					err         error
				)

				it.Before(func() {
					fetcher = NewManifestFetcher(DockerManifestType)
					rawManifest, err = fetcher.FetchRawManifestFromImageReference("busybox@sha256:a2490cec4484ee6c1068ba3a05f89934010c85242f736280b35343483b2264b6")
				})

				it("a valid OCI image manifest is returned", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(rawManifest).ToNot(BeNil())

					Expect(string(rawManifest)).To(Equal(`{
   "schemaVersion": 2,
   "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
   "config": {
      "mediaType": "application/vnd.docker.container.image.v1+json",
      "size": 1494,
      "digest": "sha256:be5888e67be651f1fbb59006f0fd791b44ed3fceaa6323ab4e37d5928874345a"
   },
   "layers": [
      {
         "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
         "size": 760854,
         "digest": "sha256:e2334dd9fee4b77e48a8f2d793904118a3acf26f1f2e72a3d79c6cae993e07f0"
      }
   ]
}`))
				})
			})
		})
	})
}
