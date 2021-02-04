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

func testServe(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		matcher types.GomegaMatcher
	)

	it.Before(func() {
		matcher = matchers.Serve("some string", "8080")
	})

	context("Match", func() {
		var (
			actual interface{}
			server *httptest.Server
		)

		context("when the http request succeeds and response contains substring", func() {
			it.Before(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					if req.Method == http.MethodHead {
						http.Error(w, "NotFound", http.StatusNotFound)
						return
					}

					switch req.URL.Path {
					case "/":
						w.WriteHeader(http.StatusOK)
						fmt.Fprintln(w, "some string")
					default:
						fmt.Fprintln(w, "unknown path")
						t.Fatal(fmt.Sprintf("unknown path: %s", req.URL.Path))
					}
				}))

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
			it("returns true", func() {
				result, err := matcher.Match(actual)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeTrue())
			})
		})

		context("the http response is nil", func() {
			it.Before(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					if req.Method == http.MethodHead {
						http.Error(w, "NotFound", http.StatusNotFound)
						return
					}

					switch req.URL.Path {
					case "/":
						// do nothing
					default:
						fmt.Fprintln(w, "unknown path")
						t.Fatal(fmt.Sprintf("unknown path: %s", req.URL.Path))
					}
				}))

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
			it("returns false", func() {
				result, err := matcher.Match(actual)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeFalse())
			})
		})

		context("the response status code is not OK", func() {
			it.Before(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					if req.Method == http.MethodHead {
						http.Error(w, "NotFound", http.StatusNotFound)
						return
					}

					switch req.URL.Path {
					case "/":
						w.WriteHeader(http.StatusNotFound)
						fmt.Fprintln(w, "some string")
					default:
						fmt.Fprintln(w, "unknown path")
						t.Fatal(fmt.Sprintf("unknown path: %s", req.URL.Path))
					}
				}))

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
			it("returns false", func() {
				result, err := matcher.Match(actual)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeFalse())
			})
		})

		context("the actual response does not match expected response", func() {
			it.Before(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					if req.Method == http.MethodHead {
						http.Error(w, "NotFound", http.StatusNotFound)
						return
					}

					switch req.URL.Path {
					case "/":
						w.WriteHeader(http.StatusOK)
						fmt.Fprintln(w, "another string")
					default:
						fmt.Fprintln(w, "unknown path")
						t.Fatal(fmt.Sprintf("unknown path: %s", req.URL.Path))
					}
				}))

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
			it("returns false", func() {
				result, err := matcher.Match(actual)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeFalse())
			})
		})

		context("failure cases", func() {
			context("the port is not in the container port mapping", func() {
				it.Before(func() {
					server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						if req.Method == http.MethodHead {
							http.Error(w, "NotFound", http.StatusNotFound)
							return
						}

						switch req.URL.Path {
						case "/":
							w.WriteHeader(http.StatusOK)
							fmt.Fprintln(w, "some string")
						default:
							fmt.Fprintln(w, "unknown path")
							t.Fatal(fmt.Sprintf("unknown path: %s", req.URL.Path))
						}
					}))

					actual = occam.Container{
						Ports: map[string]string{
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
				it("returns an error", func() {
					result, err := matcher.Match(actual)
					Expect(err).To(MatchError(ContainSubstring("ServeMatcher looking for response from container port 8080 which is not in container port map")))
					Expect(result).To(BeFalse())
				})
			})

			context("the request URL is malformed", func() {
				it.Before(func() {
					server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						if req.Method == http.MethodHead {
							http.Error(w, "NotFound", http.StatusNotFound)
							return
						}

						switch req.URL.Path {
						case "/":
							w.WriteHeader(http.StatusOK)
							fmt.Fprintln(w, "some string")
						default:
							fmt.Fprintln(w, "unknown path")
							t.Fatal(fmt.Sprintf("unknown path: %s", req.URL.Path))
						}
					}))

					actual = occam.Container{
						Ports: map[string]string{
							"8080": "malformed port",
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
				it("returns an error", func() {
					result, err := matcher.Match(actual)
					Expect(err).To(HaveOccurred())
					Expect(result).To(BeFalse())
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

			matcher = &matchers.ServeMatcher{
				Docker:           occam.NewDocker().WithExecutable(executable),
				ActualResponse:   "actual response",
				ExpectedResponse: "expected response",
			}

			actual = occam.Container{
				ID: "some-container-id",
			}
		})

		it("returns a useful error message", func() {
			message := matcher.FailureMessage(actual)
			Expect(message).To(ContainSubstring(strings.TrimSpace(`
Expected the response from docker container some-container-id:

	actual response

to contain:

	expected response

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

			matcher = &matchers.ServeMatcher{
				Docker:           occam.NewDocker().WithExecutable(executable),
				ActualResponse:   "actual response",
				ExpectedResponse: "expected response",
			}

			actual = occam.Container{
				ID: "some-container-id",
			}
		})

		it("returns a useful error message", func() {
			message := matcher.NegatedFailureMessage(actual)
			Expect(message).To(ContainSubstring(strings.TrimSpace(`
Expected the response from docker container some-container-id:

	actual response

not to contain:

	expected response

Container logs:

some logs
	`)))
		})
	})
}
