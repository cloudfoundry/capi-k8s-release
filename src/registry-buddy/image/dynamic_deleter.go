package image

import (
	"code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/dockerhub"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"log"
	"strings"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fakes/deleter.go --fake-name Deleter . Deleter
type Deleter func(name.Reference, authn.Authenticator, *log.Logger) error

func NewDynamicDeleter() Deleter {
	return func(reference name.Reference, authenticator authn.Authenticator, logger *log.Logger) error {
		var deleter Deleter
		if strings.HasPrefix(reference.Context().Name(), "index.docker.io/") {
			deleter = NewDockerhubDeleter(dockerhub.NewClient(dockerhub.DefaultDomain))
		} else {
			deleter = NewGenericDeleter(remote.Delete, remote.Get)
		}
		return deleter(reference, authenticator, logger)
	}
}
