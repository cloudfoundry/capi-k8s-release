package image_test

import (
	. "code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/dockerhub"
	"code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/image"
	"code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/image/fakes"
	"errors"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"log"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fakes/authenticator.go --fake-name Authenticator github.com/google/go-containerregistry/pkg/authn.Authenticator

var _ = Describe("NewDockerhubDeleter returned func", func() {
	var (
		deleter func(name.Reference, authn.Authenticator, *log.Logger) error
		client *fakes.DockerhubClient
		imageRef name.Reference
		auth authn.Authenticator
		logger *log.Logger
	)

	const (
		token = "my-token"
		username = "bob"
		password = "$3kr1t"
		repoName = "my-org/some-package-guid"
	)

	BeforeEach(func() {
		client = new(fakes.DockerhubClient)
		deleter = image.NewDockerhubDeleter(client)

		var err error
		imageRef, err = name.ParseReference(repoName)
		Expect(err).NotTo(HaveOccurred())

		logger = log.New(GinkgoWriter, "", log.LstdFlags)
		auth = authn.FromConfig(authn.AuthConfig{
			Username: username,
			Password: password,
		})
	})

	When("all requests are successful", func() {
		BeforeEach(func() {
			client.GetAuthorizationTokenReturns(token, nil)
			client.DeleteRepoReturns(nil)
		})

		It("fetches an auth token", func() {
			Expect(deleter(imageRef, auth, logger)).To(Succeed())

			Expect(client.GetAuthorizationTokenCallCount()).To(Equal(1))
			actualUsername, actualPassword := client.GetAuthorizationTokenArgsForCall(0)
			Expect(actualUsername).To(Equal(username))
			Expect(actualPassword).To(Equal(password))
		})

		It("deletes the image", func() {
			Expect(deleter(imageRef, auth, logger)).To(Succeed())

			Expect(client.DeleteRepoCallCount()).To(Equal(1))
			actualRepoName, actualToken := client.DeleteRepoArgsForCall(0)
			Expect(actualRepoName).To(Equal(repoName))
			Expect(actualToken).To(Equal(token))
		})
	})

	When("there's an error fetching the authentication", func() {
		BeforeEach(func() {
			auth = new(fakes.Authenticator)
			auth.(*fakes.Authenticator).AuthorizationReturns(nil,
				errors.New("auth no good"),
			)
		})

		It("returns an error", func() {
			err := deleter(imageRef, auth, logger)
			Expect(err).To(MatchError(ContainSubstring("auth no good")))

			Expect(client.GetAuthorizationTokenCallCount()).To(Equal(0))
			Expect(client.DeleteRepoCallCount()).To(Equal(0))
		})
	})

	When("there's an error fetching the token", func() {
		BeforeEach(func() {
			client.GetAuthorizationTokenReturns("", errors.New("boom"))
		})

		It("returns an error", func() {
			err := deleter(imageRef, auth, logger)
			Expect(err).To(MatchError(ContainSubstring("boom")))

			Expect(client.GetAuthorizationTokenCallCount()).To(Equal(1))
			Expect(client.DeleteRepoCallCount()).To(Equal(0))
		})
	})

	When("deleting the repo yields a NotFoundError", func() {
		BeforeEach(func() {
			client.DeleteRepoReturns(&NotFoundError{Err: errors.New("no such repo")})
		})

		It("returns an error", func() {
			Expect(deleter(imageRef, auth, logger)).To(Succeed())

			Expect(client.GetAuthorizationTokenCallCount()).To(Equal(1))
			Expect(client.DeleteRepoCallCount()).To(Equal(1))
		})
	})

	When("deleting the repo yields an error other than NotFound", func() {
		BeforeEach(func() {
			client.DeleteRepoReturns(errors.New("boom"))
		})

		It("returns an error", func() {
			err := deleter(imageRef, auth, logger)
			Expect(err).To(MatchError(ContainSubstring("boom")))

			Expect(client.GetAuthorizationTokenCallCount()).To(Equal(1))
			Expect(client.DeleteRepoCallCount()).To(Equal(1))
		})
	})
})
