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
	if err != nil { // untested
		return "", err
	}

	err = fs.Copy(path, destination)
	if err != nil {
		return "", err
	}

	content := make([]byte, 32)
	_, err = rand.Read(content)
	if err != nil { // untested
		return "", err
	}

	err = os.WriteFile(filepath.Join(destination, ".occam-key"), content, 0644)
	if err != nil { // untested
		return "", err
	}

	return destination, nil
}
