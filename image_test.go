package occam_test

import (
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testImage(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("BuildpackForKey", func() {
		it("returns the Buildpack with the key", func() {
			image := occam.Image{
				Buildpacks: []occam.ImageBuildpackMetadata{
					{
						Key: "first",
						Layers: map[string]occam.ImageBuildpackMetadataLayer{
							"first-layer-1": {
								SHA: "first-layer-1-sha",
							},
							"first-layer-2": {
								SHA: "first-layer-2-sha",
							},
						},
					},
					{
						Key: "second",
						Layers: map[string]occam.ImageBuildpackMetadataLayer{
							"second-layer-1": {
								SHA: "second-layer-1-sha",
							},
							"second-layer-2": {
								SHA: "second-layer-2-sha",
							},
						},
					},
				},
			}

			firstBuildpack, err := image.BuildpackForKey("first")
			Expect(err).NotTo(HaveOccurred())

			Expect(firstBuildpack.Key).To(Equal("first"))
			Expect(firstBuildpack.Layers["first-layer-2"].SHA).To(Equal("first-layer-2-sha"))

			secondBuildpack, err := image.BuildpackForKey("second")
			Expect(err).NotTo(HaveOccurred())

			Expect(secondBuildpack.Key).To(Equal("second"))
			Expect(secondBuildpack.Layers["second-layer-2"].SHA).To(Equal("second-layer-2-sha"))
		})

		context("failure cases", func() {
			context("when no buildpack exists with the provided key", func() {
				it("returns an error", func() {
					image := occam.Image{}

					_, err := image.BuildpackForKey("some-non-existent-key")
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
}
