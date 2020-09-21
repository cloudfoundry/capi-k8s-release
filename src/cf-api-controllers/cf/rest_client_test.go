package cf

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RestClient", func() {
	Describe("PATCH", func() {
		var (
			restClient RestClient
			testServer *httptest.Server
			authToken  string
			body       io.Reader
		)

		BeforeEach(func() {
			status := []byte(`{"status":"SUCCESS"}`)
			authToken = "valid-auth-token-returned-by-uaa"

			body = bytes.NewReader(status)
			testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.URL.String()).To(Equal("/"))

				b, err := ioutil.ReadAll(r.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(b).To(Equal(status))

				Expect(r.Header.Get("Authorization")).To(Equal(fmt.Sprintf("Bearer %s", authToken)))
				Expect(r.Header.Get("Content-Type")).To(Equal("application/json"))

				w.WriteHeader(http.StatusOK)
			}))

			restClient = RestClient{
				testServer.Client(),
			}
		})

		AfterEach(func() {
			testServer.Close()
		})

		When("request is valid", func() {
			It("receives a 200 OK response from CF API", func() {
				response, err := restClient.Patch(testServer.URL, authToken, body)
				Expect(err).NotTo(HaveOccurred())
				Expect(http.StatusOK).To(Equal(response.StatusCode))
			})
		})

	})
})
