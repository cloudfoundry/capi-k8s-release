package image_test

import (
	"code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/image"
	"code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/image/fakes"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"log"
)

var _ = Describe("NewGenericDeleter returned function", func() {
	var (
		imageDeleter        func(reference name.Reference, authenticator authn.Authenticator, logger *log.Logger) error
		imageRef            name.Reference
		rawImageDeleteFunc  *fakes.RawImageDeleteFunc
		imageDescriptorFunc *fakes.ImageDescriptorFunc
		authenticator       authn.Authenticator
		logger              *log.Logger
	)

	const (
		digestImageRef = "registry.example.com/cf-workloads/some-package@sha256:15e8a86a2cfce269dbc4321e741f729d60f41cafc7a8f7e11cd2892a4de3bf39"
		tagImageRef    = "registry.example.com/cf-workloads/some-package:some-tag"
	)

	BeforeEach(func() {
		rawImageDeleteFunc = new(fakes.RawImageDeleteFunc)
		imageDescriptorFunc = new(fakes.ImageDescriptorFunc)
		logger = log.New(GinkgoWriter, "", 0)
		authenticator = authn.FromConfig(authn.AuthConfig{
			Username: "some-user",
			Password: "some-password",
		})
		descriptor := remote.Descriptor{
			Descriptor: v1.Descriptor{
				Digest: v1.Hash{
					Algorithm: "sha256",
					Hex:       "15e8a86a2cfce269dbc4321e741f729d60f41cafc7a8f7e11cd2892a4de3bf39",
				},
			},
		}
		imageDescriptorFunc.Returns(&descriptor, nil)

		imageDeleter = image.NewGenericDeleter(rawImageDeleteFunc.Spy, imageDescriptorFunc.Spy)
	})

	When("image reference is a digest", func() {
		BeforeEach(func() {
			var err error
			imageRef, err = name.ParseReference(digestImageRef)
			Expect(err).NotTo(HaveOccurred())
		})

		It("deletes the image manifest from the registry by digest", func() {
			Expect(imageDeleter(imageRef, authenticator, logger)).To(Succeed())

			Expect(imageDescriptorFunc.CallCount()).To(Equal(1))
			Expect(rawImageDeleteFunc.CallCount()).To(Equal(1))

			ref, _ := rawImageDeleteFunc.ArgsForCall(0)
			Expect(ref.Name()).To(Equal(digestImageRef))
		})

		When("delete fails with a 404", func() {
			BeforeEach(func() {
				rawImageDeleteFunc.Returns(
					&transport.Error{StatusCode: 404},
				)
			})

			It("succeeds", func() {
				Expect(imageDeleter(imageRef, authenticator, logger)).To(Succeed())
			})
		})

		When("delete fails with a non 404", func() {
			BeforeEach(func() {
				rawImageDeleteFunc.Returns(
					&transport.Error{StatusCode: 500},
				)
			})

			It("errors", func() {
				err := imageDeleter(imageRef, authenticator, logger)
				Expect(err).To(MatchError(ContainSubstring("500")))
			})
		})
	})

	When("image reference is a tag", func() {
		BeforeEach(func() {
			var err error
			imageRef, err = name.ParseReference(tagImageRef)
			Expect(err).NotTo(HaveOccurred())
		})

		It("deletes the image manifest from the registry by digest and tag", func() {
			Expect(imageDeleter(imageRef, authenticator, logger)).To(Succeed())

			Expect(imageDescriptorFunc.CallCount()).To(Equal(1))
			Expect(rawImageDeleteFunc.CallCount()).To(Equal(2))

			ref, _ := rawImageDeleteFunc.ArgsForCall(0)
			Expect(ref.Name()).To(Equal(tagImageRef))

			ref, _ = rawImageDeleteFunc.ArgsForCall(1)
			Expect(ref.Name()).To(Equal(digestImageRef))
		})

		When("delete the image by tag fails with a 404", func() {
			BeforeEach(func() {
				rawImageDeleteFunc.ReturnsOnCall(0,
					&transport.Error{StatusCode: 404},
				)
				rawImageDeleteFunc.ReturnsOnCall(1, nil)
			})

			It("still deletes by ref", func() {
				Expect(imageDeleter(imageRef, authenticator, logger)).To(Succeed())

				Expect(rawImageDeleteFunc.CallCount()).To(Equal(2))

				ref, _ := rawImageDeleteFunc.ArgsForCall(0)
				Expect(ref.Name()).To(Equal(tagImageRef))

				ref, _ = rawImageDeleteFunc.ArgsForCall(1)
				Expect(ref.Name()).To(Equal(digestImageRef))
			})
		})

		When("delete the image by tag fails with an unsupported error", func() {
			BeforeEach(func() {
				rawImageDeleteFunc.ReturnsOnCall(0,
					&transport.Error{
						Errors: []transport.Diagnostic{
							{Code: transport.UnsupportedErrorCode},
						},
					},
				)
				rawImageDeleteFunc.ReturnsOnCall(1, nil)
			})

			It("still deletes by ref", func() {
				Expect(imageDeleter(imageRef, authenticator, logger)).To(Succeed())

				Expect(rawImageDeleteFunc.CallCount()).To(Equal(2))

				ref, _ := rawImageDeleteFunc.ArgsForCall(0)
				Expect(ref.Name()).To(Equal(tagImageRef))

				ref, _ = rawImageDeleteFunc.ArgsForCall(1)
				Expect(ref.Name()).To(Equal(digestImageRef))
			})
		})

		When("deleting the image by tag fails with a non 404", func() {
			BeforeEach(func() {
				rawImageDeleteFunc.Returns(
					&transport.Error{StatusCode: 500},
				)
			})

			It("errors and makes no further calls", func() {
				err := imageDeleter(imageRef, authenticator, logger)
				Expect(err).To(MatchError(ContainSubstring("500")))

				Expect(rawImageDeleteFunc.CallCount()).To(Equal(1))

				ref, _ := rawImageDeleteFunc.ArgsForCall(0)
				Expect(ref.Name()).To(Equal(tagImageRef))
			})
		})
	})

	When("fetching the image information fails with a 404", func() {
		BeforeEach(func() {
			imageDescriptorFunc.Returns(nil, &transport.Error{StatusCode: 404})
		})

		It("succeeds and makes no further calls", func() {
			Expect(imageDeleter(imageRef, authenticator, logger)).To(Succeed())

			Expect(rawImageDeleteFunc.CallCount()).To(Equal(0))
		})
	})

	When("fetching the image information fails with a non-404", func() {
		BeforeEach(func() {
			imageDescriptorFunc.Returns(nil, &transport.Error{StatusCode: 500})
		})

		It("returns an error and makes no further calls", func() {
			err := imageDeleter(imageRef, authenticator, logger)
			Expect(err).To(MatchError(ContainSubstring("500")))

			Expect(rawImageDeleteFunc.CallCount()).To(Equal(0))
		})
	})
})
