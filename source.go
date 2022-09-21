package occam

import (
	"crypto/rand"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2/fs"
)

// Source will copy `path` into a temporary directory and return the path to the temporary directory.
// It will also place a file with random contents into the temporary directory, to ensure that the
// contents are globally unique, which is meant to bypass reuse of cached layers.
//
// The caller must clean up the returned directory.
func Source(path string) (string, error) {
	destination, err := os.MkdirTemp("", "source")
	if err != nil {
		return "", err
	}

	return destination, move(path, destination)
}

type SimpleTesting interface {
	Fatalf(format string, args ...interface{})
	TempDir() string
}

// SourceTesting will copy `path` into a temporary directory and return the path to the temporary directory.
// It will also place a file with random contents into the temporary directory, to ensure that the
// contents are globally unique, which is meant to bypass reuse of cached layers.
//
// SimpleTesting is implemented by *testing.T, so this function can easily be consumed by unit tests.
// It will fail the tests instead of returning an error.
// The temporary directory will be automatically cleaned up, see "testing.T".TempDir
func SourceTesting(path string, t SimpleTesting) string {
	destination := t.TempDir()

	if err := move(path, destination); err != nil {
		t.Fatalf(err.Error())
	}

	return destination
}

func move(source, dest string) error {
	err := fs.Copy(source, dest)
	if err != nil {
		return err
	}

	content := make([]byte, 32)
	_, err = rand.Read(content)
	if err != nil { // untested
		return err
	}

	// untested
	return os.WriteFile(filepath.Join(dest, ".occam-key"), content, 0644)
}
