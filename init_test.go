package occam_test

import (
	"testing"

	"github.com/onsi/gomega/format"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitOccam(t *testing.T) {
	format.MaxLength = 0

	suite := spec.New("occam", spec.Report(report.Terminal{}))
	suite("CacheVolumeNames", testCacheVolumeNames)
	suite("Container", testContainer)
	suite("Docker", testDocker)
	suite("Image", testImage)
	suite("Pack", testPack)
	suite("RandomName", testRandomName)
	suite("Source", testSource)
	suite("BuildpackStore", testBuildpackStore)
	suite("ContainerStructureTest", testContainerStructureTest)
	suite("Venom", testVenom)
	suite("TestContainers", testTestContainers)
	suite.Run(t)
}
