package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
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

		// fetch image descriptor to discover the digest manifest for deletion
		descriptor, err := get(ref, remote.WithAuth(authenticator))
		if err != nil {
			if e, ok := err.(*transport.Error); ok && e.StatusCode == 404 {
				logger.Println("get image descriptor returned 404; assuming image was previously deleted")
				err = writeSuccessResponse(writer, ref)
				if err != nil { // untested / untestable
					logger.Println("Error marshalling JSON response:", err)
					writer.WriteHeader(500)
					writer.Write([]byte("unable to encode JSON response"))
				}
				return
			}

			logger.Printf("Error from get(%s): %v\n", ref.Name(), err)
			writer.WriteHeader(500)
			writer.Write([]byte("unable to fetch image metadata " + ref.Name() + "\n"))
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

		// if image reference is a tag, first attempt to delete the tag (required by GCR)
		if _, ok := ref.(name.Tag); ok {
			err = attemptDeleteByTag(ref, authenticator, logger, delete)
			if err != nil {
				logger.Printf("Error from delete(%s): %v\n", ref.Name(), err)

				writer.WriteHeader(500)
				writer.Write([]byte("unable to delete image " + ref.Name() + "\n"))
				return
			}
		}

		err = delete(digestRef, remote.WithAuth(authenticator))
		if err != nil {
			if e, ok := err.(*transport.Error); ok && e.StatusCode == 404 {
				logger.Println("delete returned 404; assuming image was previously deleted")
			} else {
				logger.Printf("Error from delete(%s): %v\n", digestRef.Name(), err)
				writer.WriteHeader(500)
				writer.Write([]byte("unable to delete image " + digestRef.Name() + "\n"))
				return
			}
		}

		err = writeSuccessResponse(writer, digestRef)
		if err != nil { // untested / untestable
			logger.Println("Error marshalling JSON response:", err)
			writer.WriteHeader(500)
			writer.Write([]byte("unable to encode JSON response"))
			return
		}

		logger.Printf("Finished deleting image: %s\n", ref.Name())
	}
}

// many registries do not support deleting manifests by tag, though some require it
// attemptDeleteByTag will attempt the deletion, but ignore UNSUPPORTED errors and 404s
func attemptDeleteByTag(ref name.Reference, authenticator authn.Authenticator, logger *log.Logger, delete ImageDeleteFunc) error {
	err := delete(ref, remote.WithAuth(authenticator))
	if err != nil {
		if e, ok := err.(*transport.Error); ok {
			if e.StatusCode == 404 {
				logger.Println("delete returned 404; assuming tag was previously deleted")
				return nil
			} else if errorUnsupported(e.Errors) {
				logger.Printf("delete returned %d; operation is unsupported for tags", e.StatusCode)
				return nil
			}
		}
	}

	return err
}

func writeSuccessResponse(w http.ResponseWriter, ref name.Reference) error {
	w.WriteHeader(http.StatusAccepted)
	return json.NewEncoder(w).Encode(DeleteImageResponseBody{ImageReference: ref.Name()})
}

func invalidDeleteImageRequest(parsedBody deleteImageRequestBody) bool {
	return parsedBody.ImageReference == ""
}

func errorUnsupported(errors []transport.Diagnostic) bool {
	for _, e := range errors {
		if e.Code == transport.UnsupportedErrorCode {
			return true
		}
	}

	return false
}
