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
				executable.ExecuteCall.Stub = func(execution pexec.Execution) (string, string, error) {
					fmt.Fprintln(execution.Stdout, `[
						{
							"Id": "some-image-id",
							"Config": {
								"Labels": {
									"io.buildpacks.lifecycle.metadata": "{\"buildpacks\": [{\"key\": \"some-buildpack\", \"layers\": {\"some-layer\": {\"sha\": \"some-sha\", \"build\": true, \"launch\": true, \"cache\": true}}}]}"
								}
							}
						}
					]`)

					return "", "", nil
				}
			})

			it("returns an image given a name", func() {
				image, err := docker.Image.Inspect("some-app:latest")
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
								},
							},
						},
					},
				}))
			})

			context("failure cases", func() {
				context("when the executable fails", func() {
					it.Before(func() {
						executable.ExecuteCall.Stub = func(execution pexec.Execution) (string, string, error) {
							fmt.Fprintln(execution.Stdout, "[]")
							fmt.Fprintln(execution.Stderr, "Error: No such image: some-app:latest")
							return "", "", errors.New("exit status 1")
						}
					})

					it("returns an error", func() {
						_, err := docker.Image.Inspect("some-app:latest")
						Expect(err).To(MatchError("failed to inspect docker image: exit status 1: Error: No such image: some-app:latest"))
					})
				})

				context("when malformed json is given", func() {
					it.Before(func() {
						executable.ExecuteCall.Stub = func(execution pexec.Execution) (string, string, error) {
							fmt.Fprintln(execution.Stdout, `malformed json %%%%`)
							return "", "", nil
						}
					})

					it("return an error", func() {
						_, err := docker.Image.Inspect("some-app:latest")
						Expect(err).To(MatchError(ContainSubstring("failed to inspect docker image")))
						Expect(err).To(MatchError(ContainSubstring("invalid character")))
					})
				})

				context("when then lifecycle metadata has malformed json", func() {
					it.Before(func() {
						executable.ExecuteCall.Stub = func(execution pexec.Execution) (string, string, error) {
							fmt.Fprintln(execution.Stdout, `[{"Config": {"Labels": {"io.buildpacks.lifecycle.metadata": "%%%"}}}]`)
							return "", "", nil
						}
					})

					it("return an error", func() {
						_, err := docker.Image.Inspect("some-app:latest")
						Expect(err).To(MatchError(ContainSubstring("failed to inspect docker image")))
						Expect(err).To(MatchError(ContainSubstring("invalid character")))
					})
				})
			})
		})
	})
}
