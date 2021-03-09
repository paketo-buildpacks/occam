package matchers_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/onsi/gomega/types"
	"github.com/paketo-buildpacks/occam/matchers"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testContainLines(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		matcher types.GomegaMatcher
	)

	context("Match", func() {
		var actual interface{}

		context("when the matcher expects a single, simple text line", func() {
			it.Before(func() {
				matcher = matchers.ContainLines("some-line-content")
			})

			context("when the actual value is a string", func() {
				context("when the actual value does match", func() {
					it.Before(func() {
						actual = strings.Join([]string{
							"zeroth-line",
							"first-line",
							"second-line",
							"some-line-content",
							"fourth-line",
							"fifth-line",
						}, "\n")
					})

					it("returns true", func() {
						result, err := matcher.Match(actual)
						Expect(err).NotTo(HaveOccurred())
						Expect(result).To(BeTrue())
					})
				})

				context("when the actual value does not match", func() {
					it.Before(func() {
						actual = strings.Join([]string{
							"zeroth-line",
							"first-line",
							"second-line",
							"third-line",
							"fourth-line",
							"fifth-line",
						}, "\n")
					})

					it("returns false", func() {
						result, err := matcher.Match(actual)
						Expect(err).NotTo(HaveOccurred())
						Expect(result).To(BeFalse())
					})
				})
			})

			context("when the actual value is a fmt.Stringer", func() {
				context("when the actual value does match", func() {
					it.Before(func() {
						actual = bytes.NewBufferString(strings.Join([]string{
							"zeroth-line",
							"first-line",
							"second-line",
							"some-line-content",
							"fourth-line",
							"fifth-line",
						}, "\n"))
					})

					it("returns true", func() {
						result, err := matcher.Match(actual)
						Expect(err).NotTo(HaveOccurred())
						Expect(result).To(BeTrue())
					})
				})

				context("when the actual value does not match", func() {
					it.Before(func() {
						actual = bytes.NewBufferString(strings.Join([]string{
							"zeroth-line",
							"first-line",
							"second-line",
							"third-line",
							"fourth-line",
							"fifth-line",
						}, "\n"))
					})

					it("returns false", func() {
						result, err := matcher.Match(actual)
						Expect(err).NotTo(HaveOccurred())
						Expect(result).To(BeFalse())
					})
				})
			})
		})

		context("when the matcher expects multiple, simple text lines", func() {
			it.Before(func() {
				matcher = matchers.ContainLines(
					"some-line-content",
					"other-line-content",
					"another-line-content",
				)
			})

			context("when the actual value does match", func() {
				it.Before(func() {
					actual = bytes.NewBufferString(strings.Join([]string{
						"zeroth-line",
						"first-line",
						"second-line",
						"some-line-content",
						"other-line-content",
						"another-line-content",
						"sixth-line",
						"seventh-line",
					}, "\n"))
				})

				it("returns true", func() {
					result, err := matcher.Match(actual)
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(BeTrue())
				})
			})

			context("when the actual value does not match", func() {
				it.Before(func() {
					actual = bytes.NewBufferString(strings.Join([]string{
						"zeroth-line",
						"first-line",
						"second-line",
						"third-line",
						"fourth-line",
						"fifth-line",
						"sixth-line",
						"seventh-line",
					}, "\n"))
				})

				it("returns false", func() {
					result, err := matcher.Match(actual)
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(BeFalse())
				})
			})
		})

		context("when the matcher expects multiple, submatcher lines", func() {
			it.Before(func() {
				matcher = matchers.ContainLines(
					MatchRegexp(`some\-.+\-content`),
					HavePrefix("other-line"),
					ContainSubstring("other-line-con"),
				)
			})

			context("when the actual value does match", func() {
				it.Before(func() {
					actual = strings.Join([]string{
						"zeroth-line",
						"first-line",
						"second-line",
						"some-line-content",
						"other-line-content",
						"another-line-content",
						"sixth-line",
						"seventh-line",
					}, "\n")
				})

				it("returns true", func() {
					result, err := matcher.Match(actual)
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(BeTrue())
				})
			})

			context("when the actual value has line prefixes", func() {
				it.Before(func() {
					actual = strings.Join([]string{
						"[detector] zeroth-line",
						"[builder] first-line",
						"[builder] second-line",
						"[builder] some-line-content",
						"[builder] other-line-content",
						"[builder] another-line-content",
						"[analyzer] sixth-line",
						"[exporter] seventh-line",
					}, "\n")
				})

				it("returns true", func() {
					result, err := matcher.Match(actual)
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(BeTrue())
				})
			})

			context("when the actual value does not match", func() {
				it.Before(func() {
					actual = strings.Join([]string{
						"zeroth-line",
						"first-line",
						"second-line",
						"third-line",
						"fourth-line",
						"fifth-line",
						"sixth-line",
						"seventh-line",
					}, "\n")
				})

				it("returns false", func() {
					result, err := matcher.Match(actual)
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(BeFalse())
				})
			})
		})

		context("failure cases", func() {
			context("when the actual is not a string or fmt.Stringer", func() {
				it.Before(func() {
					matcher = matchers.ContainLines("some-line-content")
				})

				it("returns an error", func() {
					_, err := matcher.Match(struct{}{})
					Expect(err).To(MatchError("ContainLinesMatcher requires a string or fmt.Stringer. Got actual:     <struct {}>: {}"))
				})
			})

			context("when the submatcher fails", func() {
				it.Before(func() {
					matcher = matchers.ContainLines(MatchJSON(struct{}{}))
				})

				it("returns an error", func() {
					_, err := matcher.Match("some-line-content")
					Expect(err).To(MatchError(ContainSubstring("MatchJSONMatcher matcher requires")))
				})
			})
		})
	})

	context("FailureMessage", func() {
		it.Before(func() {
			matcher = matchers.ContainLines(
				"some-line-content",
				MatchRegexp(`some\-.+\-content`),
				HavePrefix("third"),
				ContainSubstring("other-line-con"),
			)
		})

		it("returns a useful error message", func() {
			message := matcher.FailureMessage(strings.Join([]string{
				"zeroth-line",
				"first-line",
				"second-line",
				"third-line",
				"fourth-line",
				"fifth-line",
			}, "\n"))
			Expect(message).To(MatchRegexp(strings.TrimSpace(`
Expected
    <string>: 
    zeroth-line
    first-line
    second-line
    third-line
    fourth-line
    fifth-line
to contain lines
    <\[\]interface {} \| len:4, cap:4>: \[
        <string>"some\-line\-content",
        <\*matchers.MatchRegexpMatcher \| \S+>{
            Regexp: "some\\\\\-.+\\-content",
            Args: nil,
        },
        <\*matchers.HavePrefixMatcher \| \S+>{Prefix: "third", Args: nil},
        <\*matchers.ContainSubstringMatcher \| \S+>{
            Substr: "other\-line\-con",
            Args: nil,
        },
    \]
but missing
    <\[\]interface {} \| len:3, cap:4>: \[
        <string>"some\-line\-content",
        <\*matchers.MatchRegexpMatcher \| \S+>{
            Regexp: "some\\\\\-.+\\-content",
            Args: nil,
        },
        <\*matchers.ContainSubstringMatcher \| \S+>{
            Substr: "other\-line\-con",
            Args: nil,
        },
    \]
`)))
		})

		context("when all lines appear, but are misordered", func() {
			it("returns a useful error message", func() {
				message := matcher.FailureMessage(strings.Join([]string{
					"zeroth-line",
					"some-stuff-content",
					"some-line-content",
					"third-line",
					"other-line-content",
					"fifth-line",
				}, "\n"))
				Expect(message).To(ContainSubstring(strings.TrimSpace(`all lines appear, but may be misordered`)))
			})
		})
	})

	context("NegatedFailureMessage", func() {
		it.Before(func() {
			matcher = matchers.ContainLines(
				"some-line-content",
				MatchRegexp(`some\-.+\-content`),
				HavePrefix("third"),
				ContainSubstring("other-line-con"),
			)
		})

		it("returns a useful error message", func() {
			message := matcher.NegatedFailureMessage(strings.Join([]string{
				"zeroth-line",
				"first-line",
				"second-line",
				"third-line",
				"fourth-line",
				"fifth-line",
			}, "\n"))
			Expect(message).To(MatchRegexp(strings.TrimSpace(`
Expected
    <string>: 
    zeroth-line
    first-line
    second-line
    third-line
    fourth-line
    fifth-line
not to contain lines
    <\[\]interface {} \| len:4, cap:4>: \[
        <string>"some\-line\-content",
        <\*matchers.MatchRegexpMatcher \| \S+>{
            Regexp: "some\\\\\-.+\\-content",
            Args: nil,
        },
        <\*matchers.HavePrefixMatcher \| \S+>{Prefix: "third", Args: nil},
        <\*matchers.ContainSubstringMatcher \| \S+>{
            Substr: "other\-line\-con",
            Args: nil,
        },
    \]
but includes
    <\[\]interface {} \| len:1, cap:1>: \[
        <\*matchers.HavePrefixMatcher \| \S+>{Prefix: "third", Args: nil},
    \]
`)))
		})
	})
}
