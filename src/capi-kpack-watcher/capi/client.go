package capi

import (
	"bytes"
	"capi_kpack_watcher/capi_model"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
    "encoding/json"
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
	Post(url string, authToken string, body io.Reader) (*http.Response, error)
	Get(url string, authToken string, body io.Reader) (*http.Response, error)
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

func (c *Client) GetCurrentDroplet(guid string) (string, error) {
	token, err := c.uaaClient.Fetch()
	if err != nil {
		return "", err
	}

	resp, err := c.restClient.Get(
		fmt.Sprintf("%s/v3/apps/%s/droplets/current", c.host, guid),
		token,
		nil,
	)
	if err != nil {
		return "", err
	}
	log.Printf(`[CAPI/GetAppCurrentDroplet] Response StatusCode: %d`, resp.StatusCode)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(bodyBytes, &result)
	dropletGUID := result["guid"].(string)
	log.Printf(`[CAPI/GetAppCurrentDroplet] Droplet Guid: %s`, dropletGUID)
	return dropletGUID, nil
}

func (c *Client) CreateDropletCopy(dropletGUID, appGUID, image string) error {
	token, err := c.uaaClient.Fetch()
	if err != nil {
		return err
	}
	reqBody := fmt.Sprintf(`{
    "relationships": {
      "app": {
        "data": {
          "guid": "%s"
        }
      }
    }}`, appGUID)
	raw := json.RawMessage(reqBody)
	reqBodyBytes, err := raw.MarshalJSON()
	if err!=nil{
		return err
	}
	params := url.Values{}
	params.Add("source_guid", dropletGUID)
	params.Add("image_ref", image)

	resp, err := c.restClient.Post(
		fmt.Sprintf(`%s/v3/droplets?%s`, c.host, params.Encode()),
		token,
		bytes.NewReader(reqBodyBytes),
	)
	if err != nil {
		return err
	}
	log.Printf(`[CAPI/CreateDropletCopy] Response StatusCode: %d`, resp.StatusCode)
	return nil
}
