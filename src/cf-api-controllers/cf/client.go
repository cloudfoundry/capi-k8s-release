package cf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/model"
)

func NewClient(host string, restClient Rest, uaaClient TokenFetcher) *Client {
	// TODO: We may want to consider using cloudfoundry/tlsconfig for using
	// standard TLS configs in Golang.
	return &Client{
		host:       host,
		restClient: restClient,
		uaaClient:  uaaClient,
		httpClient: &http.Client{},
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

// TODO: rename this to `Client`
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ClientInterface
type ClientInterface interface {
	ListRoutes() ([]model.Route, error)
	GetSpace(spaceGUID string) (*model.Space, error)
	GetDomain(domainGUID string) (*model.Domain, error)
}

// TODO: replace this with the client the cf-cli uses?
type Client struct {
	host       string
	restClient Rest
	uaaClient  TokenFetcher
	httpClient *http.Client
}

func (c *Client) UpdateBuild(buildGUID string, build model.Build) error {
	token, err := c.uaaClient.Fetch()
	if err != nil {
		return err
	}

	raw, err := json.Marshal(build)
	if err != nil {
		return err
	}

	resp, err := c.restClient.Patch(
		fmt.Sprintf("%s/v3/builds/%s", c.host, buildGUID),
		token,
		bytes.NewReader(raw),
	)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to patch build, received status %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) UpdateDroplet(dropletGUID string, droplet model.Droplet) error {
	token, err := c.uaaClient.Fetch()
	if err != nil {
		return err
	}

	raw, err := json.Marshal(droplet)
	if err != nil {
		return err
	}

	resp, err := c.restClient.Patch(
		fmt.Sprintf("%s/v3/droplets/%s", c.host, dropletGUID),
		token,
		bytes.NewReader(raw),
	)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to patch droplet, received status %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) ListRoutes() ([]model.Route, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/routes", c.host), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to list routes, HTTP error: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to list routes, received status: %d", resp.StatusCode)
	}

	var routes []model.Route
	err = json.NewDecoder(resp.Body).Decode(&routes)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize response from CF API: %w", err)
	}

	return routes, nil
}

func (c *Client) GetSpace(spaceGUID string) (*model.Space, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/spaces/%s", c.host, spaceGUID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get space, HTTP error: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get space, received status: %d", resp.StatusCode)
	}

	space := model.Space{}
	err = json.NewDecoder(resp.Body).Decode(&space)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize response from CF API: %w", err)
	}

	return &space, nil
}

func (c *Client) GetDomain(domainGUID string) (*model.Domain, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/domains/%s", c.host, domainGUID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain, HTTP error: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get domain, received status: %d", resp.StatusCode)
	}

	domain := model.Domain{}
	err = json.NewDecoder(resp.Body).Decode(&domain)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize response from CF API: %w", err)
	}

	return &domain, nil
}
