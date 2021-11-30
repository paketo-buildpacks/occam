package packagers_test

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/occam/fakes"
	"github.com/paketo-buildpacks/occam/packagers"
	"github.com/sclevine/spec"
)

func testLibpak(t *testing.T, context spec.G, it spec.S) {

	var (
		Expect     = NewWithT(t).Expect
		executable *fakes.Executable
		packager   packagers.Libpak
	)

	it.Before(func() {
		executable = &fakes.Executable{}
		packager = packagers.NewLibpak().WithExecutable(executable)
	})

	context("Execute", func() {
		it("calls the executable with the correct arguments", func() {
			err := packager.Execute("some-buildpack-dir", "some-output-dir", "some-version", false)

			Expect(err).NotTo(HaveOccurred())
			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
				"--destination", "some-output-dir",
				"--version", "some-version",
			}))
			Expect(executable.ExecuteCall.Receives.Execution.Dir).To(Equal("some-buildpack-dir"))
		})

		context("when packaging with offline dependencies", func() {
			it("adds the appropriate flag to the packager args", func() {
				err := packager.Execute("some-buildpack-dir", "some-output-dir", "some-version", true)

				Expect(err).NotTo(HaveOccurred())
				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"--destination", "some-output-dir",
					"--version", "some-version",
					"--include-dependencies",
				}))
			})
		})

		context("failure cases", func() {
			context("when the execution returns an error", func() {
				it.Before(func() {
					executable.ExecuteCall.Returns.Error = errors.New("some error")
				})
				it("returns an error", func() {
					err := packager.Execute("some-buildpack-dir", "some-output", "some-version", true)
					Expect(err).To(MatchError("some error"))
				})
			})
		})
	})
}
