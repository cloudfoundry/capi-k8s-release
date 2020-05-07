package image_registry

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/ghttp"
	"github.com/sclevine/spec"
)

func TestImageConfigFetcher(t *testing.T) {
	spec.Run(t, "TestImageConfigFetcher", func(t *testing.T, when spec.G, it spec.S) {
		it.Before(func() {
			RegisterTestingT(t)
			format.TruncatedDiff = false
		})

		when("fetching an OCI Image Config", func() {
			when("supplying a valid image reference stored in a public registry", func() {
				var (
					fetcher     ImageConfigFetcher
					imageConfig *v1.Config
					err         error
				)

				it.Before(func() {
					fetcher = NewImageConfigFetcher()
				})

				it("returns a valid, expected OCI Image Config", func() {
					imageConfig, err = fetcher.FetchImageConfig("busybox@sha256:a2490cec4484ee6c1068ba3a05f89934010c85242f736280b35343483b2264b6")

					Expect(err).ToNot(HaveOccurred())
					Expect(imageConfig).ToNot(BeNil())

					Expect(imageConfig.Image).To(Equal("sha256:b0acc7ebf5092fcdd0fe097448529147e6619bd051f03ccf25b29bcae87e783f"))
					Expect(imageConfig.Cmd).To(ConsistOf("sh"))
					Expect(imageConfig.Env).To(ConsistOf("PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"))
				})
			})

			when("supplying a valid image reference stored in a private registry", func() {
				var (
					fetcher            ImageConfigFetcher
					imageConfig        *v1.Config
					err                error
					fakeRegistryServer *ghttp.Server
				)

				it.Before(func() {
					fetcher = NewImageConfigFetcher()
					// TODO: setup Ginkgo mock HTTP server to return mock Image Config response
					fakeRegistryServer = ghttp.NewServer()
					fakeRegistryServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v2/"),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v2/busybox/manifests/sha256:e4a4e4e4601cb82a7361cf11a0efad7eb3a05022f710a581454e22e78e30181c"),
							ghttp.RespondWith(http.StatusOK, `{
        "schemaVersion": 2,
        "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
        "config": {
                "mediaType": "application/vnd.docker.container.image.v1+json",
                "size": 1494,
                "digest": "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
        },
        "layers": [
                {
                        "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
                        "size": 760854,
                        "digest": "sha256:e2334dd9fee4b77e48a8f2d793904118a3acf26f1f2e72a3d79c6cae993e07f0"
                }
        ]
}`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v2/busybox/blobs/sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"),
						),
					)
				})

				it.After(func() {
					fakeRegistryServer.Close()
				})

				it("returns a valid, expected OCI Image Config", func() {
					registryDomain := strings.TrimPrefix(fakeRegistryServer.URL(), `http://`)
					imageConfig, err = fetcher.FetchImageConfig(fmt.Sprintf("%s/busybox@sha256:e4a4e4e4601cb82a7361cf11a0efad7eb3a05022f710a581454e22e78e30181c", registryDomain))

					Expect(err).ToNot(HaveOccurred())
					Expect(imageConfig).ToNot(BeNil())

					Expect(imageConfig.Image).To(Equal("sha256:b0acc7ebf5092fcdd0fe097448529147e6619bd051f03ccf25b29bcae87e783f"))
					Expect(imageConfig.Cmd).To(ConsistOf("sh"))
					Expect(imageConfig.Env).To(ConsistOf("PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"))
				})
			})
		})
	})
}
