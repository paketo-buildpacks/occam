package occam

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/paketo-buildpacks/packit/v2/pexec"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(pexec.Execution) error
}

//go:generate faux --interface DockerImageInspectClient --output fakes/docker_image_inspect_client.go
type DockerImageInspectClient interface {
	Execute(ref string) (Image, error)
}

type Pack struct {
	Build   PackBuild
	Builder PackBuilder
}

func NewPack() Pack {
	executable := pexec.NewExecutable("pack")

	return Pack{
		Build: PackBuild{
			executable:               executable,
			dockerImageInspectClient: NewDocker().Image.Inspect,
		},
		Builder: PackBuilder{
			Inspect: PackBuilderInspect{
				executable: executable,
			},
		},
	}
}

func (p Pack) WithExecutable(executable Executable) Pack {
	p.Build.executable = executable
	p.Builder.Inspect.executable = executable
	return p
}

func (p Pack) WithDockerImageInspectClient(client DockerImageInspectClient) Pack {
	p.Build.dockerImageInspectClient = client
	return p
}

func (p Pack) WithVerbose() Pack {
	p.Build.verbose = true
	return p
}

func (p Pack) WithNoColor() Pack {
	p.Build.noColor = true
	return p
}

type PackBuild struct {
	executable               Executable
	dockerImageInspectClient DockerImageInspectClient

	verbose bool
	noColor bool

	buildpacks          []string
	extensions          []string
	network             string
	builder             string
	clearCache          bool
	env                 map[string]string
	trustBuilder        bool
	pullPolicy          string
	sbomOutputDir       string
	volumes             []string
	gid                 string
	runImage            string
	additionalBuildArgs []string

	// TODO: remove after deprecation period
	noPull bool
}

func (pb PackBuild) WithAdditionalBuildArgs(args ...string) PackBuild {
	pb.additionalBuildArgs = append(pb.additionalBuildArgs, args...)
	return pb
}

func (pb PackBuild) WithRunImage(runImage string) PackBuild {
	pb.runImage = runImage
	return pb
}

func (pb PackBuild) WithBuildpacks(buildpacks ...string) PackBuild {
	pb.buildpacks = append(pb.buildpacks, buildpacks...)
	return pb
}

func (pb PackBuild) WithExtensions(extensions ...string) PackBuild {
	pb.extensions = append(pb.extensions, extensions...)
	return pb
}

func (pb PackBuild) WithNetwork(name string) PackBuild {
	pb.network = name
	return pb
}

func (pb PackBuild) WithBuilder(name string) PackBuild {
	pb.builder = name
	return pb
}

func (pb PackBuild) WithClearCache() PackBuild {
	pb.clearCache = true
	return pb
}

func (pb PackBuild) WithEnv(env map[string]string) PackBuild {
	pb.env = env
	return pb
}

func (pb PackBuild) WithGID(gid string) PackBuild {
	pb.gid = gid
	return pb
}

// Deprecated: Use WithPullPolicy("never") instead.
func (pb PackBuild) WithNoPull() PackBuild {
	pb.noPull = true
	return pb
}

func (pb PackBuild) WithPullPolicy(pullPolicy string) PackBuild {
	pb.pullPolicy = pullPolicy
	return pb
}

func (pb PackBuild) WithSBOMOutputDir(output string) PackBuild {
	pb.sbomOutputDir = output
	return pb
}

func (pb PackBuild) WithTrustBuilder() PackBuild {
	pb.trustBuilder = true
	return pb
}

func (pb PackBuild) WithVolumes(volumes ...string) PackBuild {
	pb.volumes = append(pb.volumes, volumes...)
	return pb
}

func (pb PackBuild) Execute(name, path string) (Image, fmt.Stringer, error) {
	args := []string{"build", name}

	if pb.verbose {
		args = append(args, "--verbose")
	}

	if pb.noColor {
		args = append(args, "--no-color")
	}

	args = append(args, "--path", path)

	for _, buildpack := range pb.buildpacks {
		args = append(args, "--buildpack", buildpack)
	}

	for _, extension := range pb.extensions {
		args = append(args, "--extension", extension)
	}

	if pb.network != "" {
		args = append(args, "--network", pb.network)
	}

	if pb.builder != "" {
		args = append(args, "--builder", pb.builder)
	}

	if pb.clearCache {
		args = append(args, "--clear-cache")
	}

	if len(pb.env) != 0 {
		var variables []string
		for key, value := range pb.env {
			variables = append(variables, fmt.Sprintf("%s=%s", key, value))
		}

		sort.Strings(variables)

		for _, v := range variables {
			args = append(args, "--env", v)
		}
	}

	if pb.noPull {
		args = append(args, "--no-pull")
	}

	if pb.pullPolicy != "" {
		args = append(args, "--pull-policy", pb.pullPolicy)
	}

	if pb.sbomOutputDir != "" {
		args = append(args, "--sbom-output-dir", pb.sbomOutputDir)
	}

	if pb.trustBuilder {
		args = append(args, "--trust-builder")
	}

	for _, volume := range pb.volumes {
		args = append(args, "--volume", volume)
	}

	if pb.gid != "" {
		args = append(args, "--gid", pb.gid)
	}

	if pb.runImage != "" {
		args = append(args, "--run-image", pb.runImage)
	}

	packEnv := os.Environ()
	packEnv = append(packEnv, fmt.Sprintf("PACK_VOLUME_KEY=%s-volume", name))

	args = append(args, pb.additionalBuildArgs...)

	buildLogBuffer := bytes.NewBuffer(nil)
	err := pb.executable.Execute(pexec.Execution{
		Args:   args,
		Stdout: buildLogBuffer,
		Stderr: buildLogBuffer,
		Env:    packEnv,
	})
	if err != nil {
		return Image{}, buildLogBuffer, fmt.Errorf("failed to pack build: %w\n\nOutput:\n%s", err, buildLogBuffer)
	}

	image, err := pb.dockerImageInspectClient.Execute(name)
	if err != nil {
		return Image{}, buildLogBuffer, fmt.Errorf("failed to pack build: %w", err)
	}

	return image, buildLogBuffer, nil
}

type PackBuilder struct {
	Inspect PackBuilderInspect
}

type PackBuilderInspect struct {
	executable Executable
}

func (pbi PackBuilderInspect) Execute(names ...string) (Builder, error) {
	args := []string{"builder", "inspect"}
	args = append(args, names...)
	args = append(args, "--output", "json")

	buffer := bytes.NewBuffer(nil)
	err := pbi.executable.Execute(pexec.Execution{
		Args:   args,
		Stdout: buffer,
	})
	if err != nil {
		return Builder{}, fmt.Errorf("failed to pack builder inspect: %w\n\nOutput:\n%s", err, buffer)
	}

	var builder Builder
	err = json.NewDecoder(buffer).Decode(&builder)
	if err != nil {
		return Builder{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return builder, nil
}
