// NOTICE: This file was modified from its original source:
// https://github.com/buildpacks/pack/blob/10f5f0a357ccc0b8cc4cabf7ee9a70814386f0cd/internal/archive/archive_test.go

package archive_test

import (
	"archive/tar"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	"code.cloudfoundry.org/capi-k8s-release/src/registry-buddy/archive"
	h "github.com/buildpacks/pack/testhelpers"
)

func TestArchive(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	spec.Run(t, "Archive", testArchive, spec.Sequential(), spec.Report(report.Terminal{}))
}

func testArchive(t *testing.T, when spec.G, it spec.S) {
	var (
		tmpDir string
	)

	it.Before(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "create-tar-test")
		if err != nil {
			t.Fatalf("failed to create tmp dir %s: %s", tmpDir, err)
		}
	})

	it.After(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatalf("failed to clean up tmp dir %s: %s", tmpDir, err)
		}
	})

	when("#WriteZipToTar", func() {
		var src string
		it.Before(func() {
			src = filepath.Join("testdata", "zip-to-tar.zip")
		})

		when("mode is set to 0777", func() {
			it("writes a tar to the dest dir with 0777", func() {
				fh, err := os.Create(filepath.Join(tmpDir, "some.tar"))
				h.AssertNil(t, err)

				tw := tar.NewWriter(fh)

				err = archive.WriteZipToTar(tw, src, "/nested/dir/dir-in-archive", 1234, 2345, 0777, true, nil)
				h.AssertNil(t, err)
				h.AssertNil(t, tw.Close())
				h.AssertNil(t, fh.Close())

				file, err := os.Open(filepath.Join(tmpDir, "some.tar"))
				h.AssertNil(t, err)
				defer file.Close()

				tr := tar.NewReader(file)

				verify := h.NewTarVerifier(t, tr, 1234, 2345)
				verify.NextFile("/nested/dir/dir-in-archive/some-file.txt", "some-content", 0777)
				verify.NextDirectory("/nested/dir/dir-in-archive/sub-dir", 0777)
				if runtime.GOOS != "windows" {
					verify.NextSymLink("/nested/dir/dir-in-archive/sub-dir/link-file", "../some-file.txt")
				}
			})
		})

		when("mode is set to -1", func() {
			it("writes a tar to the dest dir with preexisting file mode", func() {
				fh, err := os.Create(filepath.Join(tmpDir, "some.tar"))
				h.AssertNil(t, err)

				tw := tar.NewWriter(fh)

				err = archive.WriteZipToTar(tw, src, "/nested/dir/dir-in-archive", 1234, 2345, -1, true, nil)
				h.AssertNil(t, err)
				h.AssertNil(t, tw.Close())
				h.AssertNil(t, fh.Close())

				file, err := os.Open(filepath.Join(tmpDir, "some.tar"))
				h.AssertNil(t, err)
				defer file.Close()

				tr := tar.NewReader(file)

				verify := h.NewTarVerifier(t, tr, 1234, 2345)
				verify.NextFile("/nested/dir/dir-in-archive/some-file.txt", "some-content", 0644)
				verify.NextDirectory("/nested/dir/dir-in-archive/sub-dir", 0755)
				if runtime.GOOS != "windows" {
					verify.NextSymLink("/nested/dir/dir-in-archive/sub-dir/link-file", "../some-file.txt")
				}
			})

			when("files are compressed in fat (MSDOS) format", func() {
				it.Before(func() {
					src = filepath.Join("testdata", "fat-zip-to-tar.zip")
				})

				it("writes a tar to the dest dir with 0777", func() {
					fh, err := os.Create(filepath.Join(tmpDir, "some.tar"))
					h.AssertNil(t, err)

					tw := tar.NewWriter(fh)

					err = archive.WriteZipToTar(tw, src, "/nested/dir/dir-in-archive", 1234, 2345, -1, true, nil)
					h.AssertNil(t, err)
					h.AssertNil(t, tw.Close())
					h.AssertNil(t, fh.Close())

					file, err := os.Open(filepath.Join(tmpDir, "some.tar"))
					h.AssertNil(t, err)
					defer file.Close()

					tr := tar.NewReader(file)

					verify := h.NewTarVerifier(t, tr, 1234, 2345)
					verify.NextFile("/nested/dir/dir-in-archive/some-file.txt", "some-content", 0777)
					verify.NoMoreFilesExist()
				})
			})
		})

		when("normalize mod time is false", func() {
			it("does not normalize mod times", func() {
				tarFile := filepath.Join(tmpDir, "some.tar")
				fh, err := os.Create(tarFile)
				h.AssertNil(t, err)

				tw := tar.NewWriter(fh)

				err = archive.WriteZipToTar(tw, src, "/foo", 1234, 2345, 0777, false, nil)
				h.AssertNil(t, err)
				h.AssertNil(t, tw.Close())
				h.AssertNil(t, fh.Close())

				h.AssertOnTarEntry(t, tarFile, "/foo/some-file.txt",
					h.DoesNotHaveModTime(archive.NormalizedDateTime),
				)
			})
		})

		when("normalize mod time is true", func() {
			it("normalizes mod times", func() {
				tarFile := filepath.Join(tmpDir, "some.tar")
				fh, err := os.Create(tarFile)
				h.AssertNil(t, err)

				tw := tar.NewWriter(fh)

				err = archive.WriteZipToTar(tw, src, "/foo", 1234, 2345, 0777, true, nil)
				h.AssertNil(t, err)
				h.AssertNil(t, tw.Close())
				h.AssertNil(t, fh.Close())

				h.AssertOnTarEntry(t, tarFile, "/foo/some-file.txt",
					h.HasModTime(archive.NormalizedDateTime),
				)
			})
		})
	})
}
