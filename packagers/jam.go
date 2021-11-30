package packagers

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/pexec"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(pexec.Execution) error
}

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
	args := []string{
		"pack",
		"--buildpack", filepath.Join(buildpackDir, "buildpack.toml"),
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
