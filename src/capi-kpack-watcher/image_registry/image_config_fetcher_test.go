package image_registry

import (
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func TestImageConfigFetcher(t *testing.T) {
	spec.Run(t, "TestImageConfigFetcher", func(t *testing.T, when spec.G, it spec.S) {
		it.Before(func() {
			RegisterTestingT(t)
		})

		when("fetching an OCI Image Config", func() {
			when("supplying a valid image reference", func() {
				var (
					fetcher     ImageConfigFetcher
					imageConfig *v1.Config
					err         error
				)

				it.Before(func() {
					fetcher = NewImageConfigFetcher()
					// TODO: setup Ginkgo mock HTTP server to return mock Image Config response
					imageConfig, err = fetcher.FetchImageConfig("busybox@sha256:a2490cec4484ee6c1068ba3a05f89934010c85242f736280b35343483b2264b6")
				})

				it("returns a valid, expected OCI Image Config", func() {
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
