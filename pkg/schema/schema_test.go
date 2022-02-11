package schema_test

import (
	"encoding/json"
	"testing"

	. "github.com/bsm/ginkgo/v2"
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

	DescribeTable("calculates size",
		func(ps PermissionSet) {
			bin, _ := json.Marshal(ps)
			Expect(ps.ByteSize()).To(Equal(int64(len(bin))))
		},
		Entry("null", (PermissionSet)(nil)),
		Entry("blank", PermissionSet{}),
		Entry("one key, no values", PermissionSet{"read": {}}),
		Entry("one key, one value", PermissionSet{"read": {"alice"}}),
		Entry("one key, multiple values", PermissionSet{"read": {"alice", "bob"}}),
		Entry("multiple keys, multiple values", PermissionSet{"read": {"alice", "bob", "claire"}, "write": {"alice", "claire"}}),
	)
})

var _ = Describe("Resource", func() {
	DescribeTable("calculates size",
		func(r *Resource) {
			bin, _ := json.Marshal(r)
			Expect(r.ByteSize()).To(Equal(int64(len(bin))))
		},
		Entry("null", (*Resource)(nil)),
		Entry("blank", &Resource{}),
		Entry("only data", &Resource{Data: &Object{ID: "EPR.ID"}}),
		Entry("only permissions", &Resource{Permissions: PermissionSet{"read": {"alice"}}}),
		Entry("full example", &Resource{Data: &Object{ID: "EPR.ID"}, Permissions: PermissionSet{"read": {"alice"}}}),
	)
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "pkg/schema")
}
