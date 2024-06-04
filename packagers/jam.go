package packagers

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

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
	pack       Executable
	tempOutput func(dir string, pattern string) (string, error)
}

func NewJam() Jam {
	return Jam{
		executable: pexec.NewExecutable("jam"),
		pack:       pexec.NewExecutable("pack"),
		tempOutput: os.MkdirTemp,
	}
}

func (j Jam) WithExecutable(executable Executable) Jam {
	j.executable = executable
	return j
}

func (j Jam) WithPack(pack Executable) Jam {
	j.pack = pack
	return j
}

func (j Jam) WithTempOutput(tempOutput func(string, string) (string, error)) Jam {
	j.tempOutput = tempOutput
	return j
}

func (j Jam) Execute(buildpackDir, output, version string, offline bool) error {
	jamOutput, err := j.tempOutput("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(jamOutput)

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
		"--output", filepath.Join(jamOutput, fmt.Sprintf("%s.tgz", version)),
		"--version", version,
	}

	if offline {
		args = append(args, "--offline")
	}

	err = j.executable.Execute(pexec.Execution{
		Args:   args,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		return err
	}

	args = []string{
		"buildpack", "package",
		output,
		"--path", filepath.Join(jamOutput, fmt.Sprintf("%s.tgz", version)),
		"--format", "file",
		"--target", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	return j.pack.Execute(pexec.Execution{
		Args:   args,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}
