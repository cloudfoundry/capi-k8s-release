package upload

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"os"

	"code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/archive"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

type Hash v1.Hash

func Upload(zipPath, registryPath string, authenticator authn.Authenticator) (Hash, error) {
	image, err := random.Image(0, 0)
	if err != nil {
		return Hash{}, err
	}

	noopFilter := func(string) bool { return true }
	layer, err := tarball.LayerFromReader(archive.ReadZipAsTar(zipPath, "/", 0, 0, -1, true, noopFilter))
	if err != nil {
		return Hash{}, err
	}

	image, err = mutate.AppendLayers(image, layer)
	if err != nil {
		return Hash{}, err
	}

	ref, err := name.ParseReference(registryPath)
	if err != nil {
		return Hash{}, err
	}

	options := []remote.Option{remote.WithAuth(authenticator)}
	optAppRegistryCAPath := os.Getenv("APP_REGISTRY_CA")
	if optAppRegistryCAPath != "" {
		bs, err := ioutil.ReadFile(optAppRegistryCAPath)
		if err != nil {
			panic(err)
		}
		certPool := x509.NewCertPool()
		ok := certPool.AppendCertsFromPEM(bs)
		if !ok {
			panic("TODO: lol idk")
		}
		transport := http.Transport{TLSClientConfig: &tls.Config{RootCAs: certPool}}
		options = append(options, remote.WithTransport(&transport))
	}
	err = remote.Write(ref, image, options...)
	if err != nil {
		return Hash{}, err
	}

	hash, err := image.Digest()
	return Hash(hash), err
}
