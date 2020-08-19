package handlers

import (
	"code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/upload"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fakes/uploader_func.go --fake-name UploaderFunc . UploaderFunc
type UploaderFunc func(zipPath, registryPath string) (upload.Hash, error)

type postPackageBody struct {
	PackageZipPath   string `json:"package_zip_path"`
	PackageGuid      string `json:"package_guid"`
	RegistryBasePath string `json:"registry_base_path"`
}

type PostPackageResponse struct {
	Hash HashResponse `json:"hash"`
}

type HashResponse struct {
	Algorithm string `json:"algorithm"`
	Hex       string `json:"hex"`
}

func PostPackageHandler(uploadFunc UploaderFunc, logger *log.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		parsedBody := postPackageBody{}
		err := json.NewDecoder(request.Body).Decode(&parsedBody)
		if err != nil {
			logger.Println("Failed to decode json body:", err)
			writer.WriteHeader(400)
			writer.Write([]byte("unable to parse request body"))
			return
		}
		logger.Printf("Processing request: %+v\n", parsedBody)

		if invalidRequestBody(parsedBody) {
			logger.Printf("Invalid request body: %+v\n", parsedBody)
			writer.WriteHeader(422)
			writer.Write([]byte("missing required parameter"))
			return
		}

		fullRegistryPath := fmt.Sprintf("%s/%s", parsedBody.RegistryBasePath, parsedBody.PackageGuid)
		hash, err := uploadFunc(parsedBody.PackageZipPath, fullRegistryPath)
		if err != nil {
			logger.Printf("Error from uploadFunc(%s, %s): %v\n", parsedBody.PackageZipPath, fullRegistryPath, err)
			writer.WriteHeader(500)
			writer.Write([]byte("unable to convert/upload package " + parsedBody.PackageGuid))
			return
		}

		response := PostPackageResponse{Hash: HashResponse(hash)}
		err = json.NewEncoder(writer).Encode(response)
		if err != nil { // untested / untestable
			logger.Println("Error marshalling JSON response:", err)
			writer.WriteHeader(500)
			writer.Write([]byte("unable to encode JSON response"))
			return
		}

		logger.Printf("Finished processing request for package %q", parsedBody.PackageGuid)
	}
}

func invalidRequestBody(parsedBody postPackageBody) bool {
	return parsedBody.PackageZipPath == "" || parsedBody.PackageGuid == "" || parsedBody.RegistryBasePath == ""
}
