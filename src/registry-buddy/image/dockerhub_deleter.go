package image

import (
	"code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/dockerhub"
	"fmt"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"log"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fakes/dockerhub_client.go --fake-name DockerhubClient . DockerhubClient
type DockerhubClient interface {
	GetAuthorizationToken(username, password string) (string, error)
	DeleteRepo(repoName, token string) error
}

func NewDockerhubDeleter(client DockerhubClient) Deleter {
	if client == nil {
		client = dockerhub.NewClient(dockerhub.DefaultDomain)
	}

	return func(ref name.Reference, authenticator authn.Authenticator, logger *log.Logger) error {
		auth, err := authenticator.Authorization()
		if err != nil {
			return fmt.Errorf("failed fetching authorization information: %w", err)
		}

		token, err := client.GetAuthorizationToken(auth.Username, auth.Password)
		if err != nil {
			return fmt.Errorf("failed fetching authorization token from DockerHub: %w", err)
		}

		repositoryStr := ref.Context().RepositoryStr()
		err = client.DeleteRepo(repositoryStr, token)
		if err != nil {
			if _, ok := err.(*dockerhub.NotFoundError); ok {
				return nil
			}
			return fmt.Errorf("failed deleting repo %q from DockerHub: %w", repositoryStr, err)
		}

		return nil
	}

}
