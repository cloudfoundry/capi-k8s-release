package cfmetadata_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"code.cloudfoundry.org/capi-k8s-release/src/backup-metadata/internal/cfmetadata"
)

var _ = Describe("CF Client", func() {
	var (
		err      error
		fakeCF   *ghttp.Server
		cfClient *cfmetadata.Client
	)

	BeforeEach(func() {
		fakeCF = ghttp.NewServer()
		fakeCF.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/v2/info"),
				ghttp.RespondWith(http.StatusOK, `{"token_endpoint": "`+fakeCF.URL()+`"}`),
			),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/oauth/token"),
				ghttp.RespondWith(http.StatusOK, `{"access_token": "foo"}`, http.Header{
					"Content-Type": []string{"application/json"},
				}),
			),
		)

		cfClient, err = cfmetadata.NewClient(fakeCF.URL(), "test-cf-user", "test-cf-password")
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		fakeCF.Close()
	})

	Describe("#Users", func() {
		Context("given users exist", func() {
			It("returns expected number of users", func() {
				fakeCF.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/users"),
					ghttp.RespondWith(http.StatusOK, cfmetadata.UsersResponse)),
				)

				users, err := cfClient.Users()
				Expect(users).Should(HaveLen(2))
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("given users do not exist", func() {
			It("returns empty list", func() {
				fakeCF.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/users"),
					ghttp.RespondWith(http.StatusOK, `{"resources": []}`)),
				)

				users, err := cfClient.Users()
				Expect(users).Should(BeEmpty())
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("#Spaces", func() {
		Context("given spaces exist", func() {
			It("returns expected number of spaces", func() {
				fakeCF.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/spaces"),
					ghttp.RespondWith(http.StatusOK, cfmetadata.SpacesResponse)),
				)

				spaces, err := cfClient.Spaces()
				Expect(spaces).Should(HaveLen(2))
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("given spaces do not exist", func() {
			It("returns empty list", func() {
				fakeCF.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/spaces"),
					ghttp.RespondWith(http.StatusOK, `{"resources": []}`)),
				)

				spaces, err := cfClient.Spaces()
				Expect(spaces).Should(BeEmpty())
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("#Apps", func() {
		Context("given apps exist", func() {
			It("returns expected number of apps", func() {
				fakeCF.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/apps"),
					ghttp.RespondWith(http.StatusOK, cfmetadata.AppsResponse)),
				)

				apps, err := cfClient.Apps()
				Expect(apps).Should(HaveLen(1))
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("given apps do not exist", func() {
			It("returns empty list", func() {
				fakeCF.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/apps"),
					ghttp.RespondWith(http.StatusOK, `{"resources": []}`)),
				)

				apps, err := cfClient.Apps()
				Expect(apps).Should(BeEmpty())
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("#Orgs", func() {
		Context("given orgs exist", func() {
			It("returns expected number of orgs", func() {
				fakeCF.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/organizations"),
					ghttp.RespondWith(http.StatusOK, cfmetadata.OrgsResponse)),
				)

				orgs, err := cfClient.Orgs()
				Expect(orgs).Should(HaveLen(3))
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("given apps do not exist", func() {
			It("returns empty list", func() {
				fakeCF.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/organizations"),
					ghttp.RespondWith(http.StatusOK, `{"resources": []}`)),
				)

				orgs, err := cfClient.Orgs()
				Expect(orgs).Should(BeEmpty())
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})
})
