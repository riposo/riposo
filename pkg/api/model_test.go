package api_test

import (
	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/conn/storage"
	"github.com/riposo/riposo/pkg/mock"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Model", func() {
	var subject api.Model
	var txn *api.Txn

	resource := func() *schema.Resource {
		return &schema.Resource{
			Data:        &schema.Object{Extra: []byte(`{"meta":true}`)},
			Permissions: schema.PermissionSet{"write": {"alice"}},
		}
	}

	BeforeEach(func() {
		txn = mock.Txn()
		subject = api.StdModel()
	})

	AfterEach(func() {
		Expect(txn.Abort()).To(Succeed())
	})

	Describe("Get", func() {
		It("gets", func() {
			payload := resource()
			Expect(subject.Create(txn, "/objects/*", payload)).To(Succeed())
			Expect(subject.Get(txn, "/objects/EPR.ID")).To(Equal(payload))
		})
	})

	Describe("Create", func() {
		It("creates without permissions", func() {
			obj := &schema.Object{Extra: []byte(`{"meta":true}`)}
			Expect(subject.Create(txn, "/objects/*", &schema.Resource{Data: obj})).To(Succeed())
			Expect(obj).To(Equal(&schema.Object{
				ID:      "EPR.ID",
				ModTime: 1515151515677,
				Extra:   []byte(`{"meta":true}`),
			}))
			Expect(txn.Store.Get("/objects/EPR.ID")).To(Equal(obj))
			Expect(txn.Perms.GetPermissions("/objects/EPR.ID")).To(BeEmpty())
		})

		It("creates with permissions", func() {
			obj := &schema.Object{Extra: []byte(`{"meta":true}`)}
			pms := schema.PermissionSet{
				"write": {"alice", "claire"},
				"read":  {"bob"},
			}
			Expect(subject.Create(txn, "/objects/*", &schema.Resource{
				Data:        obj,
				Permissions: pms,
			})).To(Succeed())
			Expect(obj).To(Equal(&schema.Object{
				ID:      "EPR.ID",
				ModTime: 1515151515677,
				Extra:   []byte(`{"meta":true}`),
			}))
			Expect(txn.Store.Get("/objects/EPR.ID")).To(Equal(obj))
			Expect(txn.Perms.GetPermissions("/objects/EPR.ID")).To(Equal(pms))
		})
	})

	Describe("Update", func() {
		var hs storage.UpdateHandle

		BeforeEach(func() {
			Expect(subject.Create(txn, "/objects/*", resource())).To(Succeed())

			var err error
			hs, err = txn.Store.GetForUpdate("/objects/EPR.ID")
			Expect(err).NotTo(HaveOccurred())
		})

		It("updates without permissions", func() {
			Expect(subject.Update(txn, "/objects/EPR.ID", hs, &schema.Resource{
				Data: &schema.Object{Extra: []byte(`{"updated":true}`)},
			})).To(Succeed())

			Expect(hs.Object()).To(Equal(&schema.Object{
				ID:      "EPR.ID",
				ModTime: 1515151515678,
				Extra:   []byte(`{"updated":true}`),
			}))
			Expect(txn.Store.Get("/objects/EPR.ID")).To(Equal(hs.Object()))
			Expect(txn.Perms.GetPermissions("/objects/EPR.ID")).To(Equal(schema.PermissionSet{
				"write": {"alice"},
			}))
		})

		It("updates with permissions", func() {
			pms := schema.PermissionSet{"write": {"alice"}, "read": {"bob"}}
			Expect(subject.Update(txn, "/objects/EPR.ID", hs, &schema.Resource{
				Data:        &schema.Object{Extra: []byte(`{"updated":true}`)},
				Permissions: pms,
			})).To(Succeed())

			Expect(hs.Object()).To(Equal(&schema.Object{
				ID:      "EPR.ID",
				ModTime: 1515151515678,
				Extra:   []byte(`{"updated":true}`),
			}))
			Expect(txn.Store.Get("/objects/EPR.ID")).To(Equal(hs.Object()))
			Expect(txn.Perms.GetPermissions("/objects/EPR.ID")).To(Equal(pms))
		})
	})

	Describe("Patch", func() {
		var hs storage.UpdateHandle

		BeforeEach(func() {
			payload := resource()
			payload.Data.Extra = []byte(`{"a":1,"b":2}`)
			Expect(subject.Create(txn, "/objects/*", payload)).To(Succeed())

			var err error
			hs, err = txn.Store.GetForUpdate("/objects/EPR.ID")
			Expect(err).NotTo(HaveOccurred())
		})

		It("patches without permissions", func() {
			Expect(subject.Patch(txn, "/objects/EPR.ID", hs, &schema.Resource{
				Data: &schema.Object{Extra: []byte(`{"b":4,"c":3}`)},
			})).To(Succeed())

			Expect(hs.Object()).To(Equal(&schema.Object{
				ID:      "EPR.ID",
				ModTime: 1515151515678,
				Extra:   []byte(`{"a":1,"b":4,"c":3}`),
			}))
			Expect(txn.Store.Get("/objects/EPR.ID")).To(Equal(hs.Object()))
			Expect(txn.Perms.GetPermissions("/objects/EPR.ID")).To(Equal(schema.PermissionSet{
				"write": {"alice"},
			}))
		})

		It("patches with permissions", func() {
			pms := schema.PermissionSet{"write": {"alice"}, "read": {"bob"}}
			Expect(subject.Patch(txn, "/objects/EPR.ID", hs, &schema.Resource{
				Data:        &schema.Object{Extra: []byte(`{"b":4,"c":3}`)},
				Permissions: pms,
			})).To(Succeed())

			Expect(hs.Object()).To(Equal(&schema.Object{
				ID:      "EPR.ID",
				ModTime: 1515151515678,
				Extra:   []byte(`{"a":1,"b":4,"c":3}`),
			}))
			Expect(txn.Store.Get("/objects/EPR.ID")).To(Equal(hs.Object()))
			Expect(txn.Perms.GetPermissions("/objects/EPR.ID")).To(Equal(pms))
		})
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			nested := api.StdModel()

			// seed /objects/EPR.ID
			Expect(subject.Create(txn, "/objects/*", resource())).To(Succeed())

			// seed /objects/EPR.ID/nested/ITR.ID
			Expect(nested.Create(txn, "/objects/EPR.ID/nested/*", resource())).To(Succeed())
		})

		It("deletes the object", func() {
			Expect(txn.Store.Get("/objects/EPR.ID")).NotTo(BeNil())
			Expect(txn.Perms.GetPermissions("/objects/EPR.ID")).To(HaveLen(1))

			Expect(subject.Delete(txn, "/objects/EPR.ID")).To(Equal(&schema.Object{
				ID:      "EPR.ID",
				ModTime: 1515151515678,
				Deleted: true,
				Extra:   []byte(`{"meta":true}`),
			}))

			_, err := txn.Store.Get("/objects/EPR.ID")
			Expect(err).To(MatchError(storage.ErrNotFound))
			Expect(txn.Perms.GetPermissions("/objects/EPR.ID")).To(BeEmpty())
		})

		It("deletes nested", func() {
			Expect(txn.Store.Get("/objects/EPR.ID/nested/ITR.ID")).NotTo(BeNil())
			Expect(txn.Perms.GetPermissions("/objects/EPR.ID/nested/ITR.ID")).To(HaveLen(1))

			Expect(subject.Delete(txn, "/objects/EPR.ID/nested/ITR.ID")).To(Equal(&schema.Object{
				ID:      "ITR.ID",
				ModTime: 1515151515678,
				Deleted: true,
				Extra:   []byte(`{"meta":true}`),
			}))

			_, err := txn.Store.Get("/objects/EPR.ID/nested/ITR.ID")
			Expect(err).To(MatchError(storage.ErrNotFound))
			Expect(txn.Perms.GetPermissions("/objects/EPR.ID/nested/ITR.ID")).To(BeEmpty())
		})
	})

	Describe("DeleteAll", func() {
		BeforeEach(func() {
			nested := api.StdModel()

			// seed:
			//   /objects/EPR.ID
			//   /objects/ITR.ID
			//   /objects/MXR.ID
			Expect(subject.Create(txn, "/objects/*", resource())).To(Succeed())
			Expect(subject.Create(txn, "/objects/*", resource())).To(Succeed())
			Expect(subject.Create(txn, "/objects/*", resource())).To(Succeed())

			// seed:
			//   /objects/EPR.ID/nested/Q3R.ID
			//   /objects/ITR.ID/nested/U7R.ID
			//   /objects/MXR.ID/nested/ZDR.ID
			Expect(nested.Create(txn, "/objects/EPR.ID/nested/*", resource())).To(Succeed())
			Expect(nested.Create(txn, "/objects/ITR.ID/nested/*", resource())).To(Succeed())
			Expect(nested.Create(txn, "/objects/MXR.ID/nested/*", resource())).To(Succeed())
		})

		It("deletes objects", func() {
			Expect(txn.Store.Get("/objects/EPR.ID")).NotTo(BeNil())
			Expect(txn.Store.Get("/objects/ITR.ID")).NotTo(BeNil())
			Expect(txn.Store.Get("/objects/MXR.ID")).NotTo(BeNil())
			Expect(txn.Perms.GetPermissions("/objects/EPR.ID")).To(HaveLen(1))
			Expect(txn.Perms.GetPermissions("/objects/ITR.ID")).To(HaveLen(1))
			Expect(txn.Perms.GetPermissions("/objects/MXR.ID")).To(HaveLen(1))

			Expect(subject.DeleteAll(txn, "/objects/*",
				"EPR.ID",
				"BADID",
				"ITR.ID",
			)).To(Equal(riposo.Epoch(1515151515681)))

			_, err := txn.Store.Get("/objects/EPR.ID")
			Expect(err).To(MatchError(storage.ErrNotFound))
			_, err = txn.Store.Get("/objects/ITR.ID")
			Expect(err).To(MatchError(storage.ErrNotFound))
			Expect(txn.Store.Get("/objects/MXR.ID")).NotTo(BeNil())
			Expect(txn.Perms.GetPermissions("/objects/EPR.ID")).To(BeEmpty())
			Expect(txn.Perms.GetPermissions("/objects/ITR.ID")).To(BeEmpty())
			Expect(txn.Perms.GetPermissions("/objects/MXR.ID")).To(HaveLen(1))
		})

		It("deletes nested", func() {
			Expect(txn.Store.Get("/objects/EPR.ID/nested/Q3R.ID")).NotTo(BeNil())
			Expect(txn.Store.Get("/objects/ITR.ID/nested/U7R.ID")).NotTo(BeNil())
			Expect(txn.Store.Get("/objects/MXR.ID/nested/ZDR.ID")).NotTo(BeNil())
			Expect(txn.Perms.GetPermissions("/objects/EPR.ID/nested/Q3R.ID")).To(HaveLen(1))
			Expect(txn.Perms.GetPermissions("/objects/ITR.ID/nested/U7R.ID")).To(HaveLen(1))
			Expect(txn.Perms.GetPermissions("/objects/MXR.ID/nested/ZDR.ID")).To(HaveLen(1))

			Expect(subject.DeleteAll(txn, "/objects/*",
				"EPR.ID",
				"BADID",
				"ITR.ID",
			)).To(Equal(riposo.Epoch(1515151515681)))

			_, err := txn.Store.Get("/objects/EPR.ID/nested/Q3R.ID")
			Expect(err).To(MatchError(storage.ErrNotFound))
			_, err = txn.Store.Get("/objects/ITR.ID/nested/U7R.ID")
			Expect(err).To(MatchError(storage.ErrNotFound))
			Expect(txn.Store.Get("/objects/MXR.ID/nested/ZDR.ID")).NotTo(BeNil())
			Expect(txn.Perms.GetPermissions("/objects/EPR.ID/nested/Q3R.ID")).To(BeEmpty())
			Expect(txn.Perms.GetPermissions("/objects/ITR.ID/nested/U7R.ID")).To(BeEmpty())
			Expect(txn.Perms.GetPermissions("/objects/MXR.ID/nested/ZDR.ID")).To(HaveLen(1))
		})
	})
})
