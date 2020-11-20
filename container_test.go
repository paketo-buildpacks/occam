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

	context("ContainerPorts", func() {
		it("returns the container's exposed ports", func() {
			container := occam.Container{
				Ports: map[string]string{
					"1234": "11111",
					"2345": "2222",
					"3456": "22222",
				},
			}
			Expect(container.ContainerPorts()).To(ContainElement("1234"))
			Expect(container.ContainerPorts()).To(ContainElement("2345"))
			Expect(container.ContainerPorts()).To(ContainElement("3456"))
		})
	})
}
