package packagers

import (
	"fmt"
	"os"
	"runtime"

	"github.com/paketo-buildpacks/packit/v2/pexec"
)

// libpak-tools is a a number of helpful tools for the management and release of buildpacks.
// Rereference repo: https://github.com/paketo-buildpacks/libpak-tools
type LibpakTools struct {
	executable Executable
	pack       Executable
	tempOutput func(dir string, pattern string) (string, error)
}

func NewLibpakTools() LibpakTools {
	return LibpakTools{
		executable: pexec.NewExecutable("libpak-tools"),
		pack:       pexec.NewExecutable("pack"),
		tempOutput: os.MkdirTemp,
	}
}

func (l LibpakTools) WithExecutable(executable Executable) LibpakTools {
	l.executable = executable
	return l
}

func (l LibpakTools) WithPack(pack Executable) LibpakTools {
	l.pack = pack
	return l
}

func (l LibpakTools) WithTempOutput(tempOutput func(string, string) (string, error)) LibpakTools {
	l.tempOutput = tempOutput
	return l
}

func (l LibpakTools) Execute(buildpackDir, output, version string, cached bool) error {
	libpakToolsOutput, err := l.tempOutput("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(libpakToolsOutput)

	args := []string{
		"package", "compile",
		"--source", buildpackDir,
		"--destination", libpakToolsOutput,
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
		"--path", libpakToolsOutput,
		"--format", "file",
		"--target", fmt.Sprintf("linux/%s", runtime.GOARCH),
	}

	return l.pack.Execute(pexec.Execution{
		Args:   args,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}
