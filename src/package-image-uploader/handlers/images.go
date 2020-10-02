package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fakes/image_delete_func.go --fake-name ImageDeleteFunc . ImageDeleteFunc
type ImageDeleteFunc func(ref name.Reference, options ...remote.Option) error

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fakes/image_descriptor_func.go --fake-name ImageDescriptorFunc . ImageDescriptorFunc
type ImageDescriptorFunc func(ref name.Reference, options ...remote.Option) (*remote.Descriptor, error)

type deleteImageRequestBody struct {
	ImageReference string `json:"image_reference"`
}

type DeleteImageResponseBody struct {
	ImageReference string `json:"image_reference"`
}

func DeleteImageHandler(delete ImageDeleteFunc, get ImageDescriptorFunc, logger *log.Logger, authenticator authn.Authenticator) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		parsedBody := deleteImageRequestBody{}
		err := json.NewDecoder(request.Body).Decode(&parsedBody)
		if err != nil {
			logger.Printf("Failed to decode json body: %+v\n", err)
			writer.WriteHeader(400)
			writer.Write([]byte("unable to parse request body\n"))
			return
		}
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

		descriptor, err := get(ref, remote.WithAuth(authenticator))
		if err != nil {
			logger.Printf("Error from get(%s): %v\n", ref.Name(), err)
			writer.WriteHeader(500)
			writer.Write([]byte("unable to fetch image " + ref.Name() + "\n"))
			return
		}

		err = delete(ref, remote.WithAuth(authenticator))
		if err != nil {
			logger.Printf("Error from delete(%s): %v\n", ref.Name(), err)
			writer.WriteHeader(500)
			writer.Write([]byte("unable to delete image " + ref.Name() + "\n"))
			return
		}

		digestRefStr := fmt.Sprintf("%s@%s:%s", ref.Context().Name(), descriptor.Digest.Algorithm, descriptor.Digest.Hex)
		digestRef, err := name.ParseReference(digestRefStr)
		if err != nil {
			logger.Printf("Failed to parse reference '%s': %+v\n", digestRefStr, err)
			writer.WriteHeader(422)
			writer.Write([]byte("unable to parse image reference\n"))
			return
		}

		err = delete(digestRef, remote.WithAuth(authenticator))
		if err != nil {
			logger.Printf("Error from delete(%s): %v\n", digestRef.Name(), err)
			writer.WriteHeader(500)
			writer.Write([]byte("unable to delete image " + digestRef.Name() + "\n"))
			return
		}

		writer.WriteHeader(http.StatusAccepted)
		err = json.NewEncoder(writer).Encode(DeleteImageResponseBody{ImageReference: ref.Name()})
		if err != nil { // untested / untestable
			logger.Println("Error marshalling JSON response:", err)
			writer.WriteHeader(500)
			writer.Write([]byte("unable to encode JSON response"))
			return
		}

		logger.Printf("Finished deleting image: %s\n", ref.Name())
	}
}

func invalidDeleteImageRequest(parsedBody deleteImageRequestBody) bool {
	return parsedBody.ImageReference == ""
}
