package matchers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/onsi/gomega/types"
	"github.com/paketo-buildpacks/occam"
)

type ServeMatcherInterface interface {
	types.GomegaMatcher
	OnPort(string) *ServeMatcher
	WithEndpoint(string) *ServeMatcher
}

func Serve(expectedResponse string) ServeMatcherInterface {
	return &ServeMatcher{
		ExpectedResponse: expectedResponse,
		Docker:           occam.NewDocker(),
		ActualResponse:   "",
	}
}

type ServeMatcher struct {
	ExpectedResponse string
	port             string
	endpoint         string
	Docker           occam.Docker
	ActualResponse   string
}

func (sm *ServeMatcher) OnPort(port string) *ServeMatcher {
	sm.port = port
	return sm
}

func (sm *ServeMatcher) WithEndpoint(endpoint string) *ServeMatcher {
	sm.endpoint = endpoint
	return sm
}

func (matcher *ServeMatcher) Match(actual interface{}) (success bool, err error) {
	container, ok := actual.(occam.Container)
	if !ok {
		return false, fmt.Errorf("ServeMatcher expects an occam.Container, received %T", actual)
	}

	// no port specified, and there's only one to choose from
	if matcher.port == "" {
		if len(container.Ports) == 1 {
			for p := range container.Ports {
				matcher.port = p
				break
			}
		} else {
			return false, fmt.Errorf("container has multiple port mappings, but none were specified. Please specify via the OnPort method")
		}
	}

	if _, ok := container.Ports[matcher.port]; !ok {
		// EITHER: you have multiple ports, didn't specify OR you specified a bad port
		return false, fmt.Errorf("ServeMatcher looking for response from container port %s which is not in container port map", matcher.port)
	}

	response, err := http.Get(fmt.Sprintf("http://localhost:%s%s", container.HostPort(matcher.port), matcher.endpoint))

	if err != nil {
		return false, err
	}

	if response != nil {
		defer response.Body.Close()
		content, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return false, err
		}

		matcher.ActualResponse = string(content)

		if response.StatusCode == http.StatusOK && strings.Contains(string(content), matcher.ExpectedResponse) {
			return true, nil
		}
	}
	return false, nil
}

func (matcher *ServeMatcher) FailureMessage(actual interface{}) (message string) {
	container := actual.(occam.Container)

	message = fmt.Sprintf("Expected the response from docker container %s:\n\n\t%s\n\nto contain:\n\n\t%s",
		container.ID,
		matcher.ActualResponse,
		matcher.ExpectedResponse,
	)

	if logs, _ := matcher.Docker.Container.Logs.Execute(container.ID); logs != nil {
		message = fmt.Sprintf("%s\n\nContainer logs:\n\n%s", message, logs)
	}

	return message
}

func (matcher *ServeMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	container := actual.(occam.Container)

	message = fmt.Sprintf("Expected the response from docker container %s:\n\n\t%s\n\nnot to contain:\n\n\t%s",
		container.ID,
		matcher.ActualResponse,
		matcher.ExpectedResponse,
	)

	if logs, _ := matcher.Docker.Container.Logs.Execute(container.ID); logs != nil {
		message = fmt.Sprintf("%s\n\nContainer logs:\n\n%s", message, logs)
	}

	return message
}
