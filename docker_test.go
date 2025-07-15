package occam_test

import (
	"bytes"
	ctx "context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	name "github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/paketo-buildpacks/occam"
	"github.com/paketo-buildpacks/occam/fakes"
	"github.com/paketo-buildpacks/packit/v2/pexec"
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
									"io.buildpacks.lifecycle.metadata": "{\"buildpacks\": [{\"key\": \"some-buildpack\", \"layers\": {\"some-layer\": {\"sha\": \"some-sha\", \"build\": true, \"launch\": true, \"cache\": true, \"data\": {\"some-key\": \"some-value\"}}}}]}",
									"some-other-label": "some-value"
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
					Labels: map[string]string{
						"io.buildpacks.lifecycle.metadata": "{\"buildpacks\": [{\"key\": \"some-buildpack\", \"layers\": {\"some-layer\": {\"sha\": \"some-sha\", \"build\": true, \"launch\": true, \"cache\": true, \"data\": {\"some-key\": \"some-value\"}}}}]}",
						"some-other-label":                 "some-value",
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

			context("when given the optional force setting", func() {
				it("sets the  force flag on the remove command", func() {
					err := docker.Image.Remove.WithForce().Execute("some-image-id")
					Expect(err).NotTo(HaveOccurred())

					Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
						"image", "remove", "some-image-id", "--force",
					}))
				})
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

		context("Tag", func() {
			it("Tags the image with the target name", func() {
				err := docker.Image.Tag.Execute("some-image-id", "new-image-id")
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"image", "tag", "some-image-id", "new-image-id",
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
						err := docker.Image.Tag.Execute("some-image-id", "some-other-id")
						Expect(err).To(MatchError("failed to tag docker image: exit status 1: Error: No such image: some-image-id"))
					})
				})
			})
		})

		context("ExportToOCI", func() {
			it("Exports Docker image as v1.Image", func() {
				fakeImg, err := random.Image(1, 5)
				Expect(err).NotTo(HaveOccurred())

				fakeImgDigest, err := fakeImg.Digest()
				Expect(err).NotTo(HaveOccurred())

				mockClient := &fakes.DockerDaemonClient{}
				mockClient.ImageInspectWithRawCall.Stub = func(_ ctx.Context, s string) (image.InspectResponse, []byte, error) {
					return image.InspectResponse{
						ID: fakeImgDigest.String(),
					}, nil, nil
				}
				mockClient.ImageSaveCall.Stub = func(ctx ctx.Context, s []string, iso ...client.ImageSaveOption) (io.ReadCloser, error) {
					buf := bytes.NewBuffer(nil)
					ref, _ := name.ParseReference("some-image-id")
					err = tarball.Write(ref, fakeImg, buf)
					Expect(err).NotTo(HaveOccurred())
					return io.NopCloser(buf), nil
				}
				img, err := docker.Image.ExportToOCI.WithClient(mockClient).Execute("some-image-id")
				Expect(err).NotTo(HaveOccurred())
				Expect(img).NotTo(BeNil())
				digest, err := img.Digest()
				Expect(err).NotTo(HaveOccurred())
				Expect(digest.String()).To(Equal(fakeImgDigest.String()))

				layers, err := img.Layers()
				Expect(err).NotTo(HaveOccurred())
				Expect(layers).To(HaveLen(5))
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
								"Id": "some-container-id"
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
				}))

				Expect(executeArgs).To(HaveLen(2))
				Expect(executeArgs[0]).To(Equal([]string{
					"container", "run",
					"--detach",
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
					}))

					Expect(executeArgs).To(HaveLen(2))
					Expect(executeArgs[0]).To(Equal([]string{
						"container", "run",
						"--detach",
						"--env", "PORT=1234",
						"--env", "SOME_VAR=some-value",
						"some-image-id",
					}))
				})
			})

			context("port publishing", func() {
				it.Before(func() {
					executeArgs = [][]string{}
					executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
						executeArgs = append(executeArgs, execution.Args)
						fmt.Fprintln(execution.Stdout, `[
							{
								"Id": "some-container-id",
								"NetworkSettings": {
									"Ports": {
										"3000/tcp": [
											{
												"HostIp": "0.0.0.0",
												"HostPort": "54321"
											}
										]
									}
								}
							}
						]`)
						return nil
					}
				})

				context("when given optional with publish port setting", func() {
					it("sets the published port mapping on the run command", func() {
						container, err := docker.Container.Run.
							WithPublish("3000").
							Execute("some-image-id")

						Expect(err).NotTo(HaveOccurred())
						Expect(container).To(Equal(occam.Container{
							ID: "some-container-id",
							Ports: map[string]string{
								"3000": "54321",
							},
						}))

						Expect(executeArgs).To(HaveLen(2))
						Expect(executeArgs[0]).To(Equal([]string{
							"container", "run",
							"--detach",
							"--publish", "3000",
							"some-image-id",
						}))
					})
				})

				context("when given optional publish all port setting", func() {
					it("sets the --publish-all flag on the run command", func() {
						container, err := docker.Container.Run.
							WithPublishAll().
							Execute("some-image-id")

						Expect(err).NotTo(HaveOccurred())
						Expect(container).To(Equal(occam.Container{
							ID: "some-container-id",
							Ports: map[string]string{
								"3000": "54321",
							},
						}))

						Expect(executeArgs).To(HaveLen(2))
						Expect(executeArgs[0]).To(Equal([]string{
							"container", "run",
							"--detach",
							"--publish-all",
							"some-image-id",
						}))
					})
				})

				context("when given multiple optional ports to publish", func() {
					it.Before(func() {
						executeArgs = [][]string{}
						executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
							executeArgs = append(executeArgs, execution.Args)
							fmt.Fprintln(execution.Stdout, `[
							{
								"Id": "some-container-id",
								"NetworkSettings": {
									"Ports": {
										"3000/tcp": [
											{
												"HostIp": "0.0.0.0",
												"HostPort": "34321"
											}
										],
										"4000/tcp": [
											{
												"HostIp": "0.0.0.0",
												"HostPort": "44321"
											}
										],
										"5000/tcp": [
											{
												"HostIp": "0.0.0.0",
												"HostPort": "54321"
											}
										]
									}
								}
							}
						]`)
							return nil
						}
					})

					it("sets all of the published ports in the run command", func() {
						container, err := docker.Container.Run.
							WithPublish("3000").
							WithPublish("4000").
							WithPublish("5000").
							WithPublishAll().
							Execute("some-image-id")

						Expect(err).NotTo(HaveOccurred())
						Expect(container).To(Equal(occam.Container{
							ID: "some-container-id",
							Ports: map[string]string{
								"3000": "34321",
								"4000": "44321",
								"5000": "54321",
							},
						}))

						Expect(executeArgs).To(HaveLen(2))
						Expect(executeArgs[0]).To(Equal([]string{
							"container", "run",
							"--detach",
							"--publish", "3000",
							"--publish", "4000",
							"--publish", "5000",
							"--publish-all",
							"some-image-id",
						}))
					})
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
					}))

					Expect(executeArgs).To(HaveLen(2))
					Expect(executeArgs[0]).To(Equal([]string{
						"container", "run",
						"--detach",
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
					}))

					Expect(executeArgs).To(HaveLen(2))
					Expect(executeArgs[0]).To(Equal([]string{
						"container", "run",
						"--detach",
						"some-image-id",
						"/some/command",
					}))
				})
			})

			context("when given optional command args", func() {
				it("sets the command args on the run command", func() {
					container, err := docker.Container.Run.
						WithCommand("/some/command").
						WithCommandArgs([]string{"arg1", "arg2"}).
						Execute("some-image-id")

					Expect(err).NotTo(HaveOccurred())
					Expect(container).To(Equal(occam.Container{
						ID: "some-container-id",
					}))

					Expect(executeArgs).To(HaveLen(2))
					Expect(executeArgs[0]).To(Equal([]string{
						"container", "run",
						"--detach",
						"some-image-id",
						"/some/command",
						"arg1",
						"arg2",
					}))
				})
			})

			context("when given optional direct setting", func() {
				it("runs the command directly (i.e. with '--' before command)", func() {
					container, err := docker.Container.Run.
						WithCommand("/some/command").
						WithDirect().
						Execute("some-image-id")

					Expect(err).NotTo(HaveOccurred())
					Expect(container).To(Equal(occam.Container{
						ID: "some-container-id",
					}))

					Expect(executeArgs).To(HaveLen(2))
					Expect(executeArgs[0]).To(Equal([]string{
						"container", "run",
						"--detach",
						"some-image-id",
						"--",
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
					}))

					Expect(executeArgs).To(HaveLen(2))
					Expect(executeArgs[0]).To(Equal([]string{
						"container", "run",
						"--detach",
						"--tty",
						"some-image-id",
					}))
				})
			})

			context("when given optional entrypoint setting", func() {
				it("sets the entrypoint flag on the run command", func() {
					container, err := docker.Container.Run.
						WithEntrypoint("launcher").
						Execute("some-image-id")

					Expect(err).NotTo(HaveOccurred())
					Expect(container).To(Equal(occam.Container{
						ID: "some-container-id",
					}))

					Expect(executeArgs).To(HaveLen(2))
					Expect(executeArgs[0]).To(Equal([]string{
						"container", "run",
						"--detach",
						"--entrypoint", "launcher",
						"some-image-id",
					}))
				})
			})

			context("when given optional entrypoint setting", func() {
				it("sets the entrypoint flag on the run command", func() {
					container, err := docker.Container.Run.
						WithNetwork("host").
						Execute("some-image-id")

					Expect(err).NotTo(HaveOccurred())
					Expect(container).To(Equal(occam.Container{
						ID: "some-container-id",
					}))

					Expect(executeArgs).To(HaveLen(2))
					Expect(executeArgs[0]).To(Equal([]string{
						"container", "run",
						"--detach",
						"--network", "host",
						"some-image-id",
					}))
				})
			})

			// TODO: remove this when WithVolume is deprecated.
			context("when given optionial volume setting", func() {
				it("sets the volume flag on the run command", func() {
					container, err := docker.Container.Run.
						WithVolume("/tmp/host-source:/tmp/dir-on-container:rw").
						Execute("some-image-id")

					Expect(err).NotTo(HaveOccurred())
					Expect(container).To(Equal(occam.Container{
						ID: "some-container-id",
					}))

					Expect(executeArgs).To(HaveLen(2))
					Expect(executeArgs[0]).To(Equal([]string{
						"container", "run",
						"--detach",
						"--volume", "/tmp/host-source:/tmp/dir-on-container:rw",
						"some-image-id",
					}))
				})
			})

			context("when given optional volumes setting", func() {
				it("sets the volume flags on the run command", func() {
					container, err := docker.Container.Run.
						WithVolumes(
							"/tmp/host-source:/tmp/dir-on-container:rw",
							"/tmp/second-host-source:/tmp/second-dir-on-container:ro",
						).
						Execute("some-image-id")

					Expect(err).NotTo(HaveOccurred())
					Expect(container).To(Equal(occam.Container{
						ID: "some-container-id",
					}))

					Expect(executeArgs).To(HaveLen(2))
					Expect(executeArgs[0]).To(Equal([]string{
						"container", "run",
						"--detach",
						"--volume", "/tmp/host-source:/tmp/dir-on-container:rw",
						"--volume", "/tmp/second-host-source:/tmp/second-dir-on-container:ro",
						"some-image-id",
					}))
				})
			})

			context("when given optional read-only setting", func() {
				it("sets the read-only flag on the run command", func() {
					container, err := docker.Container.Run.
						WithReadOnly().
						Execute("some-image-id")

					Expect(err).NotTo(HaveOccurred())
					Expect(container).To(Equal(occam.Container{
						ID: "some-container-id",
					}))

					Expect(executeArgs).To(HaveLen(2))
					Expect(executeArgs[0]).To(Equal([]string{
						"container", "run",
						"--detach",
						"--read-only",
						"some-image-id",
					}))
				})
			})

			context("when given optional mount setting", func() {
				it("sets the mount flag on the run command", func() {
					container, err := docker.Container.Run.
						WithMounts(
							"type=tmpfs,destination=/tmp",
							"type=bind,source=/my-local,destination=/local",
						).
						Execute("some-image-id")

					Expect(err).NotTo(HaveOccurred())
					Expect(container).To(Equal(occam.Container{
						ID: "some-container-id",
					}))

					Expect(executeArgs).To(HaveLen(2))
					Expect(executeArgs[0]).To(Equal([]string{
						"container", "run",
						"--detach",
						"--mount", "type=tmpfs,destination=/tmp",
						"--mount", "type=bind,source=/my-local,destination=/local",
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

		context("Restart", func() {
			it("restarts a docker container with the given container id", func() {
				err := docker.Container.Restart.Execute("some-container-id")
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"container", "restart", "some-container-id",
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
						err := docker.Container.Restart.Execute("some-container-id")
						Expect(err).To(MatchError("failed to restart docker container: exit status 1: Error: No such container: some-container-id"))
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
							  "Networks": {
                  "bridge": {
									  "IPAddress": "10.172.0.2"
									}
								},
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
					IPAddresses: map[string]string{
						"bridge": "10.172.0.2",
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

		context("Stop", func() {
			it("stops the given container", func() {
				err := docker.Container.Stop.Execute("some-container-id")
				Expect(err).NotTo(HaveOccurred())
				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"container", "stop", "some-container-id",
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
						err := docker.Container.Stop.Execute("some-container-id")
						Expect(err).To(MatchError("failed to stop docker container: exit status 1: Error: No such container: some-container-id"))
					})
				})
			})
		})

		context("Copy", func() {
			it("will execute 'docker container cp SOURCE DEST'", func() {
				err := docker.Container.Copy.Execute("source/path", "dest-container:/path")
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"container",
					"cp",
					"source/path",
					"dest-container:/path",
				}))
			})

			context("failure cases", func() {
				context("when the cp command fails", func() {
					it.Before(func() {
						executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
							_, err := fmt.Fprint(execution.Stderr, "must specify at least one container source")
							Expect(err).NotTo(HaveOccurred())
							return errors.New("exit status 1")
						}
					})

					it("returns an error", func() {
						err := docker.Container.Copy.Execute("source", "dest")
						Expect(err).To(MatchError("'docker cp' failed: exit status 1: must specify at least one container source"))
					})
				})
			})
		})

		context("Exec", func() {
			context("Execute", func() {
				it("will execute 'docker container exec CONTAINER CMD'", func() {
					err := docker.Container.Exec.Execute("abc123", "/bin/bash", "-c", "echo hi")
					Expect(err).NotTo(HaveOccurred())

					Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
						"container", "exec",
						"abc123",
						"/bin/bash", "-c", "echo hi",
					}))
				})

				context("WithStdin", func() {
					it("passes the reader as stdin to the underlying docker exec call", func() {
						err := docker.Container.Exec.
							WithStdin(strings.NewReader("goodbye moon\nhello world")).
							Execute("abc123", "grep", "hello")
						Expect(err).NotTo(HaveOccurred())

						Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
							"container", "exec",
							"abc123",
							"grep", "hello",
						}))

						content, err := io.ReadAll(executable.ExecuteCall.Receives.Execution.Stdin)
						Expect(err).NotTo(HaveOccurred())
						Expect(string(content)).To(Equal("goodbye moon\nhello world"))
					})
				})

				context("WithUser", func() {
					it("sets the --user flag", func() {
						err := docker.Container.Exec.
							WithUser("some-user:some-group").
							Execute("abc123", "echo", "hello")
						Expect(err).NotTo(HaveOccurred())

						Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
							"container", "exec",
							"--user", "some-user:some-group",
							"abc123",
							"echo", "hello",
						}))
					})
				})

				context("WithInteractive", func() {
					it("sets the --interactive flag", func() {
						err := docker.Container.Exec.
							WithInteractive().
							Execute("abc123", "echo", "hello")
						Expect(err).NotTo(HaveOccurred())

						Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
							"container", "exec",
							"--interactive",
							"abc123",
							"echo", "hello",
						}))
					})
				})

				context("failure cases", func() {
					context("when the exec command fails", func() {
						it.Before(func() {
							executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
								_, err := fmt.Fprint(execution.Stderr, "error in exec command")
								Expect(err).NotTo(HaveOccurred())
								return errors.New("exit status 99")
							}
						})

						it("returns an error", func() {
							err := docker.Container.Exec.Execute("container", "arg0", "arg1")
							Expect(err).To(MatchError("'docker exec' failed: exit status 99: error in exec command"))
						})
					})
				})
			})

			context("ExecuteBash", func() {
				it("will execute 'docker container exec CONTAINER /bin/bash -c CMD'", func() {
					err := docker.Container.Exec.ExecuteBash("abc123", "echo hi")
					Expect(err).NotTo(HaveOccurred())

					Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
						"container",
						"exec",
						"abc123",
						"/bin/bash",
						"-c",
						"echo hi",
					}))
				})

				context("failure cases", func() {
					context("when the exec command fails", func() {
						it.Before(func() {
							executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
								_, err := fmt.Fprint(execution.Stderr, "error in exec command")
								Expect(err).NotTo(HaveOccurred())
								return errors.New("exit status 88")
							}
						})

						it("returns an error", func() {
							err := docker.Container.Exec.ExecuteBash("container", "script")
							Expect(err).To(MatchError("'docker exec' failed: exit status 88: error in exec command"))
						})
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

	context("Pull", func() {
		it("will pull the given image", func() {
			err := docker.Pull.Execute("some-image")
			Expect(err).NotTo(HaveOccurred())

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
				"pull", "some-image",
			}))
		})

		context("failure cases", func() {
			context("when the pull command fails", func() {
				it.Before(func() {
					executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
						fmt.Fprintln(execution.Stderr, "Error: failed to pull image")
						return errors.New("exit status 1")
					}
				})

				it("returns an error", func() {
					err := docker.Pull.Execute("some-image")
					Expect(err).To(MatchError("failed to pull docker image: exit status 1: Error: failed to pull image"))
				})
			})

		})
	})
}
