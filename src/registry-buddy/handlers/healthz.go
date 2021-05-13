package handlers

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"log"
	"net/http"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fakes/healthy_func.go --fake-name HealthyFunc . HealthyFunc
type HealthyFunc func(registryPath string, authenticator authn.Authenticator) error

func HealthzHandler(registryBasePath string, healthyFunc HealthyFunc, logger *log.Logger, authenticator authn.Authenticator) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		err := healthyFunc(registryBasePath, authenticator)
		if err != nil {
			logger.Printf("Error from healthyFunc(%s): %v\n", registryBasePath, err)
			writer.WriteHeader(401)
			writer.Write([]byte("unable to reach registry " + registryBasePath))
			return
		}
	}
}
