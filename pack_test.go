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
				"--cache",
				"type=build;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.build",
				"--cache",
				"type=launch;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.launch",
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
				"--cache",
				"type=build;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.build",
				"--cache",
				"type=launch;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.launch",
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
				"--cache",
				"type=build;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.build",
				"--cache",
				"type=launch;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.launch",
			}))
			Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
		})

		it("sets PACK_VOLUME_KEY", func() {
			_, _, err := pack.Build.Execute("myapp", "/some/app/path")
			Expect(err).NotTo(HaveOccurred())

			Expect(executable.ExecuteCall.Receives.Execution.Env).To(ContainElement("PACK_VOLUME_KEY=myapp-volume"))
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
					"--cache",
					"type=build;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.build",
					"--cache",
					"type=launch;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.launch",
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
					"--cache",
					"type=build;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.build",
					"--cache",
					"type=launch;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.launch",
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
					"--cache",
					"type=build;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.build",
					"--cache",
					"type=launch;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.launch",
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
					"--cache",
					"type=build;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.build",
					"--cache",
					"type=launch;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.launch",
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
					"--cache",
					"type=build;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.build",
					"--cache",
					"type=launch;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.launch",
				}))
				Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
			})
		})

		context("when given additional env", func() {
			it("merges additional environment variables with existing ones", func() {
				image, logs, err := pack.Build.
					WithEnv(map[string]string{
						"EXISTING_KEY": "existing-value",
						"SHARED_KEY":   "original-value",
					}).
					WithAdditionalEnv(map[string]string{
						"ADDITIONAL_KEY": "additional-value",
						"SHARED_KEY":     "overridden-value",
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
					"--env", "ADDITIONAL_KEY=additional-value",
					"--env", "EXISTING_KEY=existing-value",
					"--env", "SHARED_KEY=overridden-value",
					"--cache",
					"type=build;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.build",
					"--cache",
					"type=launch;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.launch",
				}))
				Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
			})

			it("adds environment variables when no existing env is set", func() {
				image, logs, err := pack.Build.
					WithAdditionalEnv(map[string]string{
						"NEW_KEY": "new-value",
						"ANOTHER_KEY": "another-value",
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
					"--env", "ANOTHER_KEY=another-value",
					"--env", "NEW_KEY=new-value",
					"--cache",
					"type=build;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.build",
					"--cache",
					"type=launch;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.launch",
				}))
				Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
			})

			it("handles empty additional env map", func() {
				image, logs, err := pack.Build.
					WithEnv(map[string]string{
						"EXISTING_KEY": "existing-value",
					}).
					WithAdditionalEnv(map[string]string{}).
					Execute("myapp", "/some/app/path")

				Expect(err).NotTo(HaveOccurred())
				Expect(image).To(Equal(occam.Image{
					ID: "some-image-id",
				}))
				Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"build", "myapp",
					"--path", "/some/app/path",
					"--env", "EXISTING_KEY=existing-value",
					"--cache",
					"type=build;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.build",
					"--cache",
					"type=launch;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.launch",
				}))
				Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
			})

			it("handles nil additional env map", func() {
				image, logs, err := pack.Build.
					WithEnv(map[string]string{
						"EXISTING_KEY": "existing-value",
					}).
					WithAdditionalEnv(nil).
					Execute("myapp", "/some/app/path")

				Expect(err).NotTo(HaveOccurred())
				Expect(image).To(Equal(occam.Image{
					ID: "some-image-id",
				}))
				Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"build", "myapp",
					"--path", "/some/app/path",
					"--env", "EXISTING_KEY=existing-value",
					"--cache",
					"type=build;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.build",
					"--cache",
					"type=launch;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.launch",
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
					"--cache",
					"type=build;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.build",
					"--cache",
					"type=launch;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.launch",
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
					"--cache",
					"type=build;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.build",
					"--cache",
					"type=launch;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.launch",
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
					"--cache",
					"type=build;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.build",
					"--cache",
					"type=launch;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.launch",
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
					"--cache",
					"type=build;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.build",
					"--cache",
					"type=launch;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.launch",
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
					"--cache",
					"type=build;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.build",
					"--cache",
					"type=launch;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.launch",
				}))
				Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
			})
		})

		context("when given optional run image", func() {
			it("includes the --run-image option", func() {
				image, logs, err := pack.Build.
					WithRunImage("custom").
					Execute("myapp", "/some/app/path")

				Expect(err).NotTo(HaveOccurred())
				Expect(image).To(Equal(occam.Image{
					ID: "some-image-id",
				}))
				Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"build", "myapp",
					"--path", "/some/app/path",
					"--run-image", "custom",
					"--cache",
					"type=build;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.build",
					"--cache",
					"type=launch;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.launch",
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
					"--cache",
					"type=build;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.build",
					"--cache",
					"type=launch;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.launch",
				}))
				Expect(dockerImageInspectClient.ExecuteCall.Receives.Ref).To(Equal("myapp"))
			})
		})

		context("when given additional build args", func() {
			it("includes the additional args", func() {
				image, logs, err := pack.Build.
					WithAdditionalBuildArgs("--not-supported-yet", "true").
					Execute("myapp", "/some/app/path")

				Expect(err).NotTo(HaveOccurred())
				Expect(image).To(Equal(occam.Image{
					ID: "some-image-id",
				}))
				Expect(logs.String()).To(Equal("some stdout output\nsome stderr output\n"))

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"build", "myapp",
					"--path", "/some/app/path",
					"--cache",
					"type=build;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.build",
					"--cache",
					"type=launch;format=volume;name=pack-cache-myapp_latest-c48abba4d0f8.launch",
					"--not-supported-yet", "true",
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

	context("Builder", func() {
		context("Inspect", func() {
			context("when given no builder image name", func() {
				it.Before(func() {
					executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
						fmt.Fprintln(execution.Stdout, `{
							"builder_name": "default-builder",
							"trusted": true,
							"default": true,
							"local_info": {
								"description": "some-builder-description",
								"created_by": {
									"name": "some-creator-name",
									"version": "some-creator-version"
								},
								"stack": {
									"id": "some-stack-id"
								},
								"lifecycle": {
									"version": "some-lifecycle-version",
									"buildpack_apis": {
										"deprecated": [
											"some-deprecated-api-version"
										],
										"supported": [
											"some-api-version",
											"other-api-version"
										]
									},
									"platform_apis": {
										"deprecated": [
											"some-deprecated-api-version"
										],
										"supported": [
											"some-api-version",
											"other-api-version"
										]
									}
								},
								"run_images": [
									{
										"name": "some-run-image"
									}
								],
								"buildpacks": [
									{
										"id": "some-buildpack-id",
										"name": "some-buildpack-name",
										"version": "some-buildpack-version",
										"homepage": "some-buildpack-homepage"
									},
									{
										"id": "other-buildpack-id",
										"name": "other-buildpack-name",
										"version": "other-buildpack-version",
										"homepage": "other-buildpack-homepage"
									}
								],
								"detection_order": [
									{
										"buildpacks": [
											{
												"id": "some-buildpack-id",
												"version": "some-buildpack-version",
												"buildpacks": [
													{
														"id": "other-buildpack-id",
														"version": "other-buildpack-version",
														"optional": true
													}
												]
											}
										]
									}
								]
							},
							"remote_info": {
								"description": "some-builder-description",
								"created_by": {
									"name": "some-creator-name",
									"version": "some-creator-version"
								},
								"stack": {
									"id": "some-stack-id"
								},
								"lifecycle": {
									"version": "some-lifecycle-version",
									"buildpack_apis": {
										"deprecated": [
											"some-deprecated-api-version"
										],
										"supported": [
											"some-api-version",
											"other-api-version"
										]
									},
									"platform_apis": {
										"deprecated": [
											"some-deprecated-api-version"
										],
										"supported": [
											"some-api-version",
											"other-api-version"
										]
									}
								},
								"run_images": [
									{
										"name": "some-run-image"
									}
								],
								"buildpacks": [
									{
										"id": "some-buildpack-id",
										"name": "some-buildpack-name",
										"version": "some-buildpack-version",
										"homepage": "some-buildpack-homepage"
									},
									{
										"id": "other-buildpack-id",
										"name": "other-buildpack-name",
										"version": "other-buildpack-version",
										"homepage": "other-buildpack-homepage"
									}
								],
								"detection_order": [
									{
										"buildpacks": [
											{
												"id": "some-buildpack-id",
												"version": "some-buildpack-version",
												"buildpacks": [
													{
														"id": "other-buildpack-id",
														"version": "other-buildpack-version",
														"optional": true
													}
												]
											}
										]
									}
								]
							}
						}`)
						return nil
					}
				})

				it("returns the default builder", func() {
					builder, err := pack.Builder.Inspect.Execute()
					Expect(err).NotTo(HaveOccurred())
					Expect(builder).To(Equal(occam.Builder{
						BuilderName: "default-builder",
						Trusted:     true,
						Default:     true,
						LocalInfo: occam.BuilderInfo{
							Description: "some-builder-description",
							CreatedBy: occam.BuilderInfoCreatedBy{
								Name:    "some-creator-name",
								Version: "some-creator-version",
							},
							Stack: occam.BuilderInfoStack{
								ID: "some-stack-id",
							},
							Lifecycle: occam.BuilderInfoLifecycle{
								Version: "some-lifecycle-version",
								BuildpackAPIs: occam.BuilderInfoLifecycleAPIs{
									Deprecated: []string{
										"some-deprecated-api-version",
									},
									Supported: []string{
										"some-api-version",
										"other-api-version",
									},
								},
								PlatformAPIs: occam.BuilderInfoLifecycleAPIs{
									Deprecated: []string{
										"some-deprecated-api-version",
									},
									Supported: []string{
										"some-api-version",
										"other-api-version",
									},
								},
							},
							RunImages: []occam.BuilderInfoRunImage{
								{Name: "some-run-image"},
							},
							Buildpacks: []occam.BuilderInfoBuildpack{
								{
									ID:       "some-buildpack-id",
									Name:     "some-buildpack-name",
									Version:  "some-buildpack-version",
									Homepage: "some-buildpack-homepage",
								},
								{
									ID:       "other-buildpack-id",
									Name:     "other-buildpack-name",
									Version:  "other-buildpack-version",
									Homepage: "other-buildpack-homepage",
								},
							},
							DetectionOrder: []occam.BuilderInfoDetectionOrder{
								{
									Buildpacks: []occam.BuilderInfoDetectionOrderBuildpack{
										{
											ID:      "some-buildpack-id",
											Version: "some-buildpack-version",
											Buildpacks: []occam.BuilderInfoDetectionOrderBuildpack{
												{
													ID:       "other-buildpack-id",
													Version:  "other-buildpack-version",
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						RemoteInfo: occam.BuilderInfo{
							Description: "some-builder-description",
							CreatedBy: occam.BuilderInfoCreatedBy{
								Name:    "some-creator-name",
								Version: "some-creator-version",
							},
							Stack: occam.BuilderInfoStack{
								ID: "some-stack-id",
							},
							Lifecycle: occam.BuilderInfoLifecycle{
								Version: "some-lifecycle-version",
								BuildpackAPIs: occam.BuilderInfoLifecycleAPIs{
									Deprecated: []string{
										"some-deprecated-api-version",
									},
									Supported: []string{
										"some-api-version",
										"other-api-version",
									},
								},
								PlatformAPIs: occam.BuilderInfoLifecycleAPIs{
									Deprecated: []string{
										"some-deprecated-api-version",
									},
									Supported: []string{
										"some-api-version",
										"other-api-version",
									},
								},
							},
							RunImages: []occam.BuilderInfoRunImage{
								{Name: "some-run-image"},
							},
							Buildpacks: []occam.BuilderInfoBuildpack{
								{
									ID:       "some-buildpack-id",
									Name:     "some-buildpack-name",
									Version:  "some-buildpack-version",
									Homepage: "some-buildpack-homepage",
								},
								{
									ID:       "other-buildpack-id",
									Name:     "other-buildpack-name",
									Version:  "other-buildpack-version",
									Homepage: "other-buildpack-homepage",
								},
							},
							DetectionOrder: []occam.BuilderInfoDetectionOrder{
								{
									Buildpacks: []occam.BuilderInfoDetectionOrderBuildpack{
										{
											ID:      "some-buildpack-id",
											Version: "some-buildpack-version",
											Buildpacks: []occam.BuilderInfoDetectionOrderBuildpack{
												{
													ID:       "other-buildpack-id",
													Version:  "other-buildpack-version",
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
					}))

					Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
						"builder", "inspect",
						"--output", "json",
					}))
				})
			})

			context("when given a specific builder name", func() {
				it.Before(func() {
					executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
						fmt.Fprintln(execution.Stdout, `{
							"local_info": {
								"stack": {
									"id": "other-stack-id"
								}
							},
							"remote_info": {
								"stack": {
									"id": "other-stack-id"
								}
							}
						}`)
						return nil
					}
				})

				it("returns the builder matching that name", func() {
					builder, err := pack.Builder.Inspect.Execute("other-builder")
					Expect(err).NotTo(HaveOccurred())
					Expect(builder).To(Equal(occam.Builder{
						LocalInfo: occam.BuilderInfo{
							Stack: occam.BuilderInfoStack{
								ID: "other-stack-id",
							},
						},
						RemoteInfo: occam.BuilderInfo{
							Stack: occam.BuilderInfoStack{
								ID: "other-stack-id",
							},
						},
					}))

					Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
						"builder", "inspect", "other-builder",
						"--output", "json",
					}))
				})
			})

			context("failure cases", func() {
				context("when the pack executable fails", func() {
					it.Before(func() {
						executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
							fmt.Fprint(execution.Stdout, "some failure message")
							return errors.New("some error")
						}
					})

					it("returns an error", func() {
						_, err := pack.Builder.Inspect.Execute()
						Expect(err).To(MatchError("failed to pack builder inspect: some error\n\nOutput:\nsome failure message"))
					})
				})

				context("when the pack returns unparseable JSON", func() {
					it.Before(func() {
						executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
							fmt.Fprintln(execution.Stdout, "%%%")
							return nil
						}
					})

					it("returns an error", func() {
						_, err := pack.Builder.Inspect.Execute()
						Expect(err).To(MatchError(ContainSubstring("failed to parse JSON: invalid character '%'")))
					})
				})
			})
		})
	})
}
