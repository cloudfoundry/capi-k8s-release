package healthz

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"net/http"
)

func Check(registryPath string, authenticator authn.Authenticator) error {
	repo, err := name.NewRepository(registryPath)
	if err != nil {
		panic(err)
	}

	scopes := []string{repo.Scope(transport.PushScope)} // TODO: double-check this one

	// This pings the registry to determine how to authenticate (basic or bearer)
	// if this fails, either the server is down or the credentials provided to get the auth token is invalid
	t, err := transport.New(repo.Registry, authenticator, http.DefaultTransport, scopes)
	if err != nil {
		return fmt.Errorf("error setting up transport to the registry: %s", err)
	}
	client := &http.Client{Transport: t}

	resp, err := client.Get(fmt.Sprintf("https://%s/v2/", repo.Registry.Name()))
	if err != nil {
		return fmt.Errorf("unable to reach the registry: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	} else {
		return fmt.Errorf("unable to reach the registry, status code: %s", resp.StatusCode)
	}
}
