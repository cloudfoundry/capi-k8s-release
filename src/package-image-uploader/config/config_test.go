package config_test

import (
	"os"

	"code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Describe("Load", func() {
		const (
			regUsername = "some-username"
			regPassword = "some-password"
		)

		BeforeEach(func() {
			err := os.Setenv("REGISTRY_USERNAME", regUsername)
			Expect(err).NotTo(HaveOccurred())
			err = os.Setenv("REGISTRY_PASSWORD", regPassword)
			Expect(err).NotTo(HaveOccurred())
			err = os.Setenv("PORT", "8081")
			Expect(err).NotTo(HaveOccurred())
		})

		It("loads the config", func() {
			cfg, err := config.Load()
			Expect(err).NotTo(HaveOccurred())

			Expect(cfg.RegistryUsername).To(Equal(regUsername))
			Expect(cfg.RegistryPassword).To(Equal(regPassword))
			Expect(cfg.Port).To(Equal(8081))
		})

		Context("when the REGISTRY_USERNAME env var is not set", func() {
			BeforeEach(func() {
				err := os.Unsetenv("REGISTRY_USERNAME")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := config.Load()
				Expect(err).To(MatchError("REGISTRY_USERNAME not configured"))
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

			It("returns an error", func() {
				_, err := config.Load()
				Expect(err).To(MatchError("PORT not configured"))
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
