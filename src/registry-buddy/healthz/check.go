package healthz

import (
	"errors"
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

	scopes := []string{repo.Scope(transport.CatalogScope)} // is this the right scope?
	t, err := transport.New(repo.Registry, authenticator, http.DefaultTransport, scopes)
	if err != nil {
		panic(err)
	}
	client := &http.Client{Transport: t}

	resp, err := client.Get(fmt.Sprintf("https://%s/v2/", repo.Registry.Name()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	} else {
		return errors.New("something bad happened... fill me in")
	}
}
