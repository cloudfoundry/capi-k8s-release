package main

import (
	"code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/handlers"
	"code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/upload"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	r := mux.NewRouter()
	r.HandleFunc("/packages", handlers.PostPackageHandler(upload.Upload, logger)).Methods("POST")
	fmt.Println("Starting server...")
	http.ListenAndServe(":8080", r)
	// TODO: should the port be configurable?
	// TODO: handle signals gracefully
}
