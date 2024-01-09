package occam_test

import (
	"testing"

	occam "github.com/paketo-buildpacks/occam"
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

	context("NewContainerFromInspectOutput", func() {
		it("It creates a new container from inspect output ", func() {

			output := []byte(`[
			{
				"Id": "container-id",
				"Config": {
					"Env": ["ENV_VAR1=value1", "ENV_VAR2=value2"]
				},
				"NetworkSettings": {
					"Ports": {
						"8080/tcp": [
							{
								"HostPort": "1234"
							}
						]
					},
					"Networks": {
						"network1": {
							"IPAddress": "192.168.0.1"
						},
						"network2": {
							"IPAddress": "192.168.0.2"
						}
					}
				}
			}
		]`)

			container, err := occam.NewContainerFromInspectOutput(output)
			Expect(err).NotTo(HaveOccurred())
			Expect(container.ID).To(Equal("container-id"))
			Expect(container.Ports).To(HaveLen(1))
			Expect(container.Ports).To(HaveKeyWithValue("8080", "1234"))
			Expect(container.Env).To(HaveLen(2))
			Expect(container.Env).To(HaveKeyWithValue("ENV_VAR1", "value1"))
			Expect(container.Env).To(HaveKeyWithValue("ENV_VAR2", "value2"))
			Expect(container.IPAddresses).To(HaveLen(2))
			Expect(container.IPAddresses["network1"]).To(Equal("192.168.0.1"))
			Expect(container.IPAddresses["network2"]).To(Equal("192.168.0.2"))
		})

		context("When there are no host ports but only container ports", func() {
			it("It creates a new container without the exposed ports included ", func() {

				output := []byte(`[
				{
					"Id": "container-id",
					"Config": {
						"Env": ["ENV_VAR1=value1", "ENV_VAR2=value2"]
					},
					"NetworkSettings": {
						"Ports": {
							"8080/tcp": []
						},
						"Networks": {
							"network1": {
								"IPAddress": "192.168.0.1"
							}
						}
					}
				}
			]`)

				container, err := occam.NewContainerFromInspectOutput(output)
				Expect(err).NotTo(HaveOccurred())
				Expect(container.ID).To(Equal("container-id"))
				Expect(container.Ports).To(HaveLen(0))
				Expect(container.Env).To(HaveLen(2))
				Expect(container.Env).To(HaveKeyWithValue("ENV_VAR1", "value1"))
				Expect(container.Env).To(HaveKeyWithValue("ENV_VAR2", "value2"))
				Expect(container.IPAddresses).To(HaveLen(1))
				Expect(container.IPAddresses["network1"]).To(Equal("192.168.0.1"))
			})
		})
	})

}
