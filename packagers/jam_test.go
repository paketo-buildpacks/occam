package packagers_test

import (
	"errors"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/occam/fakes"
	"github.com/paketo-buildpacks/occam/packagers"
	"github.com/sclevine/spec"
)

func testJam(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		executable *fakes.Executable
		packager   packagers.Jam
	)

	it.Before(func() {
		executable = &fakes.Executable{}
		packager = packagers.NewJam().WithExecutable(executable)

	})

	context("Execute", func() {
		it("creates a correct pexec.Execution", func() {
			err := packager.Execute("some-buildpack-dir", "some-output", "some-version", false)
			Expect(err).NotTo(HaveOccurred())

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
				"pack",
				"--buildpack", filepath.Join("some-buildpack-dir", "buildpack.toml"),
				"--output", "some-output",
				"--version", "some-version",
			}))
		})

		context("when packaging with offline dependencies", func() {
			it("creates a correct pexec.Execution", func() {
				err := packager.Execute("some-buildpack-dir", "some-output", "some-version", true)
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"pack",
					"--buildpack", filepath.Join("some-buildpack-dir", "buildpack.toml"),
					"--output", "some-output",
					"--version", "some-version",
					"--offline",
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
