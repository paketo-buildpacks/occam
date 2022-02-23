package packagers

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2/pexec"
)

// create-package is a packager that builds libpak buildpacks' source code
// into tarballs. This type wraps that packaging tool. Libpak implements the
// freezer.Packager interface, and can therefore be passed as an argument to
// occam.BuildpackStore.WithPackager().
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
