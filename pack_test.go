package occam_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/paketo-buildpacks/occam/fakes"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPack(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		executable               *fakes.Executable
		dockerImageInspectClient *fakes.DockerImageInspectClient
		pack                     occam.Pack
	)

	it.Before(func() {
		executable = &fakes.Executable{}
		executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
			fmt.Fprintln(execution.Stdout, "some stdout output")
			fmt.Fprintln(execution.Stderr, "some stderr output")
			return nil
		}

		dockerImageInspectClient = &fakes.DockerImageInspectClient{}
		dockerImageInspectClient.ExecuteCall.Returns.Image = occam.Image{
			ID: "some-image-id",
		}

		pack = occam.NewPack().WithExecutable(executable).WithDockerImageInspectClient(dockerImageInspectClient)
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
			Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
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
			Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
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
			Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
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
				Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
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
				Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
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
				Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
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
				Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
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
				Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
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
				Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
			})
		})

		context("when given optional pull-policy", func() {
			it("returns an image with the given name and the build logs", func() {
				image, logs, err := pack.Build.WithPullPolicy("if-not-present").Execute("myapp", "/some/app/path")

				Expect(err).NotTo(HaveOccurred())
				Expect(image).To(Equal(occam.Image{
					ID: "some-image-id",
				}))
				Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"build", "myapp",
					"--path", "/some/app/path",
					"--pull-policy", "if-not-present",
				}))
				Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
			})
		})

		context("when given optional sbom-output-dir", func() {
			it("returns an image with the given name and the build logs", func() {
				image, logs, err := pack.Build.WithSBOMOutputDir("some-dir").Execute("myapp", "/some/app/path")

				Expect(err).NotTo(HaveOccurred())
				Expect(image).To(Equal(occam.Image{
					ID: "some-image-id",
				}))
				Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"build", "myapp",
					"--path", "/some/app/path",
					"--sbom-output-dir", "some-dir",
				}))
				Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
			})
		})

		context("when given optional trust-builder", func() {
			it("returns an image with the given name and the build logs", func() {
				image, logs, err := pack.Build.WithTrustBuilder().Execute("myapp", "/some/app/path")

				Expect(err).NotTo(HaveOccurred())
				Expect(image).To(Equal(occam.Image{
					ID: "some-image-id",
				}))
				Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"build", "myapp",
					"--path", "/some/app/path",
					"--trust-builder",
				}))
				Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
			})
		})

		context("when given optional volumes", func() {
			it("includes the --volume option and args on all commands", func() {
				image, logs, err := pack.Build.
					WithVolumes(
						"/tmp/host-source:/tmp/dir-on-image:rw",
						"/tmp/second-host-source:/tmp/second-dir-on-image:ro",
					).
					Execute("myapp", "/some/app/path")

				Expect(err).NotTo(HaveOccurred())
				Expect(image).To(Equal(occam.Image{
					ID: "some-image-id",
				}))
				Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"build", "myapp",
					"--path", "/some/app/path",
					"--volume", "/tmp/host-source:/tmp/dir-on-image:rw",
					"--volume", "/tmp/second-host-source:/tmp/second-dir-on-image:ro",
				}))
				Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
			})
		})

		context("when given optional gid", func() {
			it("includes the --gid option and given argument on all commands", func() {
				image, logs, err := pack.Build.
					WithGID(
						"1001",
					).
					Execute("myapp", "/some/app/path")

				Expect(err).NotTo(HaveOccurred())
				Expect(image).To(Equal(occam.Image{
					ID: "some-image-id",
				}))
				Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"build", "myapp",
					"--path", "/some/app/path",
					"--gid", "1001",
				}))
				Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
			})
		})

		context("failure cases", func() {
			context("when the executable fails", func() {
				it.Before(func() {
					executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
						fmt.Fprintln(execution.Stdout, "some stdout output")
						fmt.Fprintln(execution.Stderr, "some stderr output")
						return errors.New("failed to execute")
					}
				})

				it("returns an error and the build logs", func() {
					_, logs, err := pack.Build.Execute("myapp", "/some/app/path")
					Expect(err).To(MatchError("failed to pack build: failed to execute\n\nOutput:\nsome stdout output\nsome stderr output\n"))
					Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))
				})
			})

			context("when the docker image client fails", func() {
				it.Before(func() {
					dockerImageInspectClient.ExecuteCall.Returns.Error = errors.New("failed to inspect image")
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
