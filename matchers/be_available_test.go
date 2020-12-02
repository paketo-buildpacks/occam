package matchers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/onsi/gomega/types"
	"github.com/paketo-buildpacks/occam"
	"github.com/paketo-buildpacks/occam/matchers"
	"github.com/paketo-buildpacks/occam/matchers/fakes"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

//go:generate faux --package github.com/paketo-buildpacks/occam --interface Executable --output fakes/executable.go

func testBeAvailable(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		matcher types.GomegaMatcher
	)

	it.Before(func() {
		matcher = matchers.BeAvailable()
	})

	context("Match", func() {
		var (
			actual interface{}
			server *httptest.Server
		)

		it.Before(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}))
			serverURL, err := url.Parse(server.URL)
			Expect(err).NotTo(HaveOccurred())

			actual = occam.Container{
				Ports: map[string]string{
					"8080": serverURL.Port(),
					"3000": "1234",
					"4000": "4321",
					"5000": "5678",
				},
				Env: map[string]string{
					"PORT": "8080",
				},
			}
		})

		it.After(func() {
			server.Close()
		})

		context("when the http request succeeds", func() {
			it("returns true", func() {
				result, err := matcher.Match(actual)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeTrue())
			})
		})

		context("when the http request fails", func() {
			it.Before(func() {
				server.Close()
			})

			it("returns false", func() {
				result, err := matcher.Match(actual)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeFalse())
			})
		})

		context("failure cases", func() {
			context("when the actual is not a container", func() {
				it("returns an error", func() {
					_, err := matcher.Match("not a container")
					Expect(err).To(MatchError("BeAvailableMatcher expects an occam.Container, received string"))
				})
			})
		})
	})

	context("FailureMessage", func() {
		var actual interface{}

		it.Before(func() {
			executable := &fakes.Executable{}
			executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
				fmt.Fprintln(execution.Stdout, "some logs")
				return nil
			}

			matcher = &matchers.BeAvailableMatcher{
				Docker: occam.NewDocker().WithExecutable(executable),
			}

			actual = occam.Container{
				ID: "some-container-id",
			}
		})

		it("returns a useful error message", func() {
			message := matcher.FailureMessage(actual)
			Expect(message).To(ContainSubstring(strings.TrimSpace(`
Expected
	docker container id: some-container-id
to be available.

Container logs:

some logs
`)))
		})
	})

	context("NegatedFailureMessage", func() {
		var actual interface{}

		it.Before(func() {
			executable := &fakes.Executable{}
			executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
				fmt.Fprintln(execution.Stdout, "some logs")
				return nil
			}

			matcher = &matchers.BeAvailableMatcher{
				Docker: occam.NewDocker().WithExecutable(executable),
			}

			actual = occam.Container{
				ID: "some-container-id",
			}
		})

		it("returns a useful error message", func() {
			message := matcher.NegatedFailureMessage(actual)
			Expect(message).To(ContainSubstring(strings.TrimSpace(`
Expected
	docker container id: some-container-id
not to be available.

Container logs:

some logs
`)))
		})
	})
}
