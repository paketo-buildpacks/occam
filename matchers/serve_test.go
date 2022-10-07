package matchers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/paketo-buildpacks/occam/matchers"
	"github.com/paketo-buildpacks/occam/matchers/fakes"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

//go:generate faux --package github.com/paketo-buildpacks/occam --interface Executable --output fakes/executable.go

func testServe(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		matcher *matchers.ServeMatcher
		server  *httptest.Server
		port    string
	)

	it.Before(func() {
		count := 0

		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Method == http.MethodHead {
				http.Error(w, "NotFound", http.StatusNotFound)
				return
			}

			w.Header().Set("Date", "Sat, 17 Sep 2022 03:28:54 GMT")

			if count > 5 {
				w.Header().Set("X-Paketo-Count", fmt.Sprintf("%d", count))
			}
			count++

			switch req.URL.Path {
			case "/":
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, "some string")
			case "/redirect":
				w.Header()["Location"] = []string{"/"}
				w.WriteHeader(http.StatusMovedPermanently)
			case "/empty":
				// do nothing
			case "/teapot":
				w.WriteHeader(http.StatusTeapot)
			default:
				fmt.Fprintln(w, "unknown path")
				t.Fatalf("unknown path: %s", req.URL.Path)
			}
		}))

		serverURL, err := url.Parse(server.URL)
		Expect(err).NotTo(HaveOccurred())

		port = serverURL.Port()

		matcher = matchers.Serve("some string")
	})

	it.After(func() {
		server.Close()
	})

	context("Match", func() {
		context("the http request succeeds and response equals expected", func() {
			it("returns true", func() {
				result, err := matcher.Match(occam.Container{
					Ports: map[string]string{"8080": port},
					Env:   map[string]string{"PORT": "8080"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeTrue())
			})
		})

		context("the http request succeeds and response matches expected", func() {
			it.Before(func() {
				matcher = matchers.Serve(ContainSubstring("me str"))
			})

			it("returns true", func() {
				result, err := matcher.Match(occam.Container{
					Ports: map[string]string{"8080": port},
					Env:   map[string]string{"PORT": "8080"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeTrue())
			})
		})

		context("the http response is nil", func() {
			it.Before(func() {
				matcher = matcher.WithEndpoint("/empty")
			})

			it("returns false", func() {
				result, err := matcher.Match(occam.Container{
					Ports: map[string]string{"8080": port},
					Env:   map[string]string{"PORT": "8080"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeFalse())
			})
		})

		context("the response status code is not OK", func() {
			it.Before(func() {
				matcher = matcher.WithEndpoint("/teapot")
			})

			it("returns false", func() {
				result, err := matcher.Match(occam.Container{
					Ports: map[string]string{"8080": port},
					Env:   map[string]string{"PORT": "8080"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeFalse())
			})
		})

		context("the actual response does not match expected response", func() {
			it.Before(func() {
				matcher = matchers.Serve("another string")
			})

			it("returns false", func() {
				result, err := matcher.Match(occam.Container{
					Ports: map[string]string{"8080": port},
					Env:   map[string]string{"PORT": "8080"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeFalse())
			})
		})

		context("when there are multiple port mappings", func() {
			it("returns true", func() {
				result, err := matcher.Match(occam.Container{
					Ports: map[string]string{
						"8080": port,
						"3030": "3030",
						"4030": "4030",
						"5030": "5030",
					},
					Env: map[string]string{"PORT": "8080"},
				})
				Expect(err).To(MatchError(ContainSubstring("container has multiple port mappings, but none were specified. Please specify via the OnPort method")))
				Expect(result).To(BeFalse())
			})
		})

		context("when given a client", func() {
			var (
				redirectFunctionCalled bool
			)

			it.Before(func() {
				redirectFunctionCalled = false

				client := &http.Client{
					CheckRedirect: func(req *http.Request, via []*http.Request) error {
						redirectFunctionCalled = true
						return nil
					},
				}

				matcher = matcher.WithClient(client)
			})

			it("uses the provided client", func() {
				result, err := matcher.WithEndpoint("/redirect").Match(occam.Container{
					Ports: map[string]string{
						"8080": port,
					},
					Env: map[string]string{"PORT": "8080"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeTrue())

				Expect(redirectFunctionCalled).To(BeTrue())
			})
		})

		context("when given a port", func() {
			it.Before(func() {
				matcher = matcher.OnPort(8080)
			})

			it("returns true", func() {
				result, err := matcher.Match(occam.Container{
					Ports: map[string]string{
						"3030": "3030",
						"4030": "4030",
						"5030": "5030",
						"8080": port,
					},
					Env: map[string]string{"PORT": "8080"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeTrue())
			})
		})

		context("failure cases", func() {
			context("the port is not in the container port mapping", func() {
				it.Before(func() {
					executable := &fakes.Executable{}
					executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
						fmt.Fprintln(execution.Stdout, "some logs")
						return nil
					}

					docker := occam.NewDocker().WithExecutable(executable)

					matcher = matchers.Serve("some string").WithDocker(docker).OnPort(8080)
				})

				it("returns an error", func() {
					result, err := matcher.Match(occam.Container{
						Ports: map[string]string{"3000": "1234"},
						Env:   map[string]string{"PORT": "8080"},
					})
					Expect(err).To(MatchError(ContainSubstring("ServeMatcher looking for response from container port 8080 which is not in container port map")))
					Expect(err).To(MatchError(ContainSubstring("Container logs:\n\nsome logs\n")))
					Expect(result).To(BeFalse())
				})
			})

			context("the request URL is malformed", func() {
				it("returns an error", func() {
					result, err := matcher.Match(occam.Container{
						Ports: map[string]string{"8080": "malformed port"},
						Env:   map[string]string{"PORT": "8080"},
					})
					Expect(err).To(HaveOccurred())
					Expect(result).To(BeFalse())
				})
			})
		})

		context("WithHeader", func() {
			it("returns true with the header is found", func() {
				matcher = matchers.Serve(matchers.WithHeader("Content-Type", "text/plain; charset=utf-8"))

				result, err := matcher.Match(occam.Container{
					Ports: map[string]string{"8080": port},
					Env:   map[string]string{"PORT": "8080"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeTrue())
			})

			it("returns false with the header is not found", func() {
				matcher = matchers.Serve(matchers.WithHeader("Content-Type", "other"))

				result, err := matcher.Match(occam.Container{
					Ports: map[string]string{"8080": port},
					Env:   map[string]string{"PORT": "8080"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeFalse())
			})

			context("Eventually WithHeader still works", func() {
				it("eventually is true", func() {
					container := occam.Container{
						Ports: map[string]string{"8080": port},
						Env:   map[string]string{"PORT": "8080"},
					}

					Eventually(container).Should(matchers.Serve(matchers.WithHeader("X-Paketo-Count", "20")))
				})
			})
		})
	})

	context("when the matcher fails", func() {
		var actual occam.Container

		it.Before(func() {
			executable := &fakes.Executable{}
			executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
				fmt.Fprintln(execution.Stdout, "some logs")
				return nil
			}

			docker := occam.NewDocker().WithExecutable(executable)

			matcher = matchers.Serve("no such content").WithDocker(docker)

			actual = occam.Container{
				ID:    "some-container-id",
				Ports: map[string]string{"8080": port},
				Env:   map[string]string{"PORT": "8080"},
			}

			result, err := matcher.Match(actual)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeFalse())
		})

		context("FailureMessage", func() {
			it("returns a useful error message", func() {
				message := matcher.FailureMessage(actual)
				Expect(message).To(ContainSubstring(strings.TrimSpace(`
Expected the response from docker container some-container-id:

	some string

to contain:

	no such content

Container logs:

some logs
	`)))
			})

			it("returns a useful error message for WithHeader", func() {
				matcher = matchers.Serve(matchers.WithHeader("Content-Type", "other"))
				_, _ = matcher.Match(occam.Container{
					Ports: map[string]string{"8080": port},
					Env:   map[string]string{"PORT": "8080"},
				})

				message := matcher.FailureMessage(actual)
				Expect(message).To(ContainSubstring(strings.TrimSpace(`
Expected the response from docker container some-container-id:

	Header 'Content-Length=11'
	Header 'Content-Type=text/plain; charset=utf-8'
	Header 'Date=Sat, 17 Sep 2022 03:28:54 GMT'

to contain:

	Header 'Content-Type=other'`)))
			})
		})

		context("NegatedFailureMessage", func() {
			it("returns a useful error message", func() {
				message := matcher.NegatedFailureMessage(actual)
				Expect(message).To(ContainSubstring(strings.TrimSpace(`
Expected the response from docker container some-container-id:

	some string

not to contain:

	no such content

Container logs:

some logs
	`)))
			})
		})

		it("returns a useful error message for WithHeader", func() {
			matcher = matchers.Serve(matchers.WithHeader("Content-Type", "other"))
			_, _ = matcher.Match(occam.Container{
				Ports: map[string]string{"8080": port},
				Env:   map[string]string{"PORT": "8080"},
			})

			message := matcher.NegatedFailureMessage(actual)
			Expect(message).To(ContainSubstring(strings.TrimSpace(`
Expected the response from docker container some-container-id:

	Header 'Content-Length=11'
	Header 'Content-Type=text/plain; charset=utf-8'
	Header 'Date=Sat, 17 Sep 2022 03:28:54 GMT'

not to contain:

	Header 'Content-Type=other'`)))
		})
	})
}
