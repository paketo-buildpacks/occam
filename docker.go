package occam

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/packit/pexec"
)

type Docker struct {
	Image struct {
		Inspect DockerImageInspect
		Remove  DockerImageRemove
	}

	Container struct {
		Inspect DockerContainerInspect
		Run     DockerContainerRun
		Remove  DockerContainerRemove
	}
}

func NewDocker() Docker {
	var docker Docker
	executable := pexec.NewExecutable("docker", lager.NewLogger("docker"))

	docker.Image.Inspect = DockerImageInspect{executable: executable}
	docker.Image.Remove = DockerImageRemove{executable: executable}
	docker.Container.Run = DockerContainerRun{executable: executable, env: map[string]string{"PORT": "8080"}}
	docker.Container.Remove = DockerContainerRemove{executable: executable}
	docker.Container.Inspect = DockerContainerInspect{executable: executable}

	return docker
}

func (d Docker) WithExecutable(executable Executable) Docker {
	d.Image.Inspect.executable = executable
	d.Image.Remove.executable = executable

	d.Container.Inspect.executable = executable
	d.Container.Remove.executable = executable
	d.Container.Run.executable = executable
	d.Container.Run.inspect = d.Container.Inspect

	return d
}

type DockerImageInspect struct {
	executable Executable
}

func (i DockerImageInspect) Execute(ref string) (Image, error) {
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	_, _, err := i.executable.Execute(pexec.Execution{
		Args:   []string{"image", "inspect", ref},
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		return Image{}, fmt.Errorf("failed to inspect docker image: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return NewImageFromInspectOutput(stdout.Bytes())
}

type DockerImageRemove struct {
	executable Executable
}

func (r DockerImageRemove) Execute(ref string) error {
	stderr := bytes.NewBuffer(nil)
	_, _, err := r.executable.Execute(pexec.Execution{
		Args:   []string{"image", "remove", ref},
		Stderr: stderr,
	})
	if err != nil {
		return fmt.Errorf("failed to remove docker image: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return nil
}

type DockerContainerRun struct {
	executable Executable
	inspect    DockerContainerInspect

	command string
	env     map[string]string
	memory  string
}

func (r DockerContainerRun) WithEnv(env map[string]string) DockerContainerRun {
	r.env = env
	return r
}

func (r DockerContainerRun) WithMemory(memoryLimit string) DockerContainerRun {
	r.memory = memoryLimit
	return r
}

func (r DockerContainerRun) WithCommand(command string) DockerContainerRun {
	r.command = command
	return r
}

func (r DockerContainerRun) Execute(imageID string) (Container, error) {
	args := []string{"container", "run", "--detach"}

	if len(r.env) > 0 {
		var keys []string
		for key := range r.env {
			keys = append(keys, key)
		}

		sort.Strings(keys)

		for _, key := range keys {
			args = append(args, "--env", fmt.Sprintf("%s=%s", key, r.env[key]))
		}
	}

	args = append(args, "--publish", r.env["PORT"], "--publish-all")

	if r.memory != "" {
		args = append(args, "--memory", r.memory)
	}

	args = append(args, imageID)

	if r.command != "" {
		args = append(args, r.command)
	}

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	_, _, err := r.executable.Execute(pexec.Execution{
		Args:   args,
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		return Container{}, fmt.Errorf("failed to run docker container: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return r.inspect.Execute(strings.TrimSpace(stdout.String()))
}

type DockerContainerRemove struct {
	executable Executable
}

func (r DockerContainerRemove) Execute(containerID string) error {
	stderr := bytes.NewBuffer(nil)
	_, _, err := r.executable.Execute(pexec.Execution{
		Args:   []string{"container", "rm", containerID},
		Stderr: stderr,
	})
	if err != nil {
		return fmt.Errorf("failed to remove docker container: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return nil
}

type DockerContainerInspect struct {
	executable Executable
}

func (i DockerContainerInspect) Execute(containerID string) (Container, error) {
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	_, _, err := i.executable.Execute(pexec.Execution{
		Args:   []string{"container", "inspect", containerID},
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		return Container{}, fmt.Errorf("failed to inspect docker container: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	container, err := NewContainerFromInspectOutput(stdout.Bytes())
	if err != nil {
		return Container{}, fmt.Errorf("failed to inspect docker container: %w", err)
	}

	return container, nil
}
