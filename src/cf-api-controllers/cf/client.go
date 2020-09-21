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

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Rest
type Rest interface {
	Patch(url string, authToken string, body io.Reader) (*http.Response, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . TokenFetcher
type TokenFetcher interface {
	Fetch() (string, error)
}

// TODO: rename this to `Client`
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ClientInterface
type ClientInterface interface {
	ListRoutes() (model.RouteList, error)
}

type Client struct {
	host       string
	restClient Rest
	uaaClient  TokenFetcher
	httpClient *http.Client
}

// determined by CC API: https://v3-apidocs.cloudfoundry.org/version/3.76.0/index.html#get-a-route
const MaxResultsPerPage int = 5000

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

// TODO: shouldn't this use the REST client?
func (c *Client) ListRoutes() (model.RouteList, error) {
	token, err := c.uaaClient.Fetch()
	if err != nil {
		return model.RouteList{}, err
	}

	pathAndQuery := fmt.Sprintf("%s/v3/routes?per_page=%d&include=space,domain", c.host, MaxResultsPerPage)
	request, err := http.NewRequest("GET", pathAndQuery, nil)
	if err != nil {
		return model.RouteList{}, err
	}
	request.Header.Set("Authorization", "bearer "+token)

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return model.RouteList{}, fmt.Errorf("failed to list routes, HTTP error: %w", err)
	}
	if resp.StatusCode != 200 {
		return model.RouteList{}, fmt.Errorf("failed to list routes, received status: %d", resp.StatusCode)
	}

	var routeList model.RouteList
	err = json.NewDecoder(resp.Body).Decode(&routeList)
	if err != nil {
		return model.RouteList{}, fmt.Errorf("failed to deserialize response from CF API: %w", err)
	}

	return routeList, nil
}
