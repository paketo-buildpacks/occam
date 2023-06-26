package occam

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/docker/docker/client"
	name "github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	daemon "github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/paketo-buildpacks/packit/v2/pexec"
)

//go:generate faux --interface DockerDaemonClient --output fakes/daemon_client.go
type DockerDaemonClient interface {
	daemon.Client
}

type Docker struct {
	Image struct {
		ExportToOCI DockerImageOCI
		Inspect     DockerImageInspect
		Remove      DockerImageRemove
		Tag         DockerImageTag
	}

	Container struct {
		Copy    DockerContainerCopy
		Exec    DockerContainerExec
		Inspect DockerContainerInspect
		Logs    DockerContainerLogs
		Remove  DockerContainerRemove
		Restart DockerContainerRestart
		Run     DockerContainerRun
		Stop    DockerContainerStop
	}

	Volume struct {
		Remove DockerVolumeRemove
	}

	Pull DockerPull
}

func NewDocker() Docker {
	var docker Docker
	executable := pexec.NewExecutable("docker")

	docker.Image.Inspect = DockerImageInspect{executable: executable}
	docker.Image.Remove = DockerImageRemove{executable: executable}
	docker.Image.Tag = DockerImageTag{executable: executable}
	docker.Image.ExportToOCI = DockerImageOCI{}

	docker.Container.Copy = DockerContainerCopy{executable: executable}
	docker.Container.Exec = DockerContainerExec{executable: executable}
	docker.Container.Inspect = DockerContainerInspect{executable: executable}
	docker.Container.Logs = DockerContainerLogs{executable: executable}
	docker.Container.Run = DockerContainerRun{
		executable: executable,
		inspect:    docker.Container.Inspect,
	}

	docker.Container.Remove = DockerContainerRemove{executable: executable}
	docker.Container.Restart = DockerContainerRestart{executable: executable}
	docker.Container.Stop = DockerContainerStop{executable: executable}

	docker.Volume.Remove = DockerVolumeRemove{executable: executable}

	docker.Pull = DockerPull{executable: executable}

	return docker
}

func (d Docker) WithExecutable(executable Executable) Docker {
	d.Image.Inspect.executable = executable
	d.Image.Remove.executable = executable
	d.Image.Tag.executable = executable

	d.Container.Copy.executable = executable
	d.Container.Exec.executable = executable
	d.Container.Inspect.executable = executable
	d.Container.Logs.executable = executable
	d.Container.Restart.executable = executable
	d.Container.Remove.executable = executable
	d.Container.Run.executable = executable
	d.Container.Run.inspect = d.Container.Inspect
	d.Container.Stop.executable = executable

	d.Volume.Remove.executable = executable

	d.Pull.executable = executable

	return d
}

type DockerImageInspect struct {
	executable Executable
}

