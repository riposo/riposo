package riposo_test

import (
	"testing"

	. "github.com/bsm/ginkgo/v2"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/pkg/riposo"
)

var _ = Describe("Path", func() {
	var subject Path

	BeforeEach(func() {
		subject = Path("/buckets/foo/collections/bar/records/baz")
	})

	It("parses from URL path", func() {
		Expect(NormPath("/v1/buckets").String()).To(Equal("/buckets/*"))
		Expect(NormPath("/v1/buckets/foo").String()).To(Equal("/buckets/foo"))
		Expect(NormPath("/v1").String()).To(Equal(""))
		Expect(NormPath("").String()).To(Equal(""))

		Expect(NormPath("/versions").String()).To(Equal("/versions/*"))
		Expect(NormPath("/buckets").String()).To(Equal("/buckets/*"))
		Expect(NormPath("/buckets/foo").String()).To(Equal("/buckets/foo"))
	})

	It("calculates parent", func() {
		Expect(subject.Parent().String()).To(Equal("/buckets/foo/collections/bar"))
		Expect(subject.Parent().Parent().String()).To(Equal("/buckets/foo"))
		Expect(subject.Parent().Parent().Parent().String()).To(Equal(""))
		Expect(subject.Parent().Parent().Parent().Parent().String()).To(Equal(""))
	})

	It("extracts object ID", func() {
		Expect(subject.ObjectID()).To(Equal("baz"))
		Expect(Path("/").ObjectID()).To(Equal(""))
	})

	It("extracts resource name", func() {
		Expect(subject.ResourceName()).To(Equal("record"))
		Expect(subject.Parent().ResourceName()).To(Equal("collection"))
		Expect(subject.Parent().Parent().ResourceName()).To(Equal("bucket"))
		Expect(subject.Parent().Parent().Parent().ResourceName()).To(Equal(""))

		Expect(Path("/").ResourceName()).To(Equal(""))
		Expect(Path("/sheep/*").ResourceName()).To(Equal("sheep"))
	})

	It("replaces object ID", func() {
		Expect(subject.WithObjectID("boo").String()).To(Equal("/buckets/foo/collections/bar/records/boo"))
	})

	It("checks if path is a node", func() {
		Expect(Path("/buckets/foo").IsNode()).To(BeFalse())
		Expect(Path("").IsNode()).To(BeFalse())
		Expect(Path("/buckets/*").IsNode()).To(BeTrue())
	})

	It("checks if a path contains another", func() {
		Expect(Path("").Contains(Path(""))).To(BeTrue())
		Expect(Path("/buckets/foo").Contains(Path("/buckets/foo"))).To(BeTrue())
		Expect(Path("/buckets/*").Contains(Path("/buckets/foo"))).To(BeTrue())
		Expect(Path("*").Contains(Path(""))).To(BeTrue())

		Expect(Path("").Contains(Path("*"))).To(BeFalse())
		Expect(Path("/buckets/foo").Contains(Path("/buckets/*"))).To(BeFalse())
		Expect(Path("/buckets/foo").Contains(Path("/muppets/foo"))).To(BeFalse())
		Expect(Path("/buckets/foo").Contains(Path("/buckets/baz"))).To(BeFalse())
		Expect(Path("/buckets/foo").Contains(Path("/buckets/foo/collections/bar"))).To(BeFalse())
		Expect(Path("/buckets/*").Contains(Path("/buckets/foo/collections/bar"))).To(BeFalse())
		Expect(Path("/buckets/*").Contains(Path("/buckets/foo/collections/*"))).To(BeFalse())
		Expect(Path("/buckets/*").Contains(Path("*"))).To(BeFalse())
	})

	It("traverses", func() {
		var seen []string
		subject.Traverse(func(p Path) {
			seen = append(seen, p.String())
		})
		Expect(seen).To(Equal([]string{
			"/buckets/foo/collections/bar/records/baz",
			"/buckets/foo/collections/bar",
			"/buckets/foo",
			"",
		}))
	})

	It("matches single patterns", func() {
		Expect(subject.Match("/buckets/**")).To(BeTrue())
		Expect(subject.Match("/buckets/**/records/*")).To(BeTrue())
		Expect(subject.Match("/buckets/*/collections/*/records/baz")).To(BeTrue())
		Expect(subject.Match("!/other/**")).To(BeTrue())

		Expect(subject.Match("")).To(BeFalse())
		Expect(subject.Match("/buckets/**/records")).To(BeFalse())
		Expect(subject.Match("/buckets/*/group/*/records/baz")).To(BeFalse())
		Expect(subject.Match("!/buckets/**")).To(BeFalse())
	})

	DescribeTable("matches multiple patterns",
		func(path string, exp bool) {
			Expect(Path(path).Match(
				"/buckets/*/collections/mock/*/**",
				"!**/secret",
				"/buckets/**/collections/mock/q*/secret",
			)).To(Equal(exp))
		},
		Entry("no match", "/buckets/foo", false),
		Entry("no deep match", "/buckets/foo/collections/bar", false),
		Entry("single * is not optional", "/buckets/foo/collections/mock", false),
		Entry("suffix match", "/buckets/foo/collections/mock/records", true),
		Entry("exclusions", "/buckets/foo/collections/mock/records/secret", false),
		Entry("exclusion override", "/buckets/foo/collections/mock/quota/secret", true),
	)
})

func BenchmarkPath_ResourceName(b *testing.B) {
	path := Path("/buckets/foo/collections/bar")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if s := path.ResourceName(); s != "collection" {
			b.Fatal("unexpected result", s)
		}
	}
}
