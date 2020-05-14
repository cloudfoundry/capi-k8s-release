package image_registry

import (
	"fmt"
	"github.com/pivotal/kpack/pkg/registry"
	"github.com/pivotal/kpack/pkg/registry/registryfakes"

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
					imageConfig, err = fetcher.FetchImageConfig("busybox@sha256:a2490cec4484ee6c1068ba3a05f89934010c85242f736280b35343483b2264b6", "", "")

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
					keychainFactory = &registryfakes.FakeKeychainFactory{}
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
							ghttp.VerifyRequest("GET", "/v2/busybox/manifests/sha256:4bc6920026921689d030c4dcb3f960cb5bdd5883dbe4622ae1f2d2accae3c0fd"),
							ghttp.RespondWith(http.StatusOK, `{
        "schemaVersion": 2,
        "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
        "config": {
                "mediaType": "application/vnd.docker.container.image.v1+json",
                "size": 1494,
                "digest": "sha256:ad0ac93746366a1c56e3c5e41910ccf4d15b678c1835dc9fb1ae2edd4b496596"
        },
        "layers": [
                {
                        "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
                        "size": 760854,
                        "digest": "sha256:ad0ac93746366a1c56e3c5e41910ccf4d15b678c1835dc9fb1ae2edd4b496596"
                }
        ]
}`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v2/busybox/blobs/sha256:ad0ac93746366a1c56e3c5e41910ccf4d15b678c1835dc9fb1ae2edd4b496596"),
							ghttp.RespondWith(http.StatusOK, `{
        "schemaVersion": 2,
        "mediaType": "application/vnd.docker.distribution.blob.v2+json",
        "config": {
				"image": "sha256:b0acc7ebf5092fcdd0fe097448529147e6619bd051f03ccf25b29bcae87e783f",
				"cmd": ["sh"],
				"env": ["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],
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
					appImageSecretRef := registry.SecretRef{
						ServiceAccount: "build-service-account",
						Namespace:      "build-namespace",
					}
					appImageKeychain := &registryfakes.FakeKeychain{}
					keychainFactory.AddKeychainForSecretRef(t, appImageSecretRef, appImageKeychain)
					registryDomain := strings.TrimPrefix(fakeRegistryServer.URL(), `http://`)
					imageConfig, err = fetcher.FetchImageConfig(fmt.Sprintf("%s/busybox@sha256:4bc6920026921689d030c4dcb3f960cb5bdd5883dbe4622ae1f2d2accae3c0fd", registryDomain),"build-service-account", "build-namespace")

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
