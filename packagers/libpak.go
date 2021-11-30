package packagers

import (
	"os"

	"github.com/paketo-buildpacks/packit/pexec"
)

type Libpak struct {
	executable Executable
}

func NewLibpak() Libpak {
	return Libpak{
		executable: pexec.NewExecutable("create-package"),
	}
}

func (l Libpak) Execute(buildpackDir, output, version string, cached bool) error {
	args := []string{
		"--destination", output,
		"--version", version,
	}

	if cached {
		args = append(args, "--include-dependencies")
	}

	return l.executable.Execute(pexec.Execution{
		Args:   args,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Dir:    buildpackDir,
	})
}

func (l Libpak) WithExecutable(executable Executable) Libpak {
	l.executable = executable
	return l
}
