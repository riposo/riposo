package schema_test

import (
	"testing"

	"github.com/riposo/riposo/pkg/schema"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
)

var _ = Describe("PermissionSet", func() {
	var subject schema.PermissionSet

	BeforeEach(func() {
		subject = schema.PermissionSet{
			"read":  {"alice", "bob"},
			"write": {"alice", "claire"},
		}
	})

	It("adds", func() {
		subject.Add("read", "alice")
		subject.Add("create", "alice")
		subject.Add("write", "bob")

		Expect(subject).To(Equal(schema.PermissionSet{
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
