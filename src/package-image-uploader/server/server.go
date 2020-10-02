package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/config"
	"code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/handlers"
	"code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/package_upload"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/gorilla/mux"
)

func NewServer(cfg *config.Config, logger *log.Logger) *http.Server {
	authenticator := authn.FromConfig(authn.AuthConfig{
		Username: cfg.RegistryUsername,
		Password: cfg.RegistryPassword,
	})

	r := mux.NewRouter()
	r.HandleFunc("/packages", handlers.PostPackageHandler(package_upload.Upload, logger, authenticator)).Methods("POST")
	r.HandleFunc("/registry", handlers.DeleteImageHandler(remote.Delete, logger, authenticator)).Methods("DELETE")
	addr := fmt.Sprintf("127.0.0.1:%d", cfg.Port)

	return &http.Server{
		Addr:              addr,
		Handler:           r,
		ErrorLog:          logger,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       15 * time.Second,
	}
}
