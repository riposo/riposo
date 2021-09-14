package schema_test

import (
	"testing"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/pkg/schema"
)

var _ = Describe("PermissionSet", func() {
	var subject PermissionSet

	BeforeEach(func() {
		subject = PermissionSet{
			"read":  {"alice", "bob"},
			"write": {"alice", "claire"},
		}
	})

	It("adds", func() {
		subject.Add("read", "alice")
		subject.Add("create", "alice")
		subject.Add("write", "bob")

		Expect(subject).To(Equal(PermissionSet{
			"create": {"alice"},
			"read":   {"alice", "bob"},
			"write":  {"alice", "bob", "claire"},
		}))
	})
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "pkg/schema")
}
