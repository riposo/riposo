package api_test

import (
	. "github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"

	. "github.com/bsm/ginkgo/v2"
	. "github.com/bsm/gomega"
)

var _ = Describe("CallbackChain", func() {
	var subject CallbackChain
	var h1, h2 = &mockCallbacks{}, &mockCallbacks{}

	BeforeEach(func() {
		subject = NewCallbackChain()

		subject.Register([]string{
			"/buckets/*/collections/mock/*/**",
			"!**/secret",
			"/buckets/**/collections/mock/q*/secret",
		}, h1)

		subject.Register([]string{
			"/buckets/*/collections/*",
		}, h2)
	})

	It("registers Callbacks", func() {
		Expect(subject.Len()).To(Equal(2))
	})

	It("rejects Callbacks without patterns", func() {
		subject.Register(nil, h1)
		subject.Register([]string{}, h1)
		Expect(subject.Len()).To(Equal(2))
	})

	It("rejects Callbacks with invalid patterns", func() {
		subject.Register([]string{""}, h1)
		subject.Register([]string{"!"}, h1)
		subject.Register([]string{"bad[pattern"}, h1)
		Expect(subject.Len()).To(Equal(2))
	})

	DescribeTable("iterates over matching Callbacks",
		func(path string, exp ...Callbacks) {
			var matches []Callbacks
			subject.ForEach(riposo.Path(path), func(h Callbacks) {
				matches = append(matches, h)
			})
			if len(exp) == 0 {
				Expect(matches).To(BeEmpty())
			} else {
				Expect(matches).To(Equal(exp))
			}
		},
		Entry("no match", "/buckets/foo"),
		Entry("wildcard match", "/buckets/foo/collections/bar", h2),
		Entry("single * is not optional", "/buckets/foo/collections/mock", h2),
		Entry("suffix match", "/buckets/foo/collections/mock/records", h1),
		Entry("exclusions", "/buckets/foo/collections/mock/records/secret"),
		Entry("exclusion override", "/buckets/foo/collections/mock/quota/secret", h1),
	)
})

type mockCallbacks struct {
	NumBeforeCreate, NumAfterCreate       int
	NumBeforeUpdate, NumAfterUpdate       int
	NumBeforePatch, NumAfterPatch         int
	NumBeforeDelete, NumAfterDelete       int
	NumBeforeDeleteAll, NumAfterDeleteAll int
}

func (m *mockCallbacks) Reset()                                              { *m = mockCallbacks{} }
func (m *mockCallbacks) OnCreate(_ *Txn, _ riposo.Path) CreateCallback       { return m }
func (m *mockCallbacks) OnUpdate(_ *Txn, _ riposo.Path) UpdateCallback       { return m }
func (m *mockCallbacks) OnPatch(_ *Txn, _ riposo.Path) PatchCallback         { return m }
func (m *mockCallbacks) OnDelete(_ *Txn, _ riposo.Path) DeleteCallback       { return m }
func (m *mockCallbacks) OnDeleteAll(_ *Txn, _ riposo.Path) DeleteAllCallback { return m }

func (m *mockCallbacks) BeforeCreate(payload *schema.Resource) error { m.NumBeforeCreate++; return nil }
func (m *mockCallbacks) AfterCreate(created *schema.Resource) error  { m.NumAfterCreate++; return nil }
func (m *mockCallbacks) BeforeUpdate(existing *schema.Object, payload *schema.Resource) error {
	m.NumBeforeUpdate++
	return nil
}
func (m *mockCallbacks) AfterUpdate(updated *schema.Resource) error { m.NumAfterUpdate++; return nil }
func (m *mockCallbacks) BeforePatch(existing *schema.Object, payload *schema.Resource) error {
	m.NumBeforePatch++
	return nil
}
func (m *mockCallbacks) AfterPatch(patched *schema.Resource) error  { m.NumAfterPatch++; return nil }
func (m *mockCallbacks) BeforeDelete(existing *schema.Object) error { m.NumBeforeDelete++; return nil }
func (m *mockCallbacks) AfterDelete(deleted *schema.Object) error   { m.NumAfterDelete++; return nil }
func (m *mockCallbacks) BeforeDeleteAll(objIDs []string) error      { m.NumBeforeDeleteAll++; return nil }
func (m *mockCallbacks) AfterDeleteAll(modTime riposo.Epoch, deleted []riposo.Path) error {
	m.NumAfterDeleteAll++
	return nil
}
