package identity_test

import (
	"strings"
	"testing"

	"github.com/riposo/riposo/pkg/identity"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NanoID", func() {
	It("generates IDs", func() {
		fn := identity.NanoID
		Expect(fn()).To(HaveLen(20))
		Expect(fn()).To(HaveLen(20))
		Expect(fn()).NotTo(Equal(fn()))
	})
})

var _ = Describe("UUID", func() {
	It("generates IDs", func() {
		fn := identity.UUID
		Expect(fn()).To(HaveLen(36))
		Expect(fn()).To(HaveLen(36))
		Expect(fn()).NotTo(Equal(fn()))
	})
})

var _ = Describe("IsValid", func() {
	It("validates IDs", func() {
		Expect(identity.IsValid(identity.NanoID())).To(BeTrue())
		Expect(identity.IsValid(identity.UUID())).To(BeTrue())
		Expect(identity.IsValid("123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ")).To(BeTrue()) // base58 alphabet
		Expect(identity.IsValid("with_underscore")).To(BeTrue())
		Expect(identity.IsValid("with-hyphen")).To(BeTrue())
		Expect(identity.IsValid("email@address.net")).To(BeTrue())

		Expect(identity.IsValid("with space")).To(BeFalse())
		Expect(identity.IsValid("with/slash")).To(BeFalse())
		Expect(identity.IsValid("with,comma")).To(BeFalse())
		Expect(identity.IsValid("")).To(BeFalse())
		Expect(identity.IsValid("日本")).To(BeFalse())
	})

	It("limits length", func() {
		borderline := strings.Repeat("x", 254)
		Expect(identity.IsValid(borderline)).To(BeTrue())
		Expect(identity.IsValid(borderline + "x")).To(BeTrue())
		Expect(identity.IsValid(borderline + "xx")).To(BeFalse())
	})
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "pkg/identity")
}
