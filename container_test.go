package occam_test

import (
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testContainer(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("HostPort", func() {
		it("returns the external port the container is bound to", func() {
			container := occam.Container{
				Ports: map[string]string{
					"1234": "11111",
				},
			}
			Expect(container.HostPort("1234")).To(Equal("11111"))
		})
	})
}
