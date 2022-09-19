package matchers

import (
	"fmt"
	"os"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

type BeAFileMatchingMatcher struct {
	expected        interface{}
	expectedMatcher types.GomegaMatcher
	contents        string
}

func BeAFileMatching(expected interface{}) *BeAFileMatchingMatcher {
	return &BeAFileMatchingMatcher{
		expected: expected,
	}
}

// Match will return whether the expectedMatcher matches on the file contents.
// The file contents are saved for use in FailureMessage and NegatedFailureMessage.
func (matcher *BeAFileMatchingMatcher) Match(actual interface{}) (success bool, err error) {
	actualFilename, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("BeAFileMatchingMatcher expects a file path")
	}

	if expectedString, ok := matcher.expected.(string); ok {
		matcher.expectedMatcher = gomega.Equal(expectedString)
	} else {
		if matcher.expectedMatcher, ok = matcher.expected.(types.GomegaMatcher); !ok {
			return false, fmt.Errorf("BeAFileMatching expects a string or a types.GomegaMatcher")
		}
	}

	bytes, err := os.ReadFile(actualFilename)
	if err != nil {
		return false, err
	}

	matcher.contents = string(bytes)

	return matcher.expectedMatcher.Match(matcher.contents)
}

// FailureMessage does not use the given param since it returns the expectedMatcher's failure message.
// Requires Match to be called first to populate the file contents.
func (matcher *BeAFileMatchingMatcher) FailureMessage(_ interface{}) string {
	return matcher.expectedMatcher.FailureMessage(matcher.contents)
}

// NegatedFailureMessage does not use the given param since it returns the expectedMatcher's negated failure message.
// Requires Match to be called first to populate the file contents.
func (matcher *BeAFileMatchingMatcher) NegatedFailureMessage(_ interface{}) string {
	return matcher.expectedMatcher.NegatedFailureMessage(matcher.contents)
}
