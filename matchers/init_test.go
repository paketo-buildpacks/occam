package matchers_test

import (
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

func TestUnitMatchers(t *testing.T) {
	var Expect = NewWithT(t).Expect

	docker := occam.NewDocker()
	err := docker.Pull.Execute("alpine:latest")
	Expect(err).NotTo(HaveOccurred())

	suite := spec.New("occam/matchers", spec.Report(report.Terminal{}))
	suite("BeAvailable", testBeAvailable)
	suite("ContainLines", testContainLines)
	suite("Serve", testServe)
	suite("BeAFileMatching", testBeAFileMatching)
	suite("HaveDirectory", testHaveDirectory)
	suite("HaveFile", testHaveFile)
	suite("HaveFileWithContent", testHaveFileWithContent)
	suite.Run(t)

	err = docker.Image.Remove.WithForce().Execute("alpine:latest")
	Expect(err).NotTo(HaveOccurred())
}
