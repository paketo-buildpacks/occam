package matchers_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitMatchers(t *testing.T) {
	suite := spec.New("occam/matchers", spec.Report(report.Terminal{}))
	suite("BeAvailable", testBeAvailable)
	suite("ContainLines", testContainLines)
	suite("Serve", testServe)
	suite("BeAFileMatching", testBeAFileMatching)
	suite.Run(t)
}
