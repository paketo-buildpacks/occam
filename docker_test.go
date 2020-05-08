package occam_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/cloudfoundry/occam"
	"github.com/cloudfoundry/occam/fakes"
	"github.com/cloudfoundry/packit/pexec"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDocker(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		executable *fakes.Executable
		docker     occam.Docker
	)

	it.Before(func() {
		executable = &fakes.Executable{}

		docker = occam.NewDocker().WithExecutable(executable)
	})

	context("Image", func() {
		context("Inspect", func() {
			it.Before(func() {
				executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
					fmt.Fprintln(execution.Stdout, `[
						{
							"Id": "some-image-id",
							"Config": {
								"Labels": {
									"io.buildpacks.lifecycle.metadata": "{\"buildpacks\": [{\"key\": \"some-buildpack\", \"layers\": {\"some-layer\": {\"sha\": \"some-sha\", \"build\": true, \"launch\": true, \"cache\": true, \"data\": {\"some-key\": \"some-value\"}}}}]}"
								}
							}
						}
					]`)

					return nil
				}
			})

			it("returns an image given a name", func() {
				image, err := docker.Image.Inspect.Execute("some-app:latest")
				Expect(err).NotTo(HaveOccurred())
				Expect(image).To(Equal(occam.Image{
					ID: "some-image-id",
					Buildpacks: []occam.ImageBuildpackMetadata{
						{
							Key: "some-buildpack",
							Layers: map[string]occam.ImageBuildpackMetadataLayer{
								"some-layer": {
									SHA:    "some-sha",
									Build:  true,
									Launch: true,
									Cache:  true,
									Metadata: map[string]interface{}{
										"some-key": "some-value",
									},
								},
							},
						},
					},
				}))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"image", "inspect", "some-app:latest",
				}))
			})

			context("failure cases", func() {
				context("when the executable fails", func() {
					it.Before(func() {
						executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
							fmt.Fprintln(execution.Stdout, "[]")
							fmt.Fprintln(execution.Stderr, "Error: No such image: some-app:latest")
							return errors.New("exit status 1")
						}
					})

					it("returns an error", func() {
						_, err := docker.Image.Inspect.Execute("some-app:latest")
						Expect(err).To(MatchError("failed to inspect docker image: exit status 1: Error: No such image: some-app:latest"))
					})
				})

				context("when malformed json is given", func() {
					it.Before(func() {
						executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
							fmt.Fprintln(execution.Stdout, `malformed json %%%%`)
							return nil
						}
					})

					it("return an error", func() {
						_, err := docker.Image.Inspect.Execute("some-app:latest")
						Expect(err).To(MatchError(ContainSubstring("failed to inspect docker image")))
						Expect(err).To(MatchError(ContainSubstring("invalid character")))
					})
				})

				context("when then lifecycle metadata has malformed json", func() {
					it.Before(func() {
						executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
							fmt.Fprintln(execution.Stdout, `[{"Config": {"Labels": {"io.buildpacks.lifecycle.metadata": "%%%"}}}]`)
							return nil
						}
					})

					it("return an error", func() {
						_, err := docker.Image.Inspect.Execute("some-app:latest")
						Expect(err).To(MatchError(ContainSubstring("failed to inspect docker image")))
						Expect(err).To(MatchError(ContainSubstring("invalid character")))
					})
				})
			})
		})

		context("Remove", func() {
			it("removes the image with the given id", func() {
				err := docker.Image.Remove.Execute("some-image-id")
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"image", "remove", "some-image-id",
				}))
			})

			context("failure cases", func() {
				context("when the executable fails", func() {
					it.Before(func() {
						executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
							fmt.Fprintln(execution.Stderr, "Error: No such image: some-image-id")
							return errors.New("exit status 1")
						}
					})

					it("returns an error", func() {
						err := docker.Image.Remove.Execute("some-image-id")
						Expect(err).To(MatchError("failed to remove docker image: exit status 1: Error: No such image: some-image-id"))
					})
				})
			})
		})
	})

	context("Container", func() {
		context("Run", func() {
			var executeArgs [][]string

			it.Before(func() {
				executeArgs = [][]string{}
				executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
					executeArgs = append(executeArgs, execution.Args)

					switch executable.ExecuteCall.CallCount {
					case 1:
						fmt.Fprintln(execution.Stdout, "some-container-id")
					case 2:
						fmt.Fprintln(execution.Stdout, `[
							{
								"Id": "some-container-id",
								"NetworkSettings": {
									"Ports": {
										"8080/tcp": [
											{
												"HostIp": "0.0.0.0",
												"HostPort": "12345"
											}
										]
									}
								}
							}
						]`)
					}

					return nil
				}
			})

			it("runs a docker container with the given image", func() {
				container, err := docker.Container.Run.Execute("some-image-id")
				Expect(err).NotTo(HaveOccurred())
				Expect(container).To(Equal(occam.Container{
					ID: "some-container-id",
					Ports: map[string]string{
						"8080": "12345",
					},
				}))

				Expect(executeArgs).To(HaveLen(2))
				Expect(executeArgs[0]).To(Equal([]string{
					"container", "run",
					"--detach",
					"--env", "PORT=8080",
					"--publish", "8080",
					"--publish-all",
					"some-image-id",
				}))
			})

			context("when given optional environment variables", func() {
				it("sets the --env flags on the run command", func() {
					container, err := docker.Container.Run.
						WithEnv(map[string]string{
							"PORT":     "1234",
							"SOME_VAR": "some-value",
						}).
						Execute("some-image-id")

					Expect(err).NotTo(HaveOccurred())
					Expect(container).To(Equal(occam.Container{
						ID: "some-container-id",
						Ports: map[string]string{
							"8080": "12345",
						},
					}))

					Expect(executeArgs).To(HaveLen(2))
					Expect(executeArgs[0]).To(Equal([]string{
						"container", "run",
						"--detach",
						"--env", "PORT=1234",
						"--env", "SOME_VAR=some-value",
						"--publish", "1234",
						"--publish-all",
						"some-image-id",
					}))
				})
			})

			context("when given optional memory setting", func() {
				it("sets the --memory flag on the run command", func() {
					container, err := docker.Container.Run.
						WithMemory("2GB").
						Execute("some-image-id")

					Expect(err).NotTo(HaveOccurred())
					Expect(container).To(Equal(occam.Container{
						ID: "some-container-id",
						Ports: map[string]string{
							"8080": "12345",
						},
					}))

					Expect(executeArgs).To(HaveLen(2))
					Expect(executeArgs[0]).To(Equal([]string{
						"container", "run",
						"--detach",
						"--env", "PORT=8080",
						"--publish", "8080",
						"--publish-all",
						"--memory", "2GB",
						"some-image-id",
					}))
				})
			})

			context("when given optional command setting", func() {
				it("sets the command field on the run command", func() {
					container, err := docker.Container.Run.
						WithCommand("/some/command").
						Execute("some-image-id")

					Expect(err).NotTo(HaveOccurred())
					Expect(container).To(Equal(occam.Container{
						ID: "some-container-id",
						Ports: map[string]string{
							"8080": "12345",
						},
					}))

					Expect(executeArgs).To(HaveLen(2))
					Expect(executeArgs[0]).To(Equal([]string{
						"container", "run",
						"--detach",
						"--env", "PORT=8080",
						"--publish", "8080",
						"--publish-all",
						"some-image-id",
						"/some/command",
					}))
				})
			})

			context("when given optional tty setting", func() {
				it("sets the tty flag on the run command", func() {
					container, err := docker.Container.Run.
						WithTTY().
						Execute("some-image-id")

					Expect(err).NotTo(HaveOccurred())
					Expect(container).To(Equal(occam.Container{
						ID: "some-container-id",
						Ports: map[string]string{
							"8080": "12345",
						},
					}))

					Expect(executeArgs).To(HaveLen(2))
					Expect(executeArgs[0]).To(Equal([]string{
						"container", "run",
						"--detach",
						"--tty",
						"--env", "PORT=8080",
						"--publish", "8080",
						"--publish-all",
						"some-image-id",
					}))
				})
			})

			context("failure cases", func() {
				context("when the executable fails", func() {
					it.Before(func() {
						executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
							fmt.Fprintln(execution.Stderr, "Unable to find image 'some-image-id' locally")
							return errors.New("exit status 1")
						}
					})

					it("returns an error", func() {
						_, err := docker.Container.Run.Execute("some-image-id")
						Expect(err).To(MatchError("failed to run docker container: exit status 1: Unable to find image 'some-image-id' locally"))
					})
				})
			})
		})

		context("Remove", func() {
			it("removes a docker container with the given container id", func() {
				err := docker.Container.Remove.Execute("some-container-id")
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"container", "rm", "some-container-id", "--force",
				}))
			})

			context("failure cases", func() {
				context("when the executable fails", func() {
					it.Before(func() {
						executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
							fmt.Fprintln(execution.Stderr, "Error: No such container: some-container-id")
							return errors.New("exit status 1")
						}
					})

					it("returns an error", func() {
						err := docker.Container.Remove.Execute("some-container-id")
						Expect(err).To(MatchError("failed to remove docker container: exit status 1: Error: No such container: some-container-id"))
					})
				})
			})
		})

		context("Inspect", func() {
			it.Before(func() {
				executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
					fmt.Fprintln(execution.Stdout, `[
						{
							"Id": "some-container-id",
							"Config": {
								"Env": [
									"PORT=8080"
								]
							},
							"NetworkSettings": {
								"Ports": {
									"8080/tcp": [
										{
											"HostIp": "0.0.0.0",
											"HostPort": "12345"
										}
									]
								}
							}
						}
					]`)

					return nil
				}
			})

			it("inspects the container with the given id", func() {
				container, err := docker.Container.Inspect.Execute("some-container-id")
				Expect(err).NotTo(HaveOccurred())
				Expect(container).To(Equal(occam.Container{
					ID: "some-container-id",
					Env: map[string]string{
						"PORT": "8080",
					},
					Ports: map[string]string{
						"8080": "12345",
					},
				}))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"container", "inspect", "some-container-id",
				}))
			})

			context("failure cases", func() {
				context("when the executable fails", func() {
					it.Before(func() {
						executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
							fmt.Fprintln(execution.Stderr, "Error: No such container: some-container-id")
							return errors.New("exit status 1")
						}
					})

					it("returns an error", func() {
						_, err := docker.Container.Inspect.Execute("some-container-id")
						Expect(err).To(MatchError("failed to inspect docker container: exit status 1: Error: No such container: some-container-id"))
					})
				})

				context("when the output is malformed json", func() {
					it.Before(func() {
						executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
							fmt.Fprintln(execution.Stdout, "%%%")
							return nil
						}
					})

					it("returns an error", func() {
						_, err := docker.Container.Inspect.Execute("some-container-id")
						Expect(err).To(MatchError(ContainSubstring("failed to inspect docker container:")))
						Expect(err).To(MatchError(ContainSubstring("invalid character")))
					})
				})
			})
		})

		context("Logs", func() {
			it.Before(func() {
				executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
					fmt.Fprintln(execution.Stdout, "on stdout")
					fmt.Fprintln(execution.Stderr, "on stderr")

					return nil
				}
			})

			it("fetches the logs for the given container", func() {
				logs, err := docker.Container.Logs.Execute("some-container-id")
				Expect(err).NotTo(HaveOccurred())
				Expect(logs.String()).To(Equal("on stdout\non stderr\n"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"container", "logs", "some-container-id",
				}))
			})

			context("failure cases", func() {
				context("when the executable fails", func() {
					it.Before(func() {
						executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
							fmt.Fprintln(execution.Stderr, "Error: No such container: some-container-id")
							return errors.New("exit status 1")
						}
					})

					it("returns an error", func() {
						_, err := docker.Container.Logs.Execute("some-container-id")
						Expect(err).To(MatchError("failed to fetch docker container logs: exit status 1: Error: No such container: some-container-id"))
					})
				})
			})
		})
	})

	context("Volume", func() {
		context("Remove", func() {
			it("will remove the given volume", func() {
				err := docker.Volume.Remove.Execute([]string{"some-volume-name", "other-volume-name"})
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"volume", "rm", "--force", "some-volume-name", "other-volume-name",
				}))
			})

			context("failure cases", func() {
				context("when the volume rm command fails", func() {
					it.Before(func() {
						executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
							fmt.Fprintln(execution.Stderr, "Error: failed to remove volume")
							return errors.New("exit status 1")
						}
					})

					it("returns an error", func() {
						err := docker.Volume.Remove.Execute([]string{"some-volume-name", "other-volume-name"})
						Expect(err).To(MatchError("failed to remove docker volume: exit status 1: Error: failed to remove volume"))
					})
				})
			})
		})
	})
}
