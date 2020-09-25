package auth

import (
	"context"

	uaaClient "github.com/cloudfoundry-community/go-uaa"
)

// UAAClient wraps over the official UAA client implementation.
type UAAClient struct {
	*uaaClient.API
}

// NewUAAClient creates a new UAA client.
func NewUAAClient(endpoint, clientName, clientSecret string) *UAAClient {
	client, err := uaaClient.New(
		endpoint,
		uaaClient.WithClientCredentials(clientName, clientSecret, 1),
		uaaClient.WithSkipSSLValidation(true),
	)
	if err != nil {
		panic(err)
	}

	return &UAAClient{client}
}

// Fetch implements the TokenFetcher interface, fetching tokens from UAA. This stands as an anti-corruption layer over
// the actual FetchToken call.
func (u *UAAClient) Fetch() (string, error) {
	token, err := u.Token(context.Background())
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}
