package image

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"log"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fakes/raw_image_delete_func.go --fake-name RawImageDeleteFunc . RawImageDeleteFunc
type RawImageDeleteFunc func(ref name.Reference, options ...remote.Option) error

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fakes/image_descriptor_func.go --fake-name ImageDescriptorFunc . ImageDescriptorFunc
type ImageDescriptorFunc func(ref name.Reference, options ...remote.Option) (*remote.Descriptor, error)

func NewGenericDeleter(delete RawImageDeleteFunc, get ImageDescriptorFunc) Deleter {
	return func(ref name.Reference, auth authn.Authenticator, logger *log.Logger) error {
		descriptor, err := get(ref, remote.WithAuth(auth))
		if err != nil {
			if e, ok := err.(*transport.Error); ok && e.StatusCode == 404 {
				logger.Println("get image descriptor returned 404; assuming image was previously deleted")
				return nil
			}

			logger.Printf("Error from get (%s): %v\n", ref.Name(), err)
			return fmt.Errorf("unable to fetch image metadata: %w", err)
		}

		digestRefStr := fmt.Sprintf("%s@%s:%s", ref.Context().Name(), descriptor.Digest.Algorithm, descriptor.Digest.Hex)
		digestRef, err := name.ParseReference(digestRefStr)
		if err != nil {
			logger.Printf("Failed to parse reference '%s': %+v\n", digestRefStr, err)
			return err // untested
		}

		// if image reference is a tag, first attempt to delete the tag (required by GCR)
		if _, ok := ref.(name.Tag); ok {
			err = attemptDeleteByTag(ref, auth, logger, delete)
			if err != nil {
				logger.Printf("Error from delete (%s): %v\n", ref.Name(), err)
				return fmt.Errorf("error deleting image by tag: %w", err)
			}
		}

		err = delete(digestRef, remote.WithAuth(auth))
		if err != nil {
			if e, ok := err.(*transport.Error); ok && e.StatusCode == 404 {
				logger.Println("delete returned 404; assuming image was previously deleted")
			} else {
				logger.Printf("Error from delete (%s): %v\n", digestRef.Name(), err)
				return fmt.Errorf("error deleting image: %w", err)
			}
		}

		return nil
	}
}

// many registries do not support deleting manifests by tag, though some require it
// attemptDeleteByTag will attempt the deletion, but ignore UNSUPPORTED errors and 404s
func attemptDeleteByTag(ref name.Reference, authenticator authn.Authenticator, logger *log.Logger, delete RawImageDeleteFunc) error {
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

func errorUnsupported(errors []transport.Diagnostic) bool {
	for _, e := range errors {
		if e.Code == transport.UnsupportedErrorCode {
			return true
		}
	}

	return false
}
