package matchers_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitMatchers(t *testing.T) {
	suite := spec.New("occam/matchers", spec.Report(report.Terminal{}))
	suite("ContainLines", testContainLines)
	suite.Run(t)
}
