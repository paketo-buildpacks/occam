package occam

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestContainers struct {
	env             map[string]string
	publishPorts    []string
	waitStrategy    wait.Strategy
	containerMounts []testcontainers.ContainerMount
	noStart         bool
}

func NewTestContainers() TestContainers {
	return TestContainers{}
}

func (r TestContainers) WithEnv(env map[string]string) TestContainers {
	r.env = env
	return r
}

func (r TestContainers) WithMounts(containerMounts ...testcontainers.ContainerMount) TestContainers {
	r.containerMounts = append(r.containerMounts, containerMounts...)
	return r
}

func (r TestContainers) WithExposedPorts(values ...string) TestContainers {
	r.publishPorts = append(r.publishPorts, values...)
	return r
}

func (r TestContainers) WithWaitingFor(waitStrategy wait.Strategy) TestContainers {
	r.waitStrategy = waitStrategy
	return r
}

func (r TestContainers) WithNoStart() TestContainers {
	r.noStart = true
	return r
}

func (r TestContainers) Execute(imageID string) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        imageID,
		ExposedPorts: r.publishPorts,
		WaitingFor:   r.waitStrategy,
		Env:          r.env,
		Mounts:       r.containerMounts,

		SkipReaper: true,
	}

	return testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          !r.noStart,
	})
}
