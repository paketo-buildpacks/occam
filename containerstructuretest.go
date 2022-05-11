package occam

import (
	"bytes"
	"fmt"

	"github.com/paketo-buildpacks/packit/v2/pexec"
)

type ContainerStructureTest struct {
	executable Executable
	verbose    bool
	noColor    bool
	pull       bool
}

func NewContainerStructureTest() ContainerStructureTest {
	return ContainerStructureTest{
		executable: pexec.NewExecutable("container-structure-test"),
	}
}

func (c ContainerStructureTest) WithExecutable(executable Executable) ContainerStructureTest {
	c.executable = executable
	return c
}

func (c ContainerStructureTest) WithVerbose() ContainerStructureTest {
	c.verbose = true
	return c
}

func (c ContainerStructureTest) WithNoColor() ContainerStructureTest {
	c.noColor = true
	return c
}

func (c ContainerStructureTest) WithPull() ContainerStructureTest {
	c.pull = true
	return c
}

func (r ContainerStructureTest) Execute(imageID string, config string) (string, error) {
	args := []string{"test"}

	if r.verbose {
		args = append(args, "--verbosity", "debug")
	}

	if r.noColor {
		args = append(args, "--no-color")
	}

	if r.pull {
		args = append(args, "--pull")
	}

	args = append(args, "--config", config, "--image", imageID)

	log := bytes.NewBuffer(nil)
	err := r.executable.Execute(pexec.Execution{
		Args:   args,
		Stdout: log,
		Stderr: log,
	})

	if err != nil {
		return log.String(), fmt.Errorf("failed to run container-structure-test: %w\n\nOutput:\n%s", err, log)
	}

	return log.String(), nil
}
