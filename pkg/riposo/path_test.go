package riposo_test

import (
	"testing"

	"github.com/riposo/riposo/pkg/riposo"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Path", func() {
	var subject riposo.Path

	BeforeEach(func() {
		subject = riposo.Path("/buckets/foo/collections/bar/records/baz")
	})

	It("parses from URL path", func() {
		Expect(riposo.NormPath("/v1/buckets").String()).To(Equal("/buckets/*"))
		Expect(riposo.NormPath("/v1/buckets/foo").String()).To(Equal("/buckets/foo"))
		Expect(riposo.NormPath("/v1").String()).To(Equal(""))
		Expect(riposo.NormPath("").String()).To(Equal(""))

		Expect(riposo.NormPath("/versions").String()).To(Equal("/versions/*"))
		Expect(riposo.NormPath("/buckets").String()).To(Equal("/buckets/*"))
		Expect(riposo.NormPath("/buckets/foo").String()).To(Equal("/buckets/foo"))
	})

	It("calculates parent", func() {
		Expect(subject.Parent().String()).To(Equal("/buckets/foo/collections/bar"))
		Expect(subject.Parent().Parent().String()).To(Equal("/buckets/foo"))
		Expect(subject.Parent().Parent().Parent().String()).To(Equal(""))
		Expect(subject.Parent().Parent().Parent().Parent().String()).To(Equal(""))
	})

	It("extracts object ID", func() {
		Expect(subject.ObjectID()).To(Equal("baz"))
		Expect(riposo.Path("/").ObjectID()).To(Equal(""))
	})

	It("extracts resource name", func() {
		Expect(subject.ResourceName()).To(Equal("record"))
		Expect(subject.Parent().ResourceName()).To(Equal("collection"))
		Expect(subject.Parent().Parent().ResourceName()).To(Equal("bucket"))
		Expect(subject.Parent().Parent().Parent().ResourceName()).To(Equal(""))

		Expect(riposo.Path("/").ResourceName()).To(Equal(""))
		Expect(riposo.Path("/sheep/*").ResourceName()).To(Equal("sheep"))
	})

	It("replaces object ID", func() {
		Expect(subject.WithObjectID("boo").String()).To(Equal("/buckets/foo/collections/bar/records/boo"))
	})

	It("checks if path is a node", func() {
		Expect(riposo.Path("/buckets/foo").IsNode()).To(BeFalse())
		Expect(riposo.Path("").IsNode()).To(BeFalse())
		Expect(riposo.Path("/buckets/*").IsNode()).To(BeTrue())
	})

	It("checks if a path contains another", func() {
		Expect(riposo.Path("").Contains(riposo.Path(""))).To(BeTrue())
		Expect(riposo.Path("/buckets/foo").Contains(riposo.Path("/buckets/foo"))).To(BeTrue())
		Expect(riposo.Path("/buckets/*").Contains(riposo.Path("/buckets/foo"))).To(BeTrue())
		Expect(riposo.Path("*").Contains(riposo.Path(""))).To(BeTrue())

		Expect(riposo.Path("").Contains(riposo.Path("*"))).To(BeFalse())
		Expect(riposo.Path("/buckets/foo").Contains(riposo.Path("/buckets/*"))).To(BeFalse())
		Expect(riposo.Path("/buckets/foo").Contains(riposo.Path("/muppets/foo"))).To(BeFalse())
		Expect(riposo.Path("/buckets/foo").Contains(riposo.Path("/buckets/baz"))).To(BeFalse())
		Expect(riposo.Path("/buckets/foo").Contains(riposo.Path("/buckets/foo/collections/bar"))).To(BeFalse())
		Expect(riposo.Path("/buckets/*").Contains(riposo.Path("/buckets/foo/collections/bar"))).To(BeFalse())
		Expect(riposo.Path("/buckets/*").Contains(riposo.Path("/buckets/foo/collections/*"))).To(BeFalse())
		Expect(riposo.Path("/buckets/*").Contains(riposo.Path("*"))).To(BeFalse())
	})

	It("traverses", func() {
		var seen []string
		subject.Traverse(func(p riposo.Path) {
			seen = append(seen, p.String())
		})
		Expect(seen).To(Equal([]string{
			"/buckets/foo/collections/bar/records/baz",
			"/buckets/foo/collections/bar",
			"/buckets/foo",
			"",
		}))
	})
})

func BenchmarkPath_ResourceName(b *testing.B) {
	path := riposo.Path("/buckets/foo/collections/bar")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if s := path.ResourceName(); s != "collection" {
			b.Fatal("unexpected result", s)
		}
	}
}
