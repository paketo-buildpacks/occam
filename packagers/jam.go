package packagers

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/paketo-buildpacks/packit/v2/pexec"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(pexec.Execution) error
}

// jam is a packager that builds packit buildpacks' source code into tarballs.
// This type wraps the jam executable, and implements the freezer.Packager
// interface, and can therefore be passed as an argument to
// occam.BuildpackStore.WithPackager().
type Jam struct {
	executable Executable
}

func NewJam() Jam {
	return Jam{
		executable: pexec.NewExecutable("jam"),
	}
}

func (j Jam) WithExecutable(executable Executable) Jam {
	j.executable = executable
	return j
}

func (j Jam) Execute(buildpackDir, output, version string, offline bool) error {

	extensionTomlPath := filepath.Join(buildpackDir, "extension.toml")

	buildpackOrExtensionToml := "buildpack.toml"
	command := "--buildpack"

	if fileExists, err := fs.Exists(extensionTomlPath); fileExists && err == nil {
		buildpackOrExtensionToml = "extension.toml"
		command = "--extension"
	}

	args := []string{
		"pack",
		command, filepath.Join(buildpackDir, buildpackOrExtensionToml),
		"--output", output,
		"--version", version,
	}

	if offline {
		args = append(args, "--offline")
	}

	return j.executable.Execute(pexec.Execution{
		Args:   args,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}
