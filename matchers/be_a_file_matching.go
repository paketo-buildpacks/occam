package matchers

import (
	"fmt"
	"os"

	"github.com/onsi/gomega/types"
)

type BeAFileMatchingMatcher struct {
	expectedMatcher types.GomegaMatcher
	contents        string
}

func BeAFileMatching(expectedMatcher types.GomegaMatcher) *BeAFileMatchingMatcher {
	return &BeAFileMatchingMatcher{
		expectedMatcher: expectedMatcher,
	}
}

// Match will return whether the expectedMatcher matches on the file contents.
// The file contents are saved for use in FailureMessage and NegatedFailureMessage.
func (matcher *BeAFileMatchingMatcher) Match(actual interface{}) (success bool, err error) {
	actualFilename, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("BeAFileMatchingMatcher expects a file path")
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
