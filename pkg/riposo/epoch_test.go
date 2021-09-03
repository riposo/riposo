package riposo_test

import (
	"time"

	"github.com/riposo/riposo/pkg/riposo"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
)

var _ = Describe("Epoch", func() {
	It("converts from time", func() {
		Expect(riposo.EpochFromTime(time.Unix(1567815678, 987456789))).To(Equal(riposo.Epoch(1567815678987)))
		Expect(riposo.EpochFromTime(time.Unix(1567815678, 987567890))).To(Equal(riposo.Epoch(1567815678988)))
	})

	It("converts to time", func() {
		Expect(riposo.Epoch(1567815678987).Time()).To(BeTemporally("==", time.Unix(1567815678, 987000000)))
		Expect(riposo.Epoch(1567815678988).Time()).To(BeTemporally("==", time.Unix(1567815678, 988000000)))
	})

	It("converts to strong ETag", func() {
		Expect(riposo.Epoch(1567815678988).ETag()).To(Equal(`"1567815678988"`))
	})

	It("formats for HTTP", func() {
		Expect(riposo.Epoch(1567815678988).HTTPFormat()).To(Equal("Sat, 07 Sep 2019 00:21:18 GMT"))
	})
})
