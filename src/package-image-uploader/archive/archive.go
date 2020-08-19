package archive

import (
	"archive/tar"
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/pkg/ioutils"
	"github.com/pkg/errors"
)

var NormalizedDateTime time.Time

func init() {
	NormalizedDateTime = time.Date(1980, time.January, 1, 0, 0, 1, 0, time.UTC)
}

type TarWriter interface {
	WriteHeader(hdr *tar.Header) error
	Write(b []byte) (int, error)
	Close() error
}

type TarWriterFactory interface {
	NewWriter(io.Writer) TarWriter
}

type defaultTarWriterFactory struct{}

func DefaultTarWriterFactory() TarWriterFactory {
	return defaultTarWriterFactory{}
}

func (defaultTarWriterFactory) NewWriter(w io.Writer) TarWriter {
	return tar.NewWriter(w)
}

func ReadZipAsTar(srcPath, basePath string, uid, gid int, mode int64, normalizeModTime bool, fileFilter func(string) bool) io.ReadCloser {
	return GenerateTar(func(tw TarWriter) error {
		return WriteZipToTar(tw, srcPath, basePath, uid, gid, mode, normalizeModTime, fileFilter)
	})
}

func GenerateTar(genFn func(TarWriter) error) io.ReadCloser {
	return GenerateTarWithWriter(genFn, DefaultTarWriterFactory())
}

// GenerateTarWithTar returns a reader to a tar from a generator function using a writer from the provided factory.
// Note that the generator will not fully execute until the reader is fully read from. Any errors returned by the
// generator will be returned when reading the reader.
func GenerateTarWithWriter(genFn func(TarWriter) error, twf TarWriterFactory) io.ReadCloser {
	errChan := make(chan error)
	pr, pw := io.Pipe()

	go func() {
		tw := twf.NewWriter(pw)
		defer func() {
			if r := recover(); r != nil {
				tw.Close()
				pw.CloseWithError(errors.Errorf("panic: %v", r))
			}
		}()

		err := genFn(tw)

		closeErr := tw.Close()
		closeErr = aggregateError(closeErr, pw.CloseWithError(err))

		errChan <- closeErr
	}()

	closed := false
	return ioutils.NewReadCloserWrapper(pr, func() error {
		if closed {
			return errors.New("reader already closed")
		}

		var completeErr error

		// closing the reader ensures that if anything attempts
		// further reading it doesn't block waiting for content
		if err := pr.Close(); err != nil {
			completeErr = aggregateError(completeErr, err)
		}

		// wait until everything closes properly
		if err := <-errChan; err != nil {
			completeErr = aggregateError(completeErr, err)
		}

		closed = true
		return completeErr
	})
}

func aggregateError(base, addition error) error {
	if addition == nil {
		return base
	}

	if base == nil {
		return addition
	}

	return errors.Wrap(addition, base.Error())
}

// ErrEntryNotExist is an error returned if an entry path doesn't exist
var ErrEntryNotExist = errors.New("not exist")

// WriteZipToTar writes the contents of a zip file to a tar writer.
func WriteZipToTar(tw TarWriter, srcZip, basePath string, uid, gid int, mode int64, normalizeModTime bool, fileFilter func(string) bool) error {
	zipReader, err := zip.OpenReader(srcZip)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	var fileMode int64
	for _, f := range zipReader.File {
		if fileFilter != nil && !fileFilter(f.Name) {
			continue
		}

		fileMode = mode
		if isFatFile(f.FileHeader) {
			fileMode = 0777
		}

		var header *tar.Header
		if f.Mode()&os.ModeSymlink != 0 {
			target, err := func() (string, error) {
				r, err := f.Open()
				if err != nil {
					return "", nil
				}
				defer r.Close()

				// contents is the target of the symlink
				target, err := ioutil.ReadAll(r)
				if err != nil {
					return "", err
				}

				return string(target), nil
			}()

			if err != nil {
				return err
			}

			header, err = tar.FileInfoHeader(f.FileInfo(), target)
			if err != nil {
				return err
			}
		} else {
			header, err = tar.FileInfoHeader(f.FileInfo(), f.Name)
			if err != nil {
				return err
			}
		}

		header.Name = filepath.ToSlash(filepath.Join(basePath, f.Name))
		finalizeHeader(header, uid, gid, fileMode, normalizeModTime)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if f.Mode().IsRegular() {
			err := func() error {
				fi, err := f.Open()
				if err != nil {
					return err
				}
				defer fi.Close()

				_, err = io.Copy(tw, fi)
				return err
			}()

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isFatFile(header zip.FileHeader) bool {
	var (
		creatorFAT  uint16 = 0
		creatorVFAT uint16 = 14
	)

	// This identifies FAT files, based on the `zip` source: https://golang.org/src/archive/zip/struct.go
	firstByte := header.CreatorVersion >> 8
	return firstByte == creatorFAT || firstByte == creatorVFAT
}

func finalizeHeader(header *tar.Header, uid, gid int, mode int64, normalizeModTime bool) {
	NormalizeHeader(header, normalizeModTime)
	if mode != -1 {
		header.Mode = mode
	}
	header.Uid = uid
	header.Gid = gid
}

// NormalizeHeader normalizes a tar.Header
//
// Normalizes the following:
// 	- ModTime
// 	- GID
// 	- UID
// 	- User Name
// 	- Group Name
func NormalizeHeader(header *tar.Header, normalizeModTime bool) {
	if normalizeModTime {
		header.ModTime = NormalizedDateTime
	}
	header.Uid = 0
	header.Gid = 0
	header.Uname = ""
	header.Gname = ""
}
