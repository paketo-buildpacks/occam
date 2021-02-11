package matchers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"

	"github.com/onsi/gomega/types"
	"github.com/paketo-buildpacks/occam"
)

func Serve(expected interface{}) *ServeMatcher {
	return &ServeMatcher{
		expected: expected,
		docker:   occam.NewDocker(),
	}
}

type ServeMatcher struct {
	expected interface{}
	port     int
	endpoint string
	docker   occam.Docker
	response string
}

func (sm *ServeMatcher) OnPort(port int) *ServeMatcher {
	sm.port = port
	return sm
}

func (sm *ServeMatcher) WithEndpoint(endpoint string) *ServeMatcher {
	sm.endpoint = endpoint
	return sm
}

func (sm *ServeMatcher) WithDocker(docker occam.Docker) *ServeMatcher {
	sm.docker = docker
	return sm
}

func (sm *ServeMatcher) Match(actual interface{}) (success bool, err error) {
	container, ok := actual.(occam.Container)
	if !ok {
		return false, fmt.Errorf("ServeMatcher expects an occam.Container, received %T", actual)
	}

	// no port specified, and there's only one to choose from
	port := strconv.Itoa(sm.port)
	if port == "0" {
		if len(container.Ports) == 1 {
			for p := range container.Ports {
				port = p
				break
			}
		} else {
			return false, fmt.Errorf("container has multiple port mappings, but none were specified. Please specify via the OnPort method")
		}
	}

	if _, ok := container.Ports[port]; !ok {
		// EITHER: you have multiple ports and didn't specify OR you specified a bad port
		return false, fmt.Errorf("ServeMatcher looking for response from container port %s which is not in container port map", port)
	}

	response, err := http.Get(fmt.Sprintf("http://localhost:%s%s", container.HostPort(port), sm.endpoint))

	if err != nil {
		return false, err
	}

	if response != nil {
		defer response.Body.Close()
		content, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return false, err
		}

		sm.response = string(content)

		if response.StatusCode == http.StatusOK {
			match, err := sm.compare(string(content), sm.expected)
			if err != nil {
				return false, err
			}

			return match, nil
		}
	}
	return false, nil
}

func (sm *ServeMatcher) compare(actual string, expected interface{}) (bool, error) {
	if m, ok := expected.(types.GomegaMatcher); ok {
		match, err := m.Match(actual)
		if err != nil {
			return false, err
		}

		return match, nil
	}

	return reflect.DeepEqual(actual, expected), nil
}

func (sm *ServeMatcher) FailureMessage(actual interface{}) (message string) {
	container := actual.(occam.Container)

	message = fmt.Sprintf("Expected the response from docker container %s:\n\n\t%s\n\nto contain:\n\n\t%s",
		container.ID,
		sm.response,
		sm.expected,
	)

	if logs, _ := sm.docker.Container.Logs.Execute(container.ID); logs != nil {
		message = fmt.Sprintf("%s\n\nContainer logs:\n\n%s", message, logs)
	}

	return message
}

func (sm *ServeMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	container := actual.(occam.Container)

	message = fmt.Sprintf("Expected the response from docker container %s:\n\n\t%s\n\nnot to contain:\n\n\t%s",
		container.ID,
		sm.response,
		sm.expected,
	)

	if logs, _ := sm.docker.Container.Logs.Execute(container.ID); logs != nil {
		message = fmt.Sprintf("%s\n\nContainer logs:\n\n%s", message, logs)
	}

	return message
}
