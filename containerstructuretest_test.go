package occam_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/occam"
	"github.com/paketo-buildpacks/occam/fakes"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/sclevine/spec"
)

func testContainerStructureTest(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		executable *fakes.Executable
		cst        occam.ContainerStructureTest
	)

	it.Before(func() {
		executable = &fakes.Executable{}
		executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
			fmt.Fprintln(execution.Stdout, "some stdout output")
			fmt.Fprintln(execution.Stderr, "some stderr output")
			return nil
		}

		cst = occam.NewContainerStructureTest().WithExecutable(executable)
	})

	context("ContainerStructureTest", func() {
		it("should work", func() {
			logs, err := cst.Execute("test/my-image", "tests.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(logs).To(Equal("some stdout output\nsome stderr output\n"))

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
				"test",
				"--config",
				"tests.yaml",
				"--image", "test/my-image",
			}))
		})

		context("WithVerbose", func() {
			it("should add --verbosity flag", func() {
				cst = cst.WithVerbose()

				_, err := cst.Execute("test/my-image", "tests.yaml")
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"test",
					"--verbosity",
					"debug",
					"--config",
					"tests.yaml",
					"--image", "test/my-image",
				}))
			})
		})

		context("WithNoColor", func() {
			it("should add --no-color flag", func() {
				cst = cst.WithNoColor()

				_, err := cst.Execute("test/my-image", "tests.yaml")
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"test",
					"--no-color",
					"--config",
					"tests.yaml",
					"--image", "test/my-image",
				}))
			})
		})

		context("WithPull", func() {
			it("should add --pull flag", func() {
				cst = cst.WithPull()

				_, err := cst.Execute("test/my-image", "tests.yaml")
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"test",
					"--pull",
					"--config",
					"tests.yaml",
					"--image", "test/my-image",
				}))
			})
		})
	})
}
