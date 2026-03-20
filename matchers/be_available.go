package matchers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/onsi/gomega/types"
	"github.com/paketo-buildpacks/occam"
)

// BeAvailable matches if the actual occam.Container is running AND an
// HTTP request to at least one of its exposed ports completes without error.
func BeAvailable() types.GomegaMatcher {
	return &BeAvailableMatcher{
		Docker: occam.NewDocker(),
	}
}

type BeAvailableMatcher struct {
	Docker occam.Docker
}

func (*BeAvailableMatcher) Match(actual interface{}) (bool, error) {
	container, ok := actual.(occam.Container)
	if !ok {
		return false, fmt.Errorf("BeAvailableMatcher expects an occam.Container, received %T", actual)
	}

	// Get a container port in order to look up the corresponding host port.
	for port := range container.Ports {
		response, err := http.Get(fmt.Sprintf("http://%s:%s", container.Host(), container.HostPort(port)))
		if response != nil {
			if err := response.Body.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to close response body: %s\n", err)
			}
		}

		if err == nil {
			return true, nil
		}
	}

	return false, nil
}

func (m *BeAvailableMatcher) FailureMessage(actual interface{}) string {
	container := actual.(occam.Container)
	message := fmt.Sprintf("Expected\n\tdocker container id: %s\nto be available.", container.ID)

	if logs, _ := m.Docker.Container.Logs.Execute(container.ID); logs != nil {
		message = fmt.Sprintf("%s\n\nContainer logs:\n\n%s", message, logs)
	}

	return message
}

func (m *BeAvailableMatcher) NegatedFailureMessage(actual interface{}) string {
	container := actual.(occam.Container)
	message := fmt.Sprintf("Expected\n\tdocker container id: %s\nnot to be available.", container.ID)

	if logs, _ := m.Docker.Container.Logs.Execute(container.ID); logs != nil {
		message = fmt.Sprintf("%s\n\nContainer logs:\n\n%s", message, logs)
	}

	return message
}
