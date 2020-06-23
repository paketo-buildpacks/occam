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
			"pack-cache-16fe664c76f0.launch",
			"pack-cache-16fe664c76f0.cache",
		}))
	})
}
