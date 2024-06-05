package packagers_test

import (
	"errors"
	"fmt"
	"runtime"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/occam/fakes"
	"github.com/paketo-buildpacks/occam/packagers"
	"github.com/sclevine/spec"
)

func testLibpak(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect = NewWithT(t).Expect

		executable *fakes.Executable
		pack       *fakes.Executable

		tempOutput func(string, string) (string, error)

		packager packagers.Libpak
	)

	it.Before(func() {
		executable = &fakes.Executable{}
		pack = &fakes.Executable{}

		tempOutput = func(string, string) (string, error) {
			return "some-libpak-output-dir", nil
		}

		packager = packagers.NewLibpak().WithExecutable(executable).WithPack(pack).WithTempOutput(tempOutput)
	})

	context("Execute", func() {
		it("calls the executable with the correct arguments", func() {
			err := packager.Execute("some-buildpack-dir", "some-output-dir", "some-version", false)

			Expect(err).NotTo(HaveOccurred())
			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
				"--destination", "some-libpak-output-dir",
				"--version", "some-version",
			}))
			Expect(executable.ExecuteCall.Receives.Execution.Dir).To(Equal("some-buildpack-dir"))

			Expect(pack.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
				"buildpack", "package",
				"some-output-dir",
				"--path", "some-libpak-output-dir",
				"--format", "file",
				"--target", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			}))
		})

		context("when packaging with offline dependencies", func() {
			it("adds the appropriate flag to the packager args", func() {
				err := packager.Execute("some-buildpack-dir", "some-output-dir", "some-version", true)

				Expect(err).NotTo(HaveOccurred())
				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"--destination", "some-libpak-output-dir",
					"--version", "some-version",
					"--include-dependencies",
				}))

				Expect(pack.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"buildpack", "package",
					"some-output-dir",
					"--path", "some-libpak-output-dir",
					"--format", "file",
					"--target", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
				}))
			})
		})

		context("failure cases", func() {
			context("when the tempDir creation fails returns an error", func() {
				it.Before(func() {
					tempOutput = func(string, string) (string, error) {
						return "", errors.New("some tempDir error")
					}

					packager = packager.WithTempOutput(tempOutput)
				})
				it("returns an error", func() {
					err := packager.Execute("some-buildpack-dir", "some-output", "some-version", true)
					Expect(err).To(MatchError("some tempDir error"))
				})
			})

			context("when the libpak execution returns an error", func() {
				it.Before(func() {
					executable.ExecuteCall.Returns.Error = errors.New("some libpak error")
				})
				it("returns an error", func() {
					err := packager.Execute("some-buildpack-dir", "some-output", "some-version", true)
					Expect(err).To(MatchError("some libpak error"))
				})
			})

			context("when the pack execution returns an error", func() {
				it.Before(func() {
					pack.ExecuteCall.Returns.Error = errors.New("some pack error")
				})
				it("returns an error", func() {
					err := packager.Execute("some-buildpack-dir", "some-output", "some-version", true)
					Expect(err).To(MatchError("some pack error"))
				})
			})
		})
	})
}
