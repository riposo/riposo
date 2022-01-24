package group_test

import (
	"testing"

	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/conn/storage"
	"github.com/riposo/riposo/pkg/mock"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"

	. "github.com/bsm/ginkgo/v2"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/internal/model/group"
)

var _ = Describe("Group Model", func() {
	var subject api.Model
	var txn *api.Txn

	BeforeEach(func() {
		txn = mock.Txn()
		subject = Model{}
	})

	AfterEach(func() {
		Expect(txn.Abort()).To(Succeed())
	})

	Describe("Create", func() {
		It("normalizes", func() {
			obj := &schema.Object{Extra: []byte(`{"members":["bob",null,"alice","alice"]}`)}
			Expect(subject.Create(txn, "/buckets/foo/groups/*", &schema.Resource{Data: obj})).To(Succeed())
			Expect(obj.Extra).To(MatchJSON(`{"members":["alice","bob"]}`))

			obj = &schema.Object{Extra: []byte(`{}`)}
			Expect(subject.Create(txn, "/buckets/foo/groups/*", &schema.Resource{Data: obj})).To(Succeed())
			Expect(obj.Extra).To(MatchJSON(`{"members":[]}`))
		})

		It("appends principal to accounts", func() {
			Expect(subject.Create(txn, "/buckets/foo/groups/*", &schema.Resource{
				Data: &schema.Object{Extra: []byte(`{"members":["alice","bob"]}`)},
			})).To(Succeed())
			Expect(txn.Perms.GetUserPrincipals("alice")).To(ContainElement("/buckets/foo/groups/EPR.ID"))
			Expect(txn.Perms.GetUserPrincipals("bob")).To(ContainElement("/buckets/foo/groups/EPR.ID"))
		})
	})

	Describe("Update", func() {
		var hs storage.UpdateHandle

		BeforeEach(func() {
			obj := &schema.Object{Extra: []byte(`{"members":["alice","bob"]}`)}
			Expect(subject.Create(txn, "/buckets/foo/groups/*", &schema.Resource{Data: obj})).To(Succeed())

			var err error
			hs, err = txn.Store.GetForUpdate("/buckets/foo/groups/EPR.ID")
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not require members", func() {
			Expect(subject.Update(txn, "/buckets/foo/groups/EPR.ID", hs, &schema.Resource{
				Data: &schema.Object{Extra: []byte(`{}`)},
			})).To(Succeed())
			Expect(hs.Object().Extra).To(MatchJSON(`{"members":[]}`))
		})

		It("normalizes", func() {
			Expect(subject.Update(txn, "/buckets/foo/groups/EPR.ID", hs, &schema.Resource{
				Data: &schema.Object{Extra: []byte(`{"members":["claire","alice","","alice"]}`)},
			})).To(Succeed())
			Expect(hs.Object().Extra).To(MatchJSON(`{"members":["alice","claire"]}`))
		})

		It("updates principals", func() {
			Expect(subject.Update(txn, "/buckets/foo/groups/EPR.ID", hs, &schema.Resource{
				Data: &schema.Object{Extra: []byte(`{"members":["alice","claire"]}`)},
			})).To(Succeed())
			Expect(txn.Perms.GetUserPrincipals("alice")).To(ContainElement("/buckets/foo/groups/EPR.ID"))
			Expect(txn.Perms.GetUserPrincipals("bob")).NotTo(ContainElement("/buckets/foo/groups/EPR.ID"))
			Expect(txn.Perms.GetUserPrincipals("claire")).To(ContainElement("/buckets/foo/groups/EPR.ID"))
		})
	})

	Describe("Patch", func() {
		var hs storage.UpdateHandle

		BeforeEach(func() {
			obj := &schema.Object{Extra: []byte(`{"members":["alice","bob"]}`)}
			Expect(subject.Create(txn, "/buckets/foo/groups/*", &schema.Resource{Data: obj})).To(Succeed())

			var err error
			hs, err = txn.Store.GetForUpdate("/buckets/foo/groups/EPR.ID")
			Expect(err).NotTo(HaveOccurred())
		})

		It("allows to remain unchanged", func() {
			Expect(subject.Patch(txn, "/buckets/foo/groups/EPR.ID", hs, &schema.Resource{
				Data: &schema.Object{Extra: []byte(`{}`)},
			})).To(Succeed())
			Expect(hs.Object().Extra).To(MatchJSON(`{"members":["alice","bob"]}`))
		})

		It("normalizes", func() {
			Expect(subject.Patch(txn, "/buckets/foo/groups/EPR.ID", hs, &schema.Resource{
				Data: &schema.Object{Extra: []byte(`{"members":["claire","alice",null,"alice"]}`)},
			})).To(Succeed())
			Expect(hs.Object().Extra).To(MatchJSON(`{"members":["alice","claire"]}`))
		})

		It("updates principals", func() {
			Expect(subject.Patch(txn, "/buckets/foo/groups/EPR.ID", hs, &schema.Resource{
				Data: &schema.Object{Extra: []byte(`{"members":["alice","claire"]}`)},
			})).To(Succeed())
			Expect(txn.Perms.GetUserPrincipals("alice")).To(ContainElement("/buckets/foo/groups/EPR.ID"))
			Expect(txn.Perms.GetUserPrincipals("bob")).NotTo(ContainElement("/buckets/foo/groups/EPR.ID"))
			Expect(txn.Perms.GetUserPrincipals("claire")).To(ContainElement("/buckets/foo/groups/EPR.ID"))
		})
	})

	Describe("Delete", func() {
		var obj *schema.Object

		BeforeEach(func() {
			obj = &schema.Object{Extra: []byte(`{"members":["alice","bob"]}`)}
			Expect(subject.Create(txn, "/buckets/foo/groups/*", &schema.Resource{Data: obj})).To(Succeed())
		})

		It("updates principals", func() {
			Expect(subject.Delete(txn, "/buckets/foo/groups/EPR.ID", obj)).To(BeAssignableToTypeOf(&schema.Object{}))
			Expect(txn.Perms.GetUserPrincipals("alice")).NotTo(ContainElement("/buckets/foo/groups/EPR.ID"))
			Expect(txn.Perms.GetUserPrincipals("bob")).NotTo(ContainElement("/buckets/foo/groups/EPR.ID"))
		})
	})

	Describe("DeleteAll", func() {
		BeforeEach(func() {
			o1 := &schema.Object{Extra: []byte(`{"members":["alice","bob"]}`)}
			o2 := &schema.Object{Extra: []byte(`{"members":["alice"]}`)}
			Expect(subject.Create(txn, "/buckets/foo/groups/*", &schema.Resource{Data: o1})).To(Succeed())
			Expect(subject.Create(txn, "/buckets/foo/groups/*", &schema.Resource{Data: o2})).To(Succeed())
		})

		It("updates principals", func() {
			Expect(txn.Perms.GetUserPrincipals("alice")).To(ConsistOf(
				"/buckets/foo/groups/EPR.ID",
				"/buckets/foo/groups/ITR.ID",
				"alice",
				"system.Authenticated",
				"system.Everyone",
			))

			modTime, deleted, err := subject.DeleteAll(txn, "/buckets/foo/groups/*", []string{
				"EPR.ID",
				"ITR.ID",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(modTime).To(Equal(riposo.Epoch(1515151515680)))
			Expect(deleted).To(HaveLen(2))

			Expect(txn.Perms.GetUserPrincipals("alice")).To(ConsistOf(
				"alice",
				"system.Authenticated",
				"system.Everyone",
			))
		})
	})
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "internal/model/group")
}
