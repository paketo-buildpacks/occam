package occam

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/packit/pexec"
)

type DockerImage struct {
	executable Executable
}

type Docker struct {
	Image DockerImage
}

func NewDocker() Docker {
	return Docker{
		Image: DockerImage{
			executable: pexec.NewExecutable("docker", lager.NewLogger("docker")),
		},
	}
}

func (d Docker) WithExecutable(executable Executable) Docker {
	d.Image.executable = executable
	return d
}

func (di DockerImage) Inspect(ref string) (Image, error) {
	inspectOutput := bytes.NewBuffer(nil)
	errorBuffer := bytes.NewBuffer(nil)
	_, _, err := di.executable.Execute(pexec.Execution{
		Args:   []string{"image", "inspect", ref},
		Stdout: inspectOutput,
		Stderr: errorBuffer,
	})
	if err != nil {
		return Image{}, fmt.Errorf("failed to inspect docker image: %w: %s", err, strings.TrimSpace(errorBuffer.String()))
	}

	var inspect []struct {
		ID     string `json:"Id"`
		Config struct {
			Labels struct {
				LifecycleMetadata string `json:"io.buildpacks.lifecycle.metadata"`
			} `json:"Labels"`
		} `json:"Config"`
	}
	err = json.Unmarshal(inspectOutput.Bytes(), &inspect)
	if err != nil {
		return Image{}, fmt.Errorf("failed to inspect docker image: %w", err)
	}

	var metadata struct {
		Buildpacks []struct {
			Key    string `json:"key"`
			Layers map[string]struct {
				SHA    string `json:"sha"`
				Build  bool   `json:"build"`
				Launch bool   `json:"launch"`
				Cache  bool   `json:"cache"`
			} `json:"layers"`
		} `json:"buildpacks"`
	}
	err = json.Unmarshal([]byte(inspect[0].Config.Labels.LifecycleMetadata), &metadata)
	if err != nil {
		return Image{}, fmt.Errorf("failed to inspect docker image: %w", err)
	}

	var buildpacks []ImageBuildpackMetadata
	for _, buildpack := range metadata.Buildpacks {
		layers := map[string]ImageBuildpackMetadataLayer{}
		for name, layer := range buildpack.Layers {
			layers[name] = ImageBuildpackMetadataLayer{
				SHA:    layer.SHA,
				Build:  layer.Build,
				Launch: layer.Launch,
				Cache:  layer.Cache,
			}
		}

		buildpacks = append(buildpacks, ImageBuildpackMetadata{
			Key:    buildpack.Key,
			Layers: layers,
		})
	}

	return Image{
		ID:         inspect[0].ID,
		Buildpacks: buildpacks,
	}, nil
}
