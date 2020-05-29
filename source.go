package occam

import (
	"crypto/rand"
	"io/ioutil"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/fs"
)

func Source(path string) (string, error) {
	destination, err := ioutil.TempDir("", "source")
	if err != nil {
		return "", err
	}

	err = fs.Copy(path, destination)
	if err != nil {
		return "", err
	}

	content := make([]byte, 32)
	_, err = rand.Read(content)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(filepath.Join(destination, ".occam-key"), content, 0644)
	if err != nil {
		return "", err
	}

	return destination, nil
}
