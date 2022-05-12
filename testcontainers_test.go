package occam_test

import (
	ctx "context"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func testTestContainers(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		testContainer occam.TestContainers
		container     testcontainers.Container
		err           error
		testImage     = "nginx:stable-alpine"
	)

	it.Before(func() {
		testContainer = occam.NewTestContainers()
	})

	it.After(func() {
		err := container.Terminate(ctx.Background())
		Expect(err).ToNot(HaveOccurred())
	})

	context("TestContainers", func() {
		it("should work", func() {
			container, err = testContainer.WithWaitingFor(wait.ForAll(wait.ForLog("start worker process"), wait.ForListeningPort("80/tcp"))).
				WithExposedPorts("80/tcp").
				WithTimeout(10).
				Execute(testImage)
			Expect(err).NotTo(HaveOccurred())
		})
	})
}
