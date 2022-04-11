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

	context("IPAddressForNetwork", func() {
		it("returns the IP Address associated ", func() {
			container := occam.Container{
				IPAddresses: map[string]string{
					"bridge": "10.172.0.2",
				},
			}
			Expect(container.IPAddressForNetwork("bridge")).To(Equal("10.172.0.2"))
		})

		context("failure cases", func() {
			context("when the provided network does not exist", func() {
				it("returns an error", func() {
					container := occam.Container{}

					_, err := container.IPAddressForNetwork("some-non-existent-network")
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
}
