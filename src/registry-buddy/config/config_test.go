package config_test

import (
	"os"

	"code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Describe("Load", func() {
		const (
			regBasePath = "example.com/some-repo"
			regUsername = "some-username"
			regPassword = "some-password"
		)

		BeforeEach(func() {
			err := os.Setenv("REGISTRY_BASE_PATH", regBasePath)
			Expect(err).NotTo(HaveOccurred())
			err = os.Setenv("REGISTRY_USERNAME", regUsername)
			Expect(err).NotTo(HaveOccurred())
			err = os.Setenv("REGISTRY_PASSWORD", regPassword)
			Expect(err).NotTo(HaveOccurred())
			err = os.Setenv("PORT", "9876")
			Expect(err).NotTo(HaveOccurred())
		})

		It("loads the config", func() {
			cfg, err := config.Load()
			Expect(err).NotTo(HaveOccurred())

			Expect(cfg.RegistryBasePath).To(Equal(regBasePath))
			Expect(cfg.RegistryUsername).To(Equal(regUsername))
			Expect(cfg.RegistryPassword).To(Equal(regPassword))
			Expect(cfg.Port).To(Equal(9876))
		})

		Context("when the REGISTRY_BASE_PATH env var is not set", func() {
			BeforeEach(func() {
				err := os.Unsetenv("REGISTRY_BASE_PATH")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := config.Load()
				Expect(err).To(MatchError("REGISTRY_BASE_PATH not configured"))
			})
		})


		Context("when the REGISTRY_PASSWORD env var is not set", func() {
			BeforeEach(func() {
				err := os.Unsetenv("REGISTRY_PASSWORD")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := config.Load()
				Expect(err).To(MatchError("REGISTRY_PASSWORD not configured"))
			})
		})

		Context("when the PORT env var is not set", func() {
			BeforeEach(func() {
				err := os.Unsetenv("PORT")
				Expect(err).NotTo(HaveOccurred())
			})

			It("defaults to 8000", func() {
				cfg, err := config.Load()
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg.Port).To(Equal(8080))
			})
		})

		Context("when the PORT env var is not a parsable integer", func() {
			BeforeEach(func() {
				err := os.Setenv("PORT", "🌝")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := config.Load()
				Expect(err).To(MatchError("PORT must be an integer"))
			})
		})
	})
})
