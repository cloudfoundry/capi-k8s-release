package auth

import (
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	uaaClient "code.cloudfoundry.org/uaa-go-client"
	uaaConfig "code.cloudfoundry.org/uaa-go-client/config"
)

// Fetch implements the TokenFetcher interface, fetching tokens from UAA. This stands as an anti-corruption layer over
// the actual FetchToken call.
func (u *UAAClient) Fetch() (string, error) {
	// allow uaa-go-client to handle token caching
	const forceUpdate = false
	token, err := u.FetchToken(forceUpdate)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

// NewUAAClient creates a new UAA client.
// The following environment variables must be set:
//   UAA_CLIENT_NAME: Name of client inside UAA (e.g. in CF, this is configured on the UAA job in a BOSH manifest).
//	 UAA_CLIENT_SECRET: Secret generated for the client in UAA, similar to above.
//	 UAA_ENDPOINT: The FQDN of UAA (e.g. https://uaa.katniss.capi.land)
func NewUAAClient() *UAAClient {
	logger := lager.NewLogger("uaa")
	clock := clock.NewClock()

	client, err := uaaClient.NewClient(
		logger,
		&uaaConfig.Config{
			ClientName:       os.Getenv("UAA_CLIENT_NAME"),
			ClientSecret:     uaaClientSecret(),
			UaaEndpoint:      os.Getenv("UAA_ENDPOINT"),
			SkipVerification: true, // TODO: actually verify in the future
		},
		clock,
	)
	if err != nil {
		panic(err)
	}

	return &UAAClient{client}
}

func uaaClientSecret() string {
	if os.Getenv("UAA_CLIENT_SECRET_FILE") != "" {
		contents, err := ioutil.ReadFile(os.Getenv("UAA_CLIENT_SECRET_FILE"))
		if err != nil {
			panic(err)
		}
		return string(contents)
	}

	return os.Getenv("UAA_CLIENT_SECRET")
}

// UAAClient wraps over the official UAA client implementation.
type UAAClient struct {
	uaaClient.Client
}
