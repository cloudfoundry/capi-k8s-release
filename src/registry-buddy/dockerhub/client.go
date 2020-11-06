package dockerhub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type client struct {
	Domain string
}

type NotFoundError struct {
	Err error
}

func (e NotFoundError) Error() string {
	return e.Err.Error()
}

const DefaultDomain = "https://hub.docker.com"

func NewClient(domain string) *client {
	if domain == "" {
		domain = DefaultDomain
	}
	return &client{
		Domain: domain,
	}
}

type authorizationRequestBody struct {
	Username string `json:"username"`
	Password string	`json:"password"`
}

type authorizationResponseBody struct {
	Token string `json:"token"`
}

func (c *client) GetAuthorizationToken(username, password string) (string, error) {
	httpClient := c.buildHttpClient()

	body, err := json.Marshal(authorizationRequestBody{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", fmt.Errorf("error marshalling authorization request: %w", err) // untested
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v2/users/login/", c.Domain), bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("error building authorization request: %w", err) // untested
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error during authorization request: %w", err) // untested
	}
	defer res.Body.Close()

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err) // untested
	}

	if res.StatusCode == http.StatusUnauthorized {
		return "", fmt.Errorf("unauthorized request (status %d): %s", res.StatusCode, string(bodyBytes))
	} else if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response (status %d): %s", res.StatusCode, string(bodyBytes))
	}

	resBody := authorizationResponseBody{}
	err = json.Unmarshal(bodyBytes, &resBody)
	if err != nil {
		return "", fmt.Errorf("error decoding authorization response: error %w, body %q", err, string(bodyBytes)) // untested
	}

	return resBody.Token, nil
}

func (c *client) DeleteRepo(repoName, token string) error {
	httpClient := c.buildHttpClient()

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v2/repositories/%s/", c.Domain, repoName), bytes.NewBuffer(nil))
	if err != nil {
		return fmt.Errorf("error building delete repository request: %w", err) // untested
	}
	req.Header.Add("AUTHORIZATION", fmt.Sprintf("JWT %s", token))

	res, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error during delete repository request: %w", err) // untested
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusAccepted {
		return nil
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err) // untested
	}

	switch res.StatusCode {
	case http.StatusNotFound:
		return &NotFoundError{Err: fmt.Errorf("not found (status %d): %s", res.StatusCode, string(bodyBytes))}
	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized request (status %d): %s", res.StatusCode, string(bodyBytes))
	default:
		return fmt.Errorf("unexpected response (status %d): %s", res.StatusCode, string(bodyBytes))
	}
}

func (c *client) buildHttpClient() *http.Client {
	return new(http.Client)
}

