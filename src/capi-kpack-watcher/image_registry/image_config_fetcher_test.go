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
							ghttp.VerifyRequest("GET", "/v2/busybox/manifests/sha256:3d234de17ec9ceb63617e4eed1a8d3bfd6ceb1bb94ae6bd5b7b820041ffd7aa2"),
							ghttp.RespondWith(http.StatusOK, `{
        "schemaVersion": 2,
        "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
        "config": {
                "mediaType": "application/vnd.docker.container.image.v1+json",
                "size": 1494,
                "digest": "sha256:be68ed756f4cae7f30097b64ab89281b6693a48c001fc7f4b00704ecdc5e9aab"
        },
        "layers": [
                {
                        "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
                        "size": 760854,
                        "digest": "sha256:be68ed756f4cae7f30097b64ab89281b6693a48c001fc7f4b00704ecdc5e9aab"
                }
        ]
}`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v2/busybox/blobs/sha256:be68ed756f4cae7f30097b64ab89281b6693a48c001fc7f4b00704ecdc5e9aab"),
							ghttp.RespondWith(http.StatusOK, `{
        "schemaVersion": 2,
		"image": "some/image",
        "mediaType": "application/vnd.docker.distribution.blob.v2+json",
        "config": {
                "mediaType": "application/vnd.docker.container.image.v1+json",
                "size": 1494,
                "digest": "sha256:8c0a103553d9c86d5d6b628f0af2c354a23dded270f3f5b090b08e34d6845adf"
        },
        "layers": [
                {
                        "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
                        "size": 760854,
                        "digest": "sha256:8c0a103553d9c86d5d6b628f0af2c354a23dded270f3f5b090b08e34d6845adf"
                }
        ]
}`),
						),
					)
				})

				it.After(func() {
					fakeRegistryServer.Close()
				})

				it("returns a valid, expected OCI Image Config", func() {
					registryDomain := strings.TrimPrefix(fakeRegistryServer.URL(), `http://`)
					imageConfig, err = fetcher.FetchImageConfig(fmt.Sprintf("%s/busybox@sha256:3d234de17ec9ceb63617e4eed1a8d3bfd6ceb1bb94ae6bd5b7b820041ffd7aa2", registryDomain))

					Expect(err).ToNot(HaveOccurred())
					Expect(imageConfig).To(BeNil())


					Expect(imageConfig.Domainname).To(Equal("sha256:b0acc7ebf5092fcdd0fe097448529147e6619bd051f03ccf25b29bcae87e783f"))
					Expect(imageConfig.Cmd).To(ConsistOf("sh"))
					Expect(imageConfig.Env).To(ConsistOf("PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"))
				})
			})
		})
	})
}
