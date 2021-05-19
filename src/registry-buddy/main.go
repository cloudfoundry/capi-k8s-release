package main

import (
	"code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/healthz"
	"code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/image"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/handlers"
	"code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/package_upload"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/gorilla/mux"

	"code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/config"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Unable to load config: %v\n", err)
	}

	done := make(chan bool, 1)
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	server := newServer(cfg, logger)
	go handleServerShutdown(server, done, shutdown, logger)

	fmt.Printf("Server is listening at %s...\n", server.Addr)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Server unable to listen: %v\n", err)
	}

	<-done
	logger.Println("Server stopped")
}

func newServer(cfg *config.Config, logger *log.Logger) *http.Server {
	authenticator := authn.FromConfig(authn.AuthConfig{
		Username: cfg.RegistryUsername,
		Password: cfg.RegistryPassword,
	})

	r := mux.NewRouter()
	r.HandleFunc("/packages", handlers.PostPackageHandler(package_upload.Upload, logger, authenticator)).Methods("POST")
	r.HandleFunc("/images", handlers.DeleteImageHandler(image.NewDynamicDeleter(), logger, authenticator)).Methods("DELETE")
	r.HandleFunc("/healthz", handlers.HealthzHandler(cfg.RegistryBasePath, healthz.Check, logger, authenticator)).Methods("GET")
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

func handleServerShutdown(server *http.Server, done chan<- bool, shutdown <-chan os.Signal, logger *log.Logger) {
	<-shutdown
	logger.Println("Server is attempting to shut down...")

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server unable to gracefully shutdown: %v\n", err)
	}
	close(done)
}
