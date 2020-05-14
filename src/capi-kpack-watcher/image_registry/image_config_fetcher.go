package image_registry

import (
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pivotal/kpack/pkg/registry"
	"github.com/pivotal/kpack/pkg/dockercreds/k8sdockercreds"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ImageConfigFetcher
type ImageConfigFetcher interface {
	FetchImageConfig(imageReference, buildServiceAccount, buildNamespace string) (*v1.Config, error)
}

type ociImageConfigFetcher struct {
}

// TODO: supply private registry credentials, defaulting to empty strings
func NewImageConfigFetcher() ImageConfigFetcher {
	return ociImageConfigFetcher{}
}

func (f ociImageConfigFetcher) FetchImageConfig(imageReference, buildServiceAccount, buildNamespace string) (*v1.Config, error) {
	ref, err := name.ParseReference(imageReference)
	if err != nil {
		return nil, err
	}

	// TODO: supply registry with credentials if present on fetcher
    keychainFactory, err := k8sdockercreds.NewSecretKeychainFactory(k8sClient)
	keychain, err := keychainFactory.KeychainForSecretRef(registry.SecretRef{
		ServiceAccount: "foo",
		Namespace:      "bar",
	})
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
