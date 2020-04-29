package oci_registry

import (
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func TestManifestFetcher(t *testing.T) {
	spec.Run(t, "TestManifestFetcher", func(t *testing.T, when spec.G, it spec.S) {
		it.Before(func() {
			RegisterTestingT(t)
		})

		when("fetching a manifest with a valid image reference string", func() {
			var (
				fetcher  ManifestFetcher
				manifest *v1.Manifest
				err      error
			)

			it.Before(func() {
				fetcher = NewManifestFetcher()
				manifest, err = fetcher.FetchManifestFromImageReference("cloudfoundry/cloud-controller-ng:1ebab1cbb5270a3d51c0a098a37cd9e8ca0f721d@sha256:374f967edd7db4d7efc2f38cb849988aa36a8248dd240d56f49484b8159fd800")
			})

			it("a valid OCI image manifest is returned", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(manifest).ToNot(BeNil())

				Expect(manifest.Config.Annotations).ToNot(BeNil())
				Expect(manifest.Config.Annotations).To(HaveKeyWithValue("io.buildpacks.build.metadata", "TODO-some-big-raw-json-blob"))
			})
		})
	})
}
