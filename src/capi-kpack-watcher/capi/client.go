package capi

import (
	"bytes"
	"capi_kpack_watcher/model"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"capi_kpack_watcher/auth"
)

// NewCAPIClient creates a client to be used to communicate with CAPI.
// The following environment variables must be set:
//   CAPI_HOST: Hostname of where CAPI is deployed (e.g. katniss.capi.land). If CAPI is deployed into Kubernetes, it
//              will be bound to a Kubernetes Service. You can then use that Service name here.
func NewCAPIClient() *Client {
	// TODO: We may want to consider using cloudfoundry/tlsconfig for using
	// standard TLS configs in Golang.
	return &Client{
		host: os.Getenv("CAPI_HOST"),
		restClient: &RestClient{
			&http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			},
		},
		uaaClient: auth.NewUAAClient(),
	}
}

//go:generate mockery -case snake -name Rest
type Rest interface {
	Patch(url string, authToken string, body io.Reader) (*http.Response, error)
}

//go:generate mockery -case snake -name TokenFetcher
type TokenFetcher interface {
	Fetch() (string, error)
}

type Client struct {
	host       string
	restClient Rest
	uaaClient  TokenFetcher
}

func (c *Client) UpdateBuild(guid string, status model.BuildStatus) error {
	token, err := c.uaaClient.Fetch()
	if err != nil {
		return err
	}

	json := status.ToJSON()

	resp, err := c.restClient.Patch(
		fmt.Sprintf("https://api.%s/v3/internal/build/%s", c.host, guid),
		token,
		bytes.NewReader(json),
	)
	if err != nil {
		return err
	}

	log.Printf("[CAPI/UpdateBuild] Sent payload: %s\n", json)
	log.Printf("[CAPI/UpdateBuild] Response status: %d\n", resp.StatusCode)

	return nil
}
