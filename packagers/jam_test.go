package packagers_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
		pack       *fakes.Executable

		tempOutput func(string, string) (string, error)

		packager packagers.Jam
	)

	it.Before(func() {
		executable = &fakes.Executable{}
		pack = &fakes.Executable{}

		tempOutput = func(string, string) (string, error) {
			return "some-jam-output", nil
		}

		packager = packagers.NewJam().WithExecutable(executable).WithPack(pack).WithTempOutput(tempOutput)

	})

	context("Execute", func() {
		it("creates a correct pexec.Execution", func() {
			err := packager.Execute("some-buildpack-dir", "some-output", "some-version", false)
			Expect(err).NotTo(HaveOccurred())

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
				"pack",
				"--buildpack", filepath.Join("some-buildpack-dir", "buildpack.toml"),
				"--output", filepath.Join("some-jam-output", "some-version.tgz"),
				"--version", "some-version",
			}))

			Expect(pack.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
				"buildpack", "package",
				"some-output",
				"--format", "file",
				"--target", fmt.Sprintf("linux/%s", runtime.GOARCH),
			}))
		})

		context("when packaging with offline dependencies", func() {
			it("creates a correct pexec.Execution", func() {
				err := packager.Execute("some-buildpack-dir", "some-output", "some-version", true)
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"pack",
					"--buildpack", filepath.Join("some-buildpack-dir", "buildpack.toml"),
					"--output", filepath.Join("some-jam-output", "some-version.tgz"),
					"--version", "some-version",
					"--offline",
				}))

				Expect(pack.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"buildpack", "package",
					"some-output",
					"--format", "file",
					"--target", fmt.Sprintf("linux/%s", runtime.GOARCH),
				}))
			})
		})

		context("when packaging a stack extension", func() {
			var extensionDir string

			it.Before(func() {
				var err error
				extensionDir, err = os.MkdirTemp("", "")
				Expect(err).NotTo(HaveOccurred())

				_, err = os.Create(filepath.Join(extensionDir, "extension.toml"))
				Expect(err).NotTo(HaveOccurred())
			})

			it.After(func() {
				Expect(os.RemoveAll(extensionDir)).To(Succeed())
			})

			it("creates a correct pexec.Execution", func() {
				err := packager.Execute(extensionDir, "some-output", "some-version", true)
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"pack",
					"--extension", filepath.Join(extensionDir, "extension.toml"),
					"--output", filepath.Join("some-jam-output", "some-version.tgz"),
					"--version", "some-version",
					"--offline",
				}))

				Expect(pack.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"extension", "package",
					"some-output",
					"--format", "file",
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

			context("when the jam execution returns an error", func() {
				it.Before(func() {
					executable.ExecuteCall.Returns.Error = errors.New("some jam error")
				})
				it("returns an error", func() {
					err := packager.Execute("some-buildpack-dir", "some-output", "some-version", true)
					Expect(err).To(MatchError("some jam error"))
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
