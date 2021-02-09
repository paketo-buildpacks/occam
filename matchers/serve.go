package matchers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/onsi/gomega/types"
	"github.com/paketo-buildpacks/occam"
)

func Serve(expectedResponse string, port string, endpoint string) types.GomegaMatcher {
	return &ServeMatcher{
		ExpectedResponse: expectedResponse,
		port:             port,
		endpoint:         endpoint,
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

func (matcher *ServeMatcher) Match(actual interface{}) (success bool, err error) {
	container, ok := actual.(occam.Container)
	if !ok {
		return false, fmt.Errorf("ServeMatcher expects an occam.Container, received %T", actual)
	}

	if _, ok := container.Ports[matcher.port]; !ok {
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
