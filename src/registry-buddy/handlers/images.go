package handlers

import (
	"code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/image"
	"encoding/json"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"log"
	"net/http"
)

type deleteImageRequestBody struct {
	ImageReference string `json:"image_reference"`
}

type DeleteImageResponseBody struct {
	ImageReference string `json:"image_reference"`
}

func DeleteImageHandler(deleter image.Deleter, logger *log.Logger, authenticator authn.Authenticator) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		parsedBody := deleteImageRequestBody{}
		err := json.NewDecoder(request.Body).Decode(&parsedBody)
		if err != nil {
			logger.Printf("Failed to decode json body: %+v\n", err)
			writer.WriteHeader(400)
			writer.Write([]byte("unable to parse request body\n"))
			return
		}
		defer request.Body.Close()
		logger.Printf("Processing request: %+v\n", parsedBody)

		if invalidDeleteImageRequest(parsedBody) {
			logger.Printf("Invalid request body: %+v\n", parsedBody)
			writer.WriteHeader(422)
			writer.Write([]byte("missing required parameter\n"))
			return
		}

		ref, err := name.ParseReference(parsedBody.ImageReference)
		if err != nil {
			logger.Printf("Failed to parse reference '%s': %+v\n", parsedBody.ImageReference, err)
			writer.WriteHeader(422)
			writer.Write([]byte("unable to parse image reference\n"))
			return
		}

		err = deleter(ref, authenticator, logger)
		if err != nil {
			logger.Printf("Error from delete (%s): %v\n", ref.Name(), err)

			writer.WriteHeader(500)
			writer.Write([]byte("unable to delete image " + ref.Name() + "\n"))
			return
		}

		err = writeSuccessResponse(writer, ref)
		if err != nil { // untested / untestable
			logger.Println("Error marshalling JSON response:", err)
			writer.WriteHeader(500)
			writer.Write([]byte("unable to encode JSON response"))
			return
		}

		logger.Printf("Finished deleting image: %s\n", ref.Name())
	}
}

func writeSuccessResponse(w http.ResponseWriter, ref name.Reference) error {
	w.WriteHeader(http.StatusAccepted)
	return json.NewEncoder(w).Encode(DeleteImageResponseBody{ImageReference: ref.Name()})
}

func invalidDeleteImageRequest(parsedBody deleteImageRequestBody) bool {
	return parsedBody.ImageReference == ""
}
