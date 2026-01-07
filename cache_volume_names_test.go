package occam_test

import (
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testCacheVolumeNames(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	it("returns the name of the cache volumes that are assigned to an image", func() {
		Expect(occam.CacheVolumeNames("some-app")).To(Equal([]string{
			"pack-cache-16fe664c76f0.build",
			"pack-cache-some-app_latest-16fe664c76f0.build",
			"pack-cache-16fe664c76f0.launch",
			"pack-cache-some-app_latest-16fe664c76f0.launch",
			"pack-cache-16fe664c76f0.cache",
			"pack-cache-some-app_latest-16fe664c76f0.cache",
			"pack-cache-some-app_latest-c38cb104abe0.kaniko",
		}))
	})

	context("when the name includes a registry", func() {
		it("returns the name of the cache volumes that are assigned to an image", func() {
			Expect(occam.CacheVolumeNames("occam.example.com/some-app")).To(Equal([]string{
				"pack-cache-e69d3a4f1e12.build",
				"pack-cache-some-app_latest-e69d3a4f1e12.build",
				"pack-cache-e69d3a4f1e12.launch",
				"pack-cache-some-app_latest-e69d3a4f1e12.launch",
				"pack-cache-e69d3a4f1e12.cache",
				"pack-cache-some-app_latest-e69d3a4f1e12.cache",
				"pack-cache-some-app_latest-03c3cd5a5c5d.kaniko",
			}))
		})
	})
}
