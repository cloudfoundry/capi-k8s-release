package auth

import (
	"io/ioutil"
	"os"
	"context"

	uaaClient "github.com/cloudfoundry-community/go-uaa"
)

// Fetch implements the TokenFetcher interface, fetching tokens from UAA. This stands as an anti-corruption layer over
// the actual FetchToken call.
func (u *UAAClient) Fetch() (string, error) {
	// allow uaa-go-client to handle token caching
	//cfg := GetSavedConfig()
	//token := cfg.GetActiveContext().Token
	token, err := u.Token(context.Background())
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
	//use kubebuilder logger for logging errors/uaa interactions
	client, err := uaaClient.New(os.Getenv("UAA_ENDPOINT"),
		uaaClient.WithClientCredentials(os.Getenv("UAA_CLIENT_NAME"), uaaClientSecret(), 1),
		uaaClient.WithSkipSSLValidation(true))
	//api, err := uaa.New(
	//	cfg.GetActiveTarget().BaseUrl,
	//	uaaClient.WithToken(&token),
	//	uaaClient.WithZoneID(cfg.ZoneSubdomain),
	//	uaaClient.WithSkipSSLValidation(true),
	//)
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
	*uaaClient.API
}
