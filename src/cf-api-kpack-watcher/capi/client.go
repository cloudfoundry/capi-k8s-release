package capi

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-kpack-watcher/auth"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-kpack-watcher/capi_model"
)

// TODO: stop using this constructor (to much implicitness/side effects)
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

func NewCFAPIClient(host string, restClient Rest, uaaClient TokenFetcher) *Client {
	// TODO: We may want to consider using cloudfoundry/tlsconfig for using
	// standard TLS configs in Golang.
	return &Client{
		host:       host,
		restClient: restClient,
		uaaClient:  uaaClient,
	}
}

// TODO: remove mockery usages after refactoring everything to use Ginkgo for consistency
//go:generate mockery -case snake -name Rest
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Rest
type Rest interface {
	Patch(url string, authToken string, body io.Reader) (*http.Response, error)
}

// TODO: remove mockery usages after refactoring everything to use Ginkgo for consistency
//go:generate mockery -case snake -name TokenFetcher
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . TokenFetcher
type TokenFetcher interface {
	Fetch() (string, error)
}

type Client struct {
	host       string
	restClient Rest
	uaaClient  TokenFetcher
}

func (c *Client) UpdateBuild(guid string, build capi_model.Build) error {
	token, err := c.uaaClient.Fetch()
	if err != nil {
		return err
	}

	json := build.ToJSON()

	resp, err := c.restClient.Patch(
		fmt.Sprintf("%s/v3/builds/%s", c.host, guid),
		token,
		bytes.NewReader(json),
	)
	if err != nil {
		return err
	}

	log.Printf("[CAPI/UpdateBuild] Sent payload: %s\n", json)
	log.Printf("[CAPI/UpdateBuild] Response build: %d\n", resp.StatusCode)

	return nil
}
