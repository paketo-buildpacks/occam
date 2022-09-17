package matchers_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam/matchers"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBeAFileMatching(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect   = NewWithT(t).Expect
		filename string
		matcher  *matchers.BeAFileMatchingMatcher
	)

	it.Before(func() {
		filename = filepath.Join(t.TempDir(), "file")

		Expect(os.WriteFile(filename, []byte("hello world"), os.ModePerm)).To(Succeed())
	})

	context("will wrap ContainsSubstring", func() {
		it("will pass", func() {
			matcher = matchers.BeAFileMatching(ContainSubstring("hello"))

			match, err := matcher.Match(filename)
			Expect(match).To(Equal(true))
			Expect(err).NotTo(HaveOccurred())
		})

		it("will fail", func() {
			matcher = matchers.BeAFileMatching(ContainSubstring("foobar"))

			match, err := matcher.Match(filename)
			Expect(match).To(Equal(false))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	context("will wrap Equal", func() {
		it("will pass", func() {
			matcher = matchers.BeAFileMatching(Equal("hello world"))

			match, err := matcher.Match(filename)
			Expect(match).To(Equal(true))
			Expect(err).NotTo(HaveOccurred())
		})

		it("will fail", func() {
			matcher = matchers.BeAFileMatching(Equal("foobar"))

			match, err := matcher.Match(filename)
			Expect(match).To(Equal(false))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	context("will wrap MatchJSON", func() {
		it.Before(func() {
			Expect(os.WriteFile(filename, []byte(`[
  {
    "a": "b"
  },
  {
    "c": [
      1,
      2,
      3
    ]
  }
]`), os.ModePerm)).To(Succeed())
		})

		it("will pass", func() {
			matcher = matchers.BeAFileMatching(MatchJSON(`[{"a":"b"},{"c":[1,2,3]}]`))

			match, err := matcher.Match(filename)
			Expect(match).To(Equal(true))
			Expect(err).NotTo(HaveOccurred())
		})

		it("will fail", func() {
			matcher = matchers.BeAFileMatching(MatchJSON(`{}`))

			match, err := matcher.Match(filename)
			Expect(match).To(Equal(false))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	context("FailureMessage", func() {
		it("wraps the given matcher", func() {
			wrappedMatcher := Equal("hello")
			matcher = matchers.BeAFileMatching(wrappedMatcher)

			_, err := matcher.Match(filename)
			Expect(err).NotTo(HaveOccurred())

			message := matcher.FailureMessage(filename)
			Expect(message).To(Equal(wrappedMatcher.FailureMessage("hello world")))
		})
	})

	context("NegatedFailureMessage", func() {
		it("wraps the given matcher", func() {
			wrappedMatcher := BeNil()
			matcher = matchers.BeAFileMatching(wrappedMatcher)

			_, err := matcher.Match(filename)
			Expect(err).NotTo(HaveOccurred())

			message := matcher.NegatedFailureMessage(filename)
			Expect(message).To(Equal(wrappedMatcher.NegatedFailureMessage("hello world")))
		})
	})

	context("failure cases", func() {
		context("Match is called with a non-string", func() {
			it("will return an error", func() {
				matcher = matchers.BeAFileMatching(ContainSubstring(""))
				_, err := matcher.Match(42)
				Expect(err).To(MatchError("BeAFileMatchingMatcher expects a file path"))
			})
		})

		context("invalid filename", func() {
			it.Before(func() {
				filename = "foo/bar"
			})

			it("will return an error", func() {
				matcher = matchers.BeAFileMatching(ContainSubstring(""))
				_, err := matcher.Match(filename)
				Expect(os.IsNotExist(err)).To(BeTrue())
			})
		})

		context("insufficient permissions", func() {
			it.Before(func() {
				tempDir := t.TempDir()
				filename = filepath.Join(tempDir, "file")
				Expect(os.Chmod(tempDir, 0000)).To(Succeed())
			})

			it("will return an error", func() {
				matcher = matchers.BeAFileMatching(ContainSubstring(""))
				_, err := matcher.Match(filename)
				Expect(os.IsPermission(err)).To(BeTrue())
			})
		})
	})
}
