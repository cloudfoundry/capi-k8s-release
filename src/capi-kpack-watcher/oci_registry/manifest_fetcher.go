package oci_registry

import (
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

//go:generate mockery -case snake -name ManifestFetcher
type ManifestFetcher interface {
	FetchManifestFromImageReference(imageReference string) (*v1.Manifest, error)
}

type manifestFetcher struct {
}

func NewManifestFetcher() ManifestFetcher {
	return manifestFetcher{}
}

func (f manifestFetcher) FetchManifestFromImageReference(imageReference string) (*v1.Manifest, error) {
	ref, err := name.ParseReference(imageReference)
	if err != nil {
		panic(err)
	}

	image, err := remote.Image(ref)
	if err != nil {
		panic(err)
	}

	return image.Manifest()
}

