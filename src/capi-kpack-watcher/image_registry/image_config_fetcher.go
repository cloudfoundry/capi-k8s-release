package image_registry

import (
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pivotal/kpack/pkg/registry"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ImageConfigFetcher
type ImageConfigFetcher interface {
	FetchImageConfig(imageReference, buildServiceAccount, buildNamespace string) (*v1.Config, error)
}

type OciImageConfigFetcher struct {
	KeychainFactory    registry.KeychainFactory
	ImageConfigFetcher ImageConfigFetcher
}

// TODO: supply private registry credentials, defaulting to empty strings
func NewOciImageConfigFetcher(keychainFactory registry.KeychainFactory) OciImageConfigFetcher {
	return OciImageConfigFetcher{KeychainFactory: keychainFactory}
}

func (f OciImageConfigFetcher) FetchImageConfig(imageReference, buildServiceAccount, buildNamespace string) (*v1.Config, error) {
	ref, err := name.ParseReference(imageReference)
	if err != nil {
		return nil, err
	}

	// TODO: supply registry with credentials if present on fetcher
	keychain, err := f.KeychainFactory.KeychainForSecretRef(registry.SecretRef{
		ServiceAccount: buildServiceAccount,
		Namespace:      buildNamespace,
	})
	if err != nil {
		return nil, err
	}

	img, err := remote.Image(ref, remote.WithAuthFromKeychain(keychain))
	if err != nil {
		return nil, err
	}

	cfgFile, err := img.ConfigFile()
	if err != nil {
		return nil, err
	}

	// TODO: address potential nil-pointer deref here
	return &cfgFile.Config, nil
}
