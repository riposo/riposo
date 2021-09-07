package api_test

import (
	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/riposo"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/ginkgo/extensions/table"
	. "github.com/bsm/gomega"
)

var _ = Describe("Hooks", func() {
	var subject *api.Hooks
	var h1, h2 = &mockHook{ID: 1}, &mockHook{ID: 2}

	BeforeEach(func() {
		subject = new(api.Hooks)

		subject.Register([]string{
			"/buckets/*/collections/mock/*/**",
			"!**/secret",
			"/buckets/**/collections/mock/q*/secret",
		}, h1)

		subject.Register([]string{
			"/buckets/*/collections/*",
		}, h2)
	})

	It("registers hooks", func() {
		Expect(subject.Len()).To(Equal(2))
	})

	It("rejects hooks without patterns", func() {
		subject.Register(nil, h1)
		subject.Register([]string{}, h1)
		Expect(subject.Len()).To(Equal(2))
	})

	It("rejects hooks with invalid patterns", func() {
		subject.Register([]string{""}, h1)
		subject.Register([]string{"!"}, h1)
		subject.Register([]string{"bad[pattern"}, h1)
		Expect(subject.Len()).To(Equal(2))
	})

	DescribeTable("iterates over matching hooks",
		func(path string, exp ...api.Hook) {
			var matches []api.Hook
			Expect(subject.ForEach(riposo.Path(path), func(h api.Hook) error {
				matches = append(matches, h)
				return nil
			})).To(Succeed())
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

type mockHook struct {
	ID int
	api.NoopHook
}
