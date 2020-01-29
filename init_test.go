package occam_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitOccam(t *testing.T) {
	suite := spec.New("occam", spec.Report(report.Terminal{}))
	suite("CacheVolumeNames", testCacheVolumeNames)
	suite("Container", testContainer)
	suite("Docker", testDocker)
	suite("Pack", testPack)
	suite("RandomName", testRandomName)
	suite.Run(t)
}
