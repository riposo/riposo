package slowhash_test

import (
	"testing"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/pkg/slowhash"
)

var _ = Describe("Generator", func() {
	It("supports bcrypt", func() {
		var subject Generator = BCrypt

		hashed, err := subject("s3cret")
		Expect(err).NotTo(HaveOccurred())
		Expect(hashed).To(HavePrefix("$2a$12$"))
		Expect(len(hashed)).To(Equal(60))
		Expect(hashed).To(HaveLen(60))
		Expect(Verify(hashed, "s3cret")).To(BeTrue())
		Expect(Verify(hashed, "nomatch")).To(BeFalse())
	})

	It("supports argon2id", func() {
		var subject Generator = Argon2ID

		hashed, err := subject("s3cret")
		Expect(err).NotTo(HaveOccurred())
		Expect(hashed).To(HavePrefix("$argon2id$v=19$m=65536,t=1,p=2$"))
		Expect(hashed).To(HaveLen(97))
		Expect(Verify(hashed, "s3cret")).To(BeTrue())
		Expect(Verify(hashed, "nomatch")).To(BeFalse())
	})

	It("verifies bcrypt 2b", func() {
		Expect(Verify("$2b$12$FveWzQHevRG15avGQHVF0OcpM9kqwtp.84TeOvxM5Wh8JRrI5RmJK", "s3cret")).To(BeTrue())
	})
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "pkg/slowhash")
}
