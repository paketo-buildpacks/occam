package occam_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSource(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		source      string
		destination string
	)

	it.Before(func() {
		var err error
		source, err = os.MkdirTemp("", "source")
		Expect(err).NotTo(HaveOccurred())

		err = os.WriteFile(filepath.Join(source, "some-file"), []byte("some-content"), 0644)
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(source)).To(Succeed())
		Expect(os.RemoveAll(destination)).To(Succeed())
	})

	it("copies the given directory to a temporary directory with a random file added for uniqueness", func() {
		var err error
		destination, err = occam.Source(source)
		Expect(err).NotTo(HaveOccurred())
		Expect(destination).To(BeADirectory())

		content, err := os.ReadFile(filepath.Join(destination, "some-file"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("some-content"))

		content, err = os.ReadFile(filepath.Join(destination, ".occam-key"))
		Expect(err).NotTo(HaveOccurred())
		Expect(content).To(HaveLen(32))
	})

	context("failure cases", func() {
		context("when the source cannot be copied", func() {
			it.Before(func() {
				Expect(os.Chmod(filepath.Join(source, "some-file"), 0000)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := occam.Source(source)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})
	})
}
