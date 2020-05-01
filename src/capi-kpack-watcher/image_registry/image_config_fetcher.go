package image_registry

import (
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ImageConfigFetcher
type ImageConfigFetcher interface {
	FetchImageConfig(imageReference string) (*v1.Config, error)
}

type ociImageConfigFetcher struct {
}

func NewImageConfigFetcher() ImageConfigFetcher {
	return ociImageConfigFetcher{}
}

func (f ociImageConfigFetcher) FetchImageConfig(imageReference string) (*v1.Config, error) {
	ref, err := name.ParseReference(imageReference)
	if err != nil {
		return nil, err
	}

	img, err := remote.Image(ref)
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
