package packagers_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPackagers(t *testing.T) {
	suite := spec.New("packagers", spec.Report(report.Terminal{}))
	suite("Libpak", testLibpak)
	suite("Jam", testJam)
	suite.Run(t)
}
