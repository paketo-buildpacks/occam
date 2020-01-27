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

func testPack(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		executable        *fakes.Executable
		dockerImageClient *fakes.DockerImageClient
		pack              occam.Pack
	)

	it.Before(func() {
		executable = &fakes.Executable{}
		executable.ExecuteCall.Stub = func(execution pexec.Execution) (string, string, error) {
			fmt.Fprintln(execution.Stdout, "some stdout output")
			fmt.Fprintln(execution.Stderr, "some stderr output")
			return "", "", nil
		}

		dockerImageClient = &fakes.DockerImageClient{}
		dockerImageClient.InspectCall.Returns.Image = occam.Image{
			ID: "some-image-id",
		}

		pack = occam.NewPack().WithExecutable(executable).WithDockerImageClient(dockerImageClient)
	})

	context("WithVerbose", func() {
		it.Before(func() {
			pack = pack.WithVerbose()
		})

		it("includes the --verbose option on all commands", func() {
			image, logs, err := pack.Build.Execute("myapp", "/some/app/path")
			Expect(err).NotTo(HaveOccurred())
			Expect(image).To(Equal(occam.Image{
				ID: "some-image-id",
			}))
			Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
				"build", "myapp",
				"--verbose",
				"--path", "/some/app/path",
			}))
			Expect(dockerImageClient.InspectCall.Receives.Ref).To(Equal("myapp"))
		})
	})

	context("WithNoColor", func() {
		it.Before(func() {
			pack = pack.WithNoColor()
		})

		it("includes the --no-color option on all commands", func() {
			image, logs, err := pack.Build.Execute("myapp", "/some/app/path")
			Expect(err).NotTo(HaveOccurred())
			Expect(image).To(Equal(occam.Image{
				ID: "some-image-id",
			}))
			Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
				"build", "myapp",
				"--no-color",
				"--path", "/some/app/path",
			}))
			Expect(dockerImageClient.InspectCall.Receives.Ref).To(Equal("myapp"))
		})
	})

	context("Build", func() {
		it("returns an image with the given name and the build logs", func() {
			image, logs, err := pack.Build.Execute("myapp", "/some/app/path")
			Expect(err).NotTo(HaveOccurred())
			Expect(image).To(Equal(occam.Image{
				ID: "some-image-id",
			}))
			Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
				"build", "myapp", "--path", "/some/app/path",
			}))
			Expect(dockerImageClient.InspectCall.Receives.Ref).To(Equal("myapp"))
		})

		context("when given optional buildpacks", func() {
			it("returns an image with the given name and the build logs", func() {
				image, logs, err := pack.Build.
					WithBuildpacks("some-buildpack", "other-buildpack").
					Execute("myapp", "/some/app/path")

				Expect(err).NotTo(HaveOccurred())
				Expect(image).To(Equal(occam.Image{
					ID: "some-image-id",
				}))
				Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"build", "myapp",
					"--path", "/some/app/path",
					"--buildpack", "some-buildpack",
					"--buildpack", "other-buildpack",
				}))
				Expect(dockerImageClient.InspectCall.Receives.Ref).To(Equal("myapp"))
			})
		})

		context("when given optional network name", func() {
			it("returns an image with the given name and the build logs", func() {
				image, logs, err := pack.Build.WithNetwork("some-network").Execute("myapp", "/some/app/path")

				Expect(err).NotTo(HaveOccurred())
				Expect(image).To(Equal(occam.Image{
					ID: "some-image-id",
				}))
				Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"build", "myapp",
					"--path", "/some/app/path",
					"--network", "some-network",
				}))
				Expect(dockerImageClient.InspectCall.Receives.Ref).To(Equal("myapp"))
			})
		})

		context("when given optional builder name", func() {
			it("returns an image with the given name and the build logs", func() {
				image, logs, err := pack.Build.WithBuilder("some-builder").Execute("myapp", "/some/app/path")

				Expect(err).NotTo(HaveOccurred())
				Expect(image).To(Equal(occam.Image{
					ID: "some-image-id",
				}))
				Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"build", "myapp",
					"--path", "/some/app/path",
					"--builder", "some-builder",
				}))
				Expect(dockerImageClient.InspectCall.Receives.Ref).To(Equal("myapp"))
			})
		})

		context("when given optional clear-cache", func() {
			it("returns an image with the given name and the build logs", func() {
				image, logs, err := pack.Build.WithClearCache().Execute("myapp", "/some/app/path")

				Expect(err).NotTo(HaveOccurred())
				Expect(image).To(Equal(occam.Image{
					ID: "some-image-id",
				}))
				Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"build", "myapp",
					"--path", "/some/app/path",
					"--clear-cache",
				}))
				Expect(dockerImageClient.InspectCall.Receives.Ref).To(Equal("myapp"))
			})
		})

		context("when given optional env", func() {
			it("returns an image with the given name and the build logs", func() {
				image, logs, err := pack.Build.
					WithEnv(map[string]string{
						"SOME_KEY":  "some-value",
						"OTHER_KEY": "other-value",
					}).
					Execute("myapp", "/some/app/path")

				Expect(err).NotTo(HaveOccurred())
				Expect(image).To(Equal(occam.Image{
					ID: "some-image-id",
				}))
				Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"build", "myapp",
					"--path", "/some/app/path",
					"--env", "OTHER_KEY=other-value",
					"--env", "SOME_KEY=some-value",
				}))
				Expect(dockerImageClient.InspectCall.Receives.Ref).To(Equal("myapp"))
			})
		})

		context("when given optional no-pull", func() {
			it("returns an image with the given name and the build logs", func() {
				image, logs, err := pack.Build.WithNoPull().Execute("myapp", "/some/app/path")

				Expect(err).NotTo(HaveOccurred())
				Expect(image).To(Equal(occam.Image{
					ID: "some-image-id",
				}))
				Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"build", "myapp",
					"--path", "/some/app/path",
					"--no-pull",
				}))
				Expect(dockerImageClient.InspectCall.Receives.Ref).To(Equal("myapp"))
			})
		})

		context("failure cases", func() {
			context("when the executable fails", func() {
				it.Before(func() {
					executable.ExecuteCall.Stub = func(execution pexec.Execution) (string, string, error) {
						fmt.Fprintln(execution.Stdout, "some stdout output")
						fmt.Fprintln(execution.Stderr, "some stderr output")
						return "", "", errors.New("failed to execute")
					}
				})

				it("returns an error and the build logs", func() {
					_, logs, err := pack.Build.Execute("myapp", "/some/app/path")
					Expect(err).To(MatchError("failed to pack build: failed to execute"))
					Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))
				})
			})

			context("when the docker image client fails", func() {
				it.Before(func() {
					dockerImageClient.InspectCall.Returns.Error = errors.New("failed to inspect image")
				})

				it("returns an error and the build logs", func() {
					_, logs, err := pack.Build.Execute("myapp", "/some/app/path")
					Expect(err).To(MatchError("failed to pack build: failed to inspect image"))
					Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))
				})
			})
		})
	})
}
