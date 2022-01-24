package identity_test

import (
	"strings"
	"testing"

	. "github.com/bsm/ginkgo/v2"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/pkg/identity"
)

var _ = Describe("NanoID", func() {
	It("generates IDs", func() {
		fn := NanoID
		Expect(fn()).To(HaveLen(20))
		Expect(fn()).To(HaveLen(20))
		Expect(fn()).NotTo(Equal(fn()))
	})
})

var _ = Describe("UUID", func() {
	It("generates IDs", func() {
		fn := UUID
		Expect(fn()).To(HaveLen(36))
		Expect(fn()).To(HaveLen(36))
		Expect(fn()).NotTo(Equal(fn()))
	})
})

var _ = Describe("IsValid", func() {
	It("validates IDs", func() {
		Expect(IsValid(NanoID())).To(BeTrue())
		Expect(IsValid(UUID())).To(BeTrue())
		Expect(IsValid("123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ")).To(BeTrue()) // base58 alphabet
		Expect(IsValid("with_underscore")).To(BeTrue())
		Expect(IsValid("with-hyphen")).To(BeTrue())
		Expect(IsValid("email@address.net")).To(BeTrue())

		Expect(IsValid("with space")).To(BeFalse())
		Expect(IsValid("with/slash")).To(BeFalse())
		Expect(IsValid("with,comma")).To(BeFalse())
		Expect(IsValid("")).To(BeFalse())
		Expect(IsValid("日本")).To(BeFalse())
	})

	It("limits length", func() {
		borderline := strings.Repeat("x", 254)
		Expect(IsValid(borderline)).To(BeTrue())
		Expect(IsValid(borderline + "x")).To(BeTrue())
		Expect(IsValid(borderline + "xx")).To(BeFalse())
	})
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "pkg/identity")
}
