package capi

import (
	"bytes"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRestClient_PATCH(t *testing.T) {
	spec.Run(t, "TestRestClient_PATCH", func(t *testing.T, when spec.G, it spec.S) {
		var (
			restClient     RestClient
			testServer *httptest.Server
			authToken string
			body io.Reader
		)

		it.Before(func() {
			status := []byte(`{"status":"SUCCESS"}`)

			body = bytes.NewReader(status)
			testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.URL.String(), "/")

				b, err := ioutil.ReadAll(r.Body)
				if err != nil {
					panic(err)
				}
				assert.Equal(t, b, status)

				assert.Equal(t, r.Header.Get("Authorization"), authToken)
				assert.Equal(t, r.Header.Get("Content-Type"), "application/json")

				w.WriteHeader(http.StatusOK)
			}))
            authToken = "valid-auth-token-returned-by-uaa"
			restClient = RestClient{
				client: testServer.Client(),
			}
		})

		when("request is valid", func() {
			it.After(func() {
				testServer.Close()
			})

			it("responds with StatusOK returned from CAPI", func() {
				response, err := restClient.Patch(testServer.URL, authToken, body)
				assert.Equal(t, http.StatusOK, response.StatusCode)
				assert.NoError(t, err)
			})
		})
	})
}
