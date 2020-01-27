package occam_test

import (
	"testing"

	"github.com/cloudfoundry/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testRandomName(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	it("generates a random name", func() {
		name, err := occam.RandomName()
		Expect(err).NotTo(HaveOccurred())
		Expect(name).To(MatchRegexp(`^occam\-[0123456789abcdefghjkmnpqrstvwxyz]{26}$`))
	})
}
