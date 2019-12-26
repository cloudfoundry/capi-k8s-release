package capi

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"

	"capi_kpack_watcher/model"
)

// PATCHBuild send a PATCH request to CAPI with a guid and status about a build.
func (c *client) PATCHBuild(guid string, status model.BuildStatus) error {
	const endpoint string = "v3/internal/builds"

	url := fmt.Sprintf("https://api.%s/%s/%s", c.host, endpoint, guid)

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(status.ToJSON()))
	if err != nil {
		return err
	}

	// TODO: Ideally this header should be set on the client, rather than for
	// every request.
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	log.Printf("[PATCHBuild] Sent payload: %s\n", status.ToJSON())
	log.Printf("[PATCHBuild] Repsonse Status: %d\n", resp.StatusCode)
	return nil
}

// NewCAPIClient creates a client to be used to communicate with CAPI. This
// client is a HTTP client for talking to CAPI over a REST API.
func NewCAPIClient() CAPI {
	// TODO: We may want to consider using cloudfoundry/tlsconfig for using
	// standard TLS configs in Golang.
	return &client{
		host: os.Getenv("CAPI_HOST"),
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}
}

type client struct {
	host       string
	httpClient *http.Client
}
