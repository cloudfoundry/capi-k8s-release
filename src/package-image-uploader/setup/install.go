package setup

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func SetupEnv(registry string) error {
	_, err := loadConfig(registry)
	if err != nil {
		return err
	}
	return nil
}

type config struct {
	username string
	password string
	registry string
}

func loadConfig(registry string) (config, error) {

	reg, err := name.ParseReference(registry+"/something", name.WeakValidation)
	if err != nil {
		return config{}, err
	}

	auth, err := authn.DefaultKeychain.Resolve(reg.Context().Registry)
	if err != nil {
		return config{}, err
	}

	basicAuth, err := auth.Authorization()
	if err != nil {
		return config{}, err
	}

	return config{
		username: basicAuth.Username,
		password: basicAuth.Password,
		registry: reg.Context().RegistryStr(),
	}, nil
}
