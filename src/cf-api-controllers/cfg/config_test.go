package cfg_test

import (
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cfg"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
)

var _ = Describe("Config", func() {
	Describe("Load", func() {
		const (
			expectedCFAPIHost               = "api.cloudfoundry.example.com"
			expectedUAAEndpoint             = "uaa.cloudfoundry.example.com"
			expectedUAAClientName           = "example-uaa-client-name"
			expectedWorkloadsNamespace      = "example-cf-workloads"
			expectedUAAClientSecretFromEnv  = "uaa-client-secret-from-env"
			expectedUAAClientSecretFromFile = "uaa-client-secret-from-file"
		)

		BeforeEach(func() {
			err := os.Setenv("CF_API_HOST", expectedCFAPIHost)
			Expect(err).NotTo(HaveOccurred())
			err = os.Setenv("UAA_ENDPOINT", expectedUAAEndpoint)
			Expect(err).NotTo(HaveOccurred())
			err = os.Setenv("UAA_CLIENT_NAME", expectedUAAClientName)
			Expect(err).NotTo(HaveOccurred())
			err = os.Setenv("UAA_CLIENT_SECRET", expectedUAAClientSecretFromEnv)
			Expect(err).NotTo(HaveOccurred())
			err = os.Setenv("WORKLOADS_NAMESPACE", expectedWorkloadsNamespace)
			Expect(err).NotTo(HaveOccurred())
		})

		It("loads the config from env", func() {
			config, err := cfg.Load()
			Expect(err).NotTo(HaveOccurred())

			Expect(config.CFAPIHost()).To(Equal(expectedCFAPIHost))
			Expect(config.UAAEndpoint()).To(Equal(expectedUAAEndpoint))
			Expect(config.UAAClientName()).To(Equal(expectedUAAClientName))
			Expect(config.UAAClientSecret()).To(Equal(expectedUAAClientSecretFromEnv))
			Expect(config.WorkloadsNamespace()).To(Equal(expectedWorkloadsNamespace))
		})

		Context("when the CF_API_HOST env var is not set", func() {
			BeforeEach(func() {
				err := os.Unsetenv("CF_API_HOST")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := cfg.Load()
				Expect(err).To(MatchError("`CF_API_HOST` environment variable must be set"))
			})
		})

		Context("when the UAA_ENDPOINT env var is not set", func() {
			BeforeEach(func() {
				err := os.Unsetenv("UAA_ENDPOINT")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := cfg.Load()
				Expect(err).To(MatchError("`UAA_ENDPOINT` environment variable must be set"))
			})
		})

		Context("when the UAA_CLIENT_NAME env var is not set", func() {
			BeforeEach(func() {
				err := os.Unsetenv("UAA_CLIENT_NAME")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := cfg.Load()
				Expect(err).To(MatchError("`UAA_CLIENT_NAME` environment variable must be set"))
			})
		})

		Context("when the WORKLOADS_NAMESPACE env var is not set", func() {
			BeforeEach(func() {
				err := os.Unsetenv("WORKLOADS_NAMESPACE")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := cfg.Load()
				Expect(err).To(MatchError("`WORKLOADS_NAMESPACE` environment variable must be set"))
			})
		})

		Describe("loading UAA Client Secret", func() {
			BeforeEach(func() {
				err := os.Unsetenv("UAA_CLIENT_SECRET_FILE")
				Expect(err).NotTo(HaveOccurred())
				err = os.Unsetenv("UAA_CLIENT_SECRET")
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when UAA_CLIENT_SECRET_FILE is set", func() {
				Context("and the file can be read without error", func() {
					var tempSecretFile *os.File

					BeforeEach(func() {
						var err error
						tempSecretFile, err = ioutil.TempFile("", "fake-uaa-client-secret")
						Expect(err).NotTo(HaveOccurred())

						_, err = tempSecretFile.WriteString(expectedUAAClientSecretFromFile)
						Expect(err).NotTo(HaveOccurred())

						err = os.Setenv("UAA_CLIENT_SECRET_FILE", tempSecretFile.Name())
						Expect(err).NotTo(HaveOccurred())
					})

					AfterEach(func() {
						err := os.Remove(tempSecretFile.Name())
						Expect(err).NotTo(HaveOccurred())
					})

					It("loads the client secret from the specified file", func() {
						config, err := cfg.Load()
						Expect(err).NotTo(HaveOccurred())

						Expect(config.UAAClientSecret()).To(Equal(expectedUAAClientSecretFromFile))
					})
				})

				Context("and an error occurs while reading the file", func() {
					BeforeEach(func() {
						err := os.Setenv("UAA_CLIENT_SECRET_FILE", "hello-im-a-fake-file.zip")
						Expect(err).NotTo(HaveOccurred())
					})

					It("returns an error", func() {
						_, err := cfg.Load()
						Expect(err).To(HaveOccurred())

						_, ok := err.(*os.PathError)
						Expect(ok).To(BeTrue())
					})
				})
			})

			Context("when UAA_CLIENT_SECRET_FILE is not set", func() {
				Context("and UAA_CLIENT_SECRET is set", func() {
					BeforeEach(func() {
						err := os.Setenv("UAA_CLIENT_SECRET", expectedUAAClientSecretFromEnv)
						Expect(err).NotTo(HaveOccurred())
					})

					It("loads the client secret from the environment", func() {
						config, err := cfg.Load()
						Expect(err).NotTo(HaveOccurred())

						Expect(config.UAAClientSecret()).To(Equal(expectedUAAClientSecretFromEnv))
					})
				})

				Context("UAA_CLIENT_SECRET is not set", func() {
					It("returns an error", func() {
						_, err := cfg.Load()
						errMsg := "`UAA_CLIENT_SECRET_FILE` or `UAA_CLIENT_SECRET` environment variable must be set"
						Expect(err).To(MatchError(errMsg))
					})
				})
			})
		})
	})
})
