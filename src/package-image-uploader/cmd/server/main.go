package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/config"

	"code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/handlers"
	"code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/upload"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/gorilla/mux"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal(err)
	}

	authenticator := authn.FromConfig(authn.AuthConfig{
		Username: cfg.RegistryUsername,
		Password: cfg.RegistryPassword,
	})

	r := mux.NewRouter()
	r.HandleFunc("/packages", handlers.PostPackageHandler(upload.Upload, logger, authenticator)).Methods("POST")
	fmt.Println("Starting server...")
	http.ListenAndServe(fmt.Sprintf("localhost:%d", cfg.Port), r)
	// TODO: handle signals gracefully
}
