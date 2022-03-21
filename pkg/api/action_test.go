package api_test

import (
	"github.com/riposo/riposo/pkg/mock"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"

	. "github.com/bsm/ginkgo/v2"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/pkg/api"
)

var _ = Describe("Actions", func() {
	var subject Actions
	var txn *Txn
	var cbs *mockCallbacks

	BeforeEach(func() {
		txn = mock.Txn()
		txn.User = mock.User("alice")

		cbs = new(mockCallbacks)
		subject = NewActions(DefaultModel{}, []Callbacks{cbs})

		// seed one
		Expect(subject.Create(txn, "/objects/*", &schema.Resource{
			Data: &schema.Object{Extra: []byte(`{"meta":true}`)},
		})).To(Succeed())
		cbs.Reset()
	})

	AfterEach(func() {
		Expect(txn.Rollback()).To(Succeed())
	})

	It("gets", func() {
		Expect(subject.Get(txn, "/objects/EPR.ID")).To(Equal(&schema.Resource{
			Data: &schema.Object{
				ID:      "EPR.ID",
				ModTime: 1515151515677,
				Extra:   []byte(`{"meta":true}`),
			},
			Permissions: schema.PermissionSet{"write": {"alice"}},
		}))
	})

	It("creates with callbacks", func() {
		obj := &schema.Object{Extra: []byte(`{"created":true}`)}
		Expect(subject.Create(txn, "/objects/*", &schema.Resource{Data: obj})).To(Succeed())
		Expect(obj).To(Equal(&schema.Object{
			ID:      "ITR.ID",
			ModTime: 1515151515678,
			Extra:   []byte(`{"created":true}`),
		}))
		Expect(cbs.Calls).To(Equal([]string{"BeforeCreate", "AfterCreate"}))
		Expect(cbs.Paths).To(Equal([]string{"/objects/*"}))
	})

	It("updates with callbacks", func() {
		exst, err := txn.Store.GetForUpdate("/objects/EPR.ID")
		Expect(err).NotTo(HaveOccurred())

		Expect(subject.Update(txn, "/objects/EPR.ID", exst, &schema.Resource{
			Data: &schema.Object{Extra: []byte(`{"updated":true}`)},
		})).To(Equal(&schema.Resource{
			Data: &schema.Object{
				ID:      "EPR.ID",
				ModTime: 1515151515678,
				Extra:   []byte(`{"updated":true}`),
			},
			Permissions: schema.PermissionSet{"write": {"alice"}},
		}))
		Expect(cbs.Calls).To(Equal([]string{"BeforeUpdate", "AfterUpdate"}))
		Expect(cbs.Paths).To(Equal([]string{"/objects/EPR.ID"}))
	})

	It("patches with callbacks", func() {
		exst, err := txn.Store.GetForUpdate("/objects/EPR.ID")
		Expect(err).NotTo(HaveOccurred())

		Expect(subject.Patch(txn, "/objects/EPR.ID", exst, &schema.Resource{
			Data: &schema.Object{Extra: []byte(`{"patched":true}`)},
		})).To(Equal(&schema.Resource{
			Data: &schema.Object{
				ID:      "EPR.ID",
				ModTime: 1515151515678,
				Extra:   []byte(`{"meta":true,"patched":true}`),
			},
			Permissions: schema.PermissionSet{"write": {"alice"}},
		}))
		Expect(cbs.Calls).To(Equal([]string{"BeforePatch", "AfterPatch"}))
		Expect(cbs.Paths).To(Equal([]string{"/objects/EPR.ID"}))
	})

	It("deletes with callbacks", func() {
		exst, err := txn.Store.GetForUpdate("/objects/EPR.ID")
		Expect(err).NotTo(HaveOccurred())

		Expect(subject.Delete(txn, "/objects/EPR.ID", exst)).To(Equal(&schema.Object{
			ID:      "EPR.ID",
			ModTime: 1515151515678,
			Deleted: true,
			Extra:   []byte(`{"meta":true}`),
		}))
		Expect(cbs.Calls).To(Equal([]string{"BeforeDelete", "AfterDelete"}))
		Expect(cbs.Paths).To(Equal([]string{"/objects/EPR.ID"}))
	})

	It("deletes all with callbacks", func() {
		obj, err := txn.Store.Get("/objects/EPR.ID")
		Expect(err).NotTo(HaveOccurred())

		Expect(subject.DeleteAll(txn, "/objects/*", []*schema.Object{obj})).To(Equal(riposo.Epoch(1515151515678)))
		Expect(cbs.Calls).To(Equal([]string{"BeforeDeleteAll", "AfterDeleteAll"}))
		Expect(cbs.Paths).To(Equal([]string{"/objects/*"}))
	})
})