func (i DockerImageInspect) Execute(ref string) (Image, error) {
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	err := i.executable.Execute(pexec.Execution{
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
	force      bool
}

func (r DockerImageRemove) WithForce() DockerImageRemove {
	r.force = true
	return r
}

func (r DockerImageRemove) Execute(ref string) error {
	args := []string{"image", "remove", ref}

	if r.force {
		args = append(args, "--force")
	}

	stderr := bytes.NewBuffer(nil)
	err := r.executable.Execute(pexec.Execution{
		Args:   args,
		Stderr: stderr,
	})
	if err != nil {
		return fmt.Errorf("failed to remove docker image: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return nil
}

type DockerImageTag struct {
	executable Executable
}

func (r DockerImageTag) Execute(ref, target string) error {
	stderr := bytes.NewBuffer(nil)
	err := r.executable.Execute(pexec.Execution{
		Args:   []string{"image", "tag", ref, target},
		Stderr: stderr,
	})
	if err != nil {
		return fmt.Errorf("failed to tag docker image: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return nil
}

type DockerImageOCI struct {
	client      DockerDaemonClient
	nameOptions []name.Option
}

func (r DockerImageOCI) WithNameOptions(opts ...name.Option) DockerImageOCI {
	r.nameOptions = opts
	return r
}

func (r DockerImageOCI) WithClient(client DockerDaemonClient) DockerImageOCI {
	r.client = client
	return r
}

func (r DockerImageOCI) Execute(ref string) (v1.Image, error) {
	if r.client == nil {
		client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return nil, err
		}

		r.client = client
	}
	nameRef, err := name.ParseReference(ref, r.nameOptions...)
	if err != nil {
		return nil, err
	}

	return daemon.Image(nameRef, daemon.WithClient(r.client))
}

type DockerContainerRun struct {
	executable Executable
	inspect    DockerContainerInspect

	command      string
	commandArgs  []string
	direct       bool
	entrypoint   string
	env          map[string]string
	memory       string
	network      string
	publishAll   bool
	publishPorts []string
	tty          bool
	volumes      []string
	readOnly     bool
	mounts       []string
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

func (r DockerContainerRun) WithCommandArgs(commandArgs []string) DockerContainerRun {
	r.commandArgs = commandArgs
	return r
}

func (r DockerContainerRun) WithDirect() DockerContainerRun {
	r.direct = true
	return r
}

func (r DockerContainerRun) WithTTY() DockerContainerRun {
	r.tty = true
	return r
}

func (r DockerContainerRun) WithEntrypoint(entrypoint string) DockerContainerRun {
	r.entrypoint = entrypoint
	return r
}

func (r DockerContainerRun) WithPublish(value string) DockerContainerRun {
	r.publishPorts = append(r.publishPorts, value)
	return r
}

func (r DockerContainerRun) WithPublishAll() DockerContainerRun {
	r.publishAll = true
	return r
}

// Deprecated: Use WithVolumes(...volumes) instead.
func (r DockerContainerRun) WithVolume(volume string) DockerContainerRun {
	r.volumes = append(r.volumes, volume)
	return r
}

func (r DockerContainerRun) WithVolumes(volumes ...string) DockerContainerRun {
	r.volumes = append(r.volumes, volumes...)
	return r
}

func (r DockerContainerRun) WithNetwork(network string) DockerContainerRun {
	r.network = network
	return r
}

func (r DockerContainerRun) WithReadOnly() DockerContainerRun {
	r.readOnly = true
	return r
}

func (r DockerContainerRun) WithMounts(mounts ...string) DockerContainerRun {
	r.mounts = append(r.mounts, mounts...)
	return r
}

func (r DockerContainerRun) Execute(imageID string) (Container, error) {
	args := []string{"container", "run", "--detach"}

	if r.tty {
		args = append(args, "--tty")
	}

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

	if len(r.publishPorts) > 0 {
		for _, port := range r.publishPorts {
			args = append(args, "--publish", port)
		}
	}

	if r.publishAll {
		args = append(args, "--publish-all")
	}

	if r.memory != "" {
		args = append(args, "--memory", r.memory)
	}

	if r.entrypoint != "" {
		args = append(args, "--entrypoint", r.entrypoint)
	}

	if r.network != "" {
		args = append(args, "--network", r.network)
	}

	for _, volume := range r.volumes {
		args = append(args, "--volume", volume)
	}

	if r.readOnly {
		args = append(args, "--read-only")
	}

	for _, mount := range r.mounts {
		args = append(args, "--mount", mount)
	}

	args = append(args, imageID)

	if r.direct {
		args = append(args, "--")
	}

	if r.command != "" {
		args = append(args, r.command)
	}
	args = append(args, r.commandArgs...)

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	err := r.executable.Execute(pexec.Execution{
		Args:   args,
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		return Container{}, fmt.Errorf("failed to run docker container: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return r.inspect.Execute(strings.TrimSpace(stdout.String()))
}

type DockerContainerRestart struct {
	executable Executable
}

func (r DockerContainerRestart) Execute(containerID string) error {
	stderr := bytes.NewBuffer(nil)
	err := r.executable.Execute(pexec.Execution{
		Args:   []string{"container", "restart", containerID},
		Stderr: stderr,
	})
	if err != nil {
		return fmt.Errorf("failed to restart docker container: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return nil
}

type DockerContainerRemove struct {
	executable Executable
}

func (r DockerContainerRemove) Execute(containerID string) error {
	stderr := bytes.NewBuffer(nil)
	err := r.executable.Execute(pexec.Execution{
		Args:   []string{"container", "rm", containerID, "--force"},
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
	err := i.executable.Execute(pexec.Execution{
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

type DockerContainerLogs struct {
	executable Executable
}

func (l DockerContainerLogs) Execute(containerID string) (fmt.Stringer, error) {
	output := bytes.NewBuffer(nil)
	err := l.executable.Execute(pexec.Execution{
		Args:   []string{"container", "logs", containerID},
		Stdout: output,
		Stderr: output,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch docker container logs: %w: %s", err, strings.TrimSpace(output.String()))
	}

	return output, nil
}

type DockerContainerStop struct {
	executable Executable
}

func (s DockerContainerStop) Execute(containerID string) error {
	output := bytes.NewBuffer(nil)
	err := s.executable.Execute(pexec.Execution{
		Args:   []string{"container", "stop", containerID},
		Stdout: output,
		Stderr: output,
	})
	if err != nil {
		return fmt.Errorf("failed to stop docker container: %w: %s", err, strings.TrimSpace(output.String()))
	}

	return nil
}

type DockerContainerCopy struct {
	executable Executable
}

func (docker DockerContainerCopy) Execute(source, dest string) error {
	stderr := bytes.NewBuffer(nil)
	err := docker.executable.Execute(pexec.Execution{
		Args:   []string{"container", "cp", source, dest},
		Stderr: stderr,
	})
	if err != nil {
		return fmt.Errorf("'docker cp' failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return nil
}

type DockerContainerExec struct {
	executable  Executable
	stdin       io.Reader
	user        string
	interactive bool
}

func (e DockerContainerExec) WithStdin(stdin io.Reader) DockerContainerExec {
	e.stdin = stdin
	return e
}

func (e DockerContainerExec) WithUser(user string) DockerContainerExec {
	e.user = user
	return e
}

func (e DockerContainerExec) WithInteractive() DockerContainerExec {
	e.interactive = true
	return e
}

func (e DockerContainerExec) Execute(container string, arguments ...string) error {
	args := []string{"container", "exec"}
	if e.interactive {
		args = append(args, "--interactive")
	}
	if e.user != "" {
		args = append(args, "--user", e.user)
	}
	args = append(args, container)
	args = append(args, arguments...)

	stderr := bytes.NewBuffer(nil)
	err := e.executable.Execute(pexec.Execution{
		Args:   args,
		Stderr: stderr,
		Stdin:  e.stdin,
	})
	if err != nil {
		return fmt.Errorf("'docker exec' failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return nil
}

func (e DockerContainerExec) ExecuteBash(container, script string) error {
	return e.Execute(container, "/bin/bash", "-c", script)
}

type DockerVolumeRemove struct {
	executable Executable
}

func (r DockerVolumeRemove) Execute(volumes []string) error {
	args := []string{"volume", "rm", "--force"}
	args = append(args, volumes...)

	stderr := bytes.NewBuffer(nil)
	err := r.executable.Execute(pexec.Execution{
		Args:   args,
		Stderr: stderr,
	})
	if err != nil {
		return fmt.Errorf("failed to remove docker volume: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return nil
}

type DockerPull struct {
	executable Executable
}

func (p DockerPull) Execute(image string) error {

	stderr := bytes.NewBuffer(nil)
	err := p.executable.Execute(pexec.Execution{
		Args:   []string{"pull", image},
		Stderr: stderr,
	})
	if err != nil {
		return fmt.Errorf("failed to pull docker image: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return nil
}
