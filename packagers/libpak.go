package packagers

import (
	"fmt"
	"os"
	"runtime"

	"github.com/paketo-buildpacks/packit/v2/pexec"
)

// create-package is a packager that builds libpak buildpacks' source code
// into tarballs. This type wraps that packaging tool. Libpak implements the
// freezer.Packager interface, and can therefore be passed as an argument to
// occam.BuildpackStore.WithPackager().
type Libpak struct {
	executable Executable
	pack       Executable
	tempOutput func(dir string, pattern string) (string, error)
}

func NewLibpak() Libpak {
	return Libpak{
		executable: pexec.NewExecutable("create-package"),
		pack:       pexec.NewExecutable("pack"),
		tempOutput: os.MkdirTemp,
	}
}

func (l Libpak) WithExecutable(executable Executable) Libpak {
	l.executable = executable
	return l
}

func (l Libpak) WithPack(pack Executable) Libpak {
	l.pack = pack
	return l
}

func (l Libpak) WithTempOutput(tempOutput func(string, string) (string, error)) Libpak {
	l.tempOutput = tempOutput
	return l
}

func (l Libpak) Execute(buildpackDir, output, version string, cached bool) error {
	libpakOutput, err := l.tempOutput("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(libpakOutput)

	args := []string{
		"--destination", libpakOutput,
		"--version", version,
	}

	if cached {
		args = append(args, "--include-dependencies")
	}

	err = l.executable.Execute(pexec.Execution{
		Args:   args,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Dir:    buildpackDir,
	})

	if err != nil {
		return err
	}

	args = []string{
		"buildpack", "package",
		output,
		"--path", libpakOutput,
		"--format", "file",
		"--target", fmt.Sprintf("linux/%s", runtime.GOARCH),
	}

	return l.pack.Execute(pexec.Execution{
		Args:   args,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}
