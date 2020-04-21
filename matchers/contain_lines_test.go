package matchers_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cloudfoundry/occam/matchers"
	"github.com/onsi/gomega/types"
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
							"[detector] zeroth-line",
							"[builder] first-line",
							"[builder] second-line",
							"[builder] some-line-content",
							"[builder] fourth-line",
							"[exporter] fifth-line",
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
							"[detector] zeroth-line",
							"[builder] first-line",
							"[builder] second-line",
							"[builder] third-line",
							"[builder] fourth-line",
							"[exporter] fifth-line",
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
							"[detector] zeroth-line",
							"[builder] first-line",
							"[builder] second-line",
							"[builder] some-line-content",
							"[builder] fourth-line",
							"[exporter] fifth-line",
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
							"[detector] zeroth-line",
							"[builder] first-line",
							"[builder] second-line",
							"[builder] third-line",
							"[builder] fourth-line",
							"[exporter] fifth-line",
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
						"[detector] zeroth-line",
						"[builder] first-line",
						"[builder] second-line",
						"[builder] some-line-content",
						"[builder] other-line-content",
						"[builder] another-line-content",
						"[builder] sixth-line",
						"[exporter] seventh-line",
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
						"[detector] zeroth-line",
						"[builder] first-line",
						"[builder] second-line",
						"[builder] third-line",
						"[builder] fourth-line",
						"[builder] fifth-line",
						"[builder] sixth-line",
						"[exporter] seventh-line",
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
						"[detector] zeroth-line",
						"[builder] first-line",
						"[builder] second-line",
						"[builder] some-line-content",
						"[builder] other-line-content",
						"[builder] another-line-content",
						"[builder] sixth-line",
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
						"[detector] zeroth-line",
						"[builder] first-line",
						"[builder] second-line",
						"[builder] third-line",
						"[builder] fourth-line",
						"[builder] fifth-line",
						"[builder] sixth-line",
						"[exporter] seventh-line",
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
					_, err := matcher.Match("[builder] some-line-content")
					Expect(err).To(MatchError(ContainSubstring("MatchJSONMatcher matcher requires")))
				})
			})

			context("when there are no lines with the [builder] prefix", func() {
				it.Before(func() {
					matcher = matchers.ContainLines("some-line-content")
				})

				it("returns an error", func() {
					_, err := matcher.Match("some-line-content")
					Expect(err).To(MatchError("ContainLinesMatcher requires lines with [builder] prefix, found none:     <string>: some-line-content"))
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
				"[detector] zeroth-line",
				"[builder] first-line",
				"[builder] second-line",
				"[builder] third-line",
				"[builder] fourth-line",
				"[exporter] fifth-line",
			}, "\n"))
			Expect(message).To(ContainSubstring(strings.TrimSpace(`
Expected
    <string>: 
    first-line
    second-line
    third-line
    fourth-line
to contain lines
    <[]interface {} | len:4, cap:4>: [
        "some-line-content",
        {
            Regexp: "some\\-.+\\-content",
            Args: nil,
        },
        {Prefix: "third", Args: nil},
        {
            Substr: "other-line-con",
            Args: nil,
        },
    ]
but missing
    <[]interface {} | len:3, cap:4>: [
        "some-line-content",
        {
            Regexp: "some\\-.+\\-content",
            Args: nil,
        },
        {
            Substr: "other-line-con",
            Args: nil,
        },
    ]
`)))
		})

		context("when all lines appear, but are misordered", func() {
			it("returns a useful error message", func() {
				message := matcher.FailureMessage(strings.Join([]string{
					"[detector] zeroth-line",
					"[builder] some-stuff-content",
					"[builder] some-line-content",
					"[builder] third-line",
					"[builder] other-line-content",
					"[exporter] fifth-line",
				}, "\n"))
				Expect(message).To(ContainSubstring(strings.TrimSpace(`
Expected
    <string>: 
    some-stuff-content
    some-line-content
    third-line
    other-line-content
to contain lines
    <[]interface {} | len:4, cap:4>: [
        "some-line-content",
        {
            Regexp: "some\\-.+\\-content",
            Args: nil,
        },
        {Prefix: "third", Args: nil},
        {
            Substr: "other-line-con",
            Args: nil,
        },
    ]
all lines appear, but may be misordered
`)))
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
				"[detector] zeroth-line",
				"[builder] first-line",
				"[builder] second-line",
				"[builder] third-line",
				"[builder] fourth-line",
				"[exporter] fifth-line",
			}, "\n"))
			Expect(message).To(ContainSubstring(strings.TrimSpace(`
Expected
    <string>: 
    first-line
    second-line
    third-line
    fourth-line
not to contain lines
    <[]interface {} | len:4, cap:4>: [
        "some-line-content",
        {
            Regexp: "some\\-.+\\-content",
            Args: nil,
        },
        {Prefix: "third", Args: nil},
        {
            Substr: "other-line-con",
            Args: nil,
        },
    ]
but includes
    <[]interface {} | len:1, cap:1>: [
        {Prefix: "third", Args: nil},
    ]
`)))
		})
	})
}
