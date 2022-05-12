package occam

import (
	"context"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestContainers struct {
	env             map[string]string
	publishPorts    []string
	waitStrategy    wait.Strategy
	containerMounts []testcontainers.ContainerMount
	noStart         bool
	startupTimeout  int
}

func NewTestContainers() TestContainers {
	return TestContainers{startupTimeout: 20}
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

func (r TestContainers) WithTimeout(t int) TestContainers {
	r.startupTimeout = t
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
	c, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(r.startupTimeout))
	defer cancel()

	return testcontainers.GenericContainer(c, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          !r.noStart,
	})
}
