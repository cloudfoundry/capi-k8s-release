package dockerhub_test

import (
	. "code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/dockerhub"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Client", func() {
	type Client interface {
		GetAuthorizationToken(username, password string) (string, error)
		DeleteRepo(repoName, token string) error
	}

	var (
		testServer *ghttp.Server
		client     Client
	)

	const (
		username = "my-username"
		password = "my-password"
	)

	BeforeEach(func() {
		testServer = ghttp.NewServer()
		client = NewClient(testServer.URL())
	})

	AfterEach(func() {
		testServer.Close()
	})

	Describe("GetAuthorizationToken", func() {
		var (
			responseJSON   map[string]string
			responseStatus int
		)

		BeforeEach(func() {
			responseStatus = 200
			testServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v2/users/login/"),
					ghttp.VerifyJSONRepresenting(map[string]string{
						"username": username,
						"password": password,
					}),
					ghttp.VerifyHeaderKV("Content-Type", "application/json"),
					ghttp.RespondWithJSONEncodedPtr(&responseStatus, &responseJSON),
				),
			)
		})

		It("sends a POST request to the dockerhub API", func() {
			_, _ = client.GetAuthorizationToken(username, password)
			Expect(testServer.ReceivedRequests()).To(HaveLen(1))
		})

		When("The response is 200", func() {
			const expectedToken = "my-token"

			BeforeEach(func() {
				responseStatus = 200
				responseJSON = map[string]string{
					"token": expectedToken,
				}
			})

			It("returns the token", func() {
				actualToken, err := client.GetAuthorizationToken(username, password)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualToken).To(Equal(expectedToken))
			})
		})

		When("The response is 401", func() {
			BeforeEach(func() {
				responseStatus = 401
				responseJSON = map[string]string{"error": "boom"}
			})

			It("returns an unauthorized error", func() {
				_, err := client.GetAuthorizationToken(username, password)
				Expect(err).To(MatchError(ContainSubstring("unauthorized")))
				Expect(err).To(MatchError(ContainSubstring("401")))
				Expect(err).To(MatchError(ContainSubstring("boom")))
			})
		})

		When("The response is another non-success value", func() {
			BeforeEach(func() {
				responseStatus = 500
				responseJSON = map[string]string{"failure": ":facepalm:"}
			})

			It("returns an error", func() {
				_, err := client.GetAuthorizationToken(username, password)
				Expect(err).To(MatchError(ContainSubstring("500")))
				Expect(err).To(MatchError(ContainSubstring(":facepalm:")))
			})
		})
	})

	Describe("DeleteRepo", func() {
		var (
			responseBody   string
			responseStatus int
		)

		const (
			repoName = "my-repo/package-guid"
			token    = "valid-token"
		)

		BeforeEach(func() {
			responseStatus = 202
			testServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE",
						fmt.Sprintf("/v2/repositories/%s/", repoName),
					),
					ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("JWT %s", token)),
					ghttp.RespondWithPtr(&responseStatus, &responseBody),
				),
			)
		})

		It("sends a DELETE request to the dockerhub API", func() {
			_ = client.DeleteRepo(repoName, token)
			Expect(testServer.ReceivedRequests()).To(HaveLen(1))
		})

		When("The response is 202", func() {
			BeforeEach(func() {
				responseStatus = 202
			})

			It("returns no error", func() {
				Expect(client.DeleteRepo(repoName, token)).To(Succeed())
			})
		})

		When("The response is 404", func() {
			BeforeEach(func() {
				responseStatus = 404
				responseBody = "no such record"
			})

			It("returns a NotFound error", func() {
				err := client.DeleteRepo(repoName, token)
				Expect(err).To(MatchError(ContainSubstring("404")))
				Expect(err).To(MatchError(ContainSubstring("no such record")))
				Expect(err).To(BeAssignableToTypeOf(new(NotFoundError)))
			})
		})

		When("The response is 401", func() {
			BeforeEach(func() {
				responseStatus = 401
				responseBody = ""
			})

			It("returns an unauthorized error", func() {
				err := client.DeleteRepo(repoName, token)
				Expect(err).To(MatchError(ContainSubstring("401")))
				Expect(err).To(MatchError(ContainSubstring("unauthorized")))
			})
		})

		When("The response is another non-success value", func() {
			BeforeEach(func() {
				responseStatus = 500
				responseBody = ":facepalm:"
			})

			It("returns an error", func() {
				err := client.DeleteRepo(repoName, token)
				Expect(err).To(MatchError(ContainSubstring("500")))
				Expect(err).To(MatchError(ContainSubstring(":facepalm:")))
			})
		})
	})
})
