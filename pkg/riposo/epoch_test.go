package riposo_test

import (
	"time"

	. "github.com/bsm/ginkgo/v2"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/pkg/riposo"
)

var _ = Describe("Epoch", func() {
	It("converts from time", func() {
		Expect(EpochFromTime(time.Unix(1567815678, 987456789))).To(Equal(Epoch(1567815678987)))
		Expect(EpochFromTime(time.Unix(1567815678, 987567890))).To(Equal(Epoch(1567815678988)))
	})

	It("converts to time", func() {
		Expect(Epoch(1567815678987).Time()).To(BeTemporally("==", time.Unix(1567815678, 987000000)))
		Expect(Epoch(1567815678988).Time()).To(BeTemporally("==", time.Unix(1567815678, 988000000)))
	})

	It("converts to strong ETag", func() {
		Expect(Epoch(1567815678988).ETag()).To(Equal(`"1567815678988"`))
	})

	It("formats for HTTP", func() {
		Expect(Epoch(1567815678988).HTTPFormat()).To(Equal("Sat, 07 Sep 2019 00:21:18 GMT"))
	})
})
