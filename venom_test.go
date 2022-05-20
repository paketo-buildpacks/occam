package occam_test

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/occam"
	"github.com/paketo-buildpacks/occam/fakes"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/sclevine/spec"
)

func testVenom(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		executable *fakes.Executable
		venom      occam.Venom
	)

	it.Before(func() {
		executable = &fakes.Executable{}
		executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
			fmt.Fprintln(execution.Stdout, "some stdout output")
			fmt.Fprintln(execution.Stderr, "some stderr output")
			return nil
		}

		venom = occam.NewVenom().WithExecutable(executable)
	})

	context("Venom", func() {
		it("should work", func() {
			logs, err := venom.Execute("test.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(logs).To(Equal("some stdout output\nsome stderr output\n"))

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
				"run",
				"test.yaml",
			}))
		})

		it("should propagate an error", func() {
			executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
				return errors.New("Test Error")
			}

			_, err := venom.Execute("test.yaml")
			Expect(err).To(MatchError(ContainSubstring("Test Error")))
		})

		context("WithVerbose", func() {
			it("should add -vv flag", func() {
				_, err := venom.WithVerbose().Execute("test.yaml")
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"run",
					"-vv",
					"test.yaml",
				}))
			})
		})

		context("WithPort", func() {
			it("should add port variable", func() {
				_, err := venom.WithPort("8080").Execute("test.yaml")
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"run",
					"--var",
					"port=8080",
					"test.yaml",
				}))
			})
		})

		context("WithVars", func() {
			it("it should preserve order of provided vars", func() {
				_, err := venom.WithVar("key1", "value1").
					WithVar("key2", "value2").
					WithVar("key2", "value2").
					WithVar("key3", "value3").
					WithVar("key4", "value4").
					Execute("test.yaml")
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"run",
					"--var",
					"key1=value1",
					"--var",
					"key2=value2",
					"--var",
					"key3=value3",
					"--var",
					"key4=value4",
					"test.yaml",
				}))
			})

		})
	})
}
