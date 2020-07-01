package cf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-kpack-watcher/cf/api_model"
)

func NewClient(host string, restClient Rest, uaaClient TokenFetcher) *Client {
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

func (c *Client) UpdateBuild(guid string, build api_model.Build) error {
	token, err := c.uaaClient.Fetch()
	if err != nil {
		return err
	}

	raw, err := json.Marshal(build)
	if err != nil {
		return err
	}

	resp, err := c.restClient.Patch(
		fmt.Sprintf("%s/v3/builds/%s", c.host, guid),
		token,
		bytes.NewReader(raw),
	)
	if err != nil {
		return err
	}

	log.Printf("[CF API/UpdateBuild] Sent payload: %s\n", raw)
	log.Printf("[CF API/UpdateBuild] Response build: %d\n", resp.StatusCode)

	return nil
}
