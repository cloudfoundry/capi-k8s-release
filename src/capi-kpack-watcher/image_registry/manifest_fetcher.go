package image_registry

import (
	"context"
	"errors"

	"github.com/containers/image/docker"
)

const (
	OCIManifestType    = "oci"
	DockerManifestType = "docker"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ManifestFetcher
type ManifestFetcher interface {
	FetchRawManifestFromImageReference(imageReference string) ([]byte, error)
}

type ociManifestFetcher struct {
}

type dockerManifestFetcher struct {
}

func NewManifestFetcher(manifestType string) ManifestFetcher {
	switch manifestType {
	case DockerManifestType:
		return dockerManifestFetcher{}
	case OCIManifestType:
		return ociManifestFetcher{}
	default:
		return ociManifestFetcher{}
	}
}
func (o ociManifestFetcher) FetchRawManifestFromImageReference(imageReference string) ([]byte, error) {
	return nil, errors.New("TODO: implement this")
}

func (d dockerManifestFetcher) FetchRawManifestFromImageReference(imageReference string) ([]byte, error) {
	// idk why you have to prefix "//" to the reference but that's what this library wants
	ref, err := docker.ParseReference("//" + imageReference)
	if err != nil {
		return nil, err
	}

	imgSrc, err := ref.NewImageSource(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	// TODO: do anything with MIME type return value?
	rawManifest, _, err := imgSrc.GetManifest(context.TODO(), nil)

	return rawManifest, err
}
