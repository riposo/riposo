package util_test

import (
	"encoding/json"

	"github.com/riposo/riposo/pkg/util"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
)

var _ = Describe("Set", func() {
	var subject util.Set

	BeforeEach(func() {
		subject = util.NewSet("a", "c", "b")
	})

	It("returns sorted slice", func() {
		Expect(subject.Slice()).To(Equal([]string{"a", "b", "c"}))
	})

	It("has len", func() {
		Expect(subject.Len()).To(Equal(3))
	})

	It("adds/removes", func() {
		subject.Add("c")
		subject.Add("x")
		Expect(subject.Slice()).To(Equal([]string{"a", "b", "c", "x"}))

		subject.Remove("b")
		subject.Remove("y")
		Expect(subject.Slice()).To(Equal([]string{"a", "c", "x"}))
	})

	It("checks inclusion", func() {
		Expect(subject.Has("a")).To(BeTrue())
		Expect(subject.Has("b")).To(BeTrue())
		Expect(subject.Has("x")).To(BeFalse())

		Expect(subject.HasAny("a", "b")).To(BeTrue())
		Expect(subject.HasAny("a", "x")).To(BeTrue())
		Expect(subject.HasAny("x", "y")).To(BeFalse())
		Expect(subject.HasAny()).To(BeFalse())
	})

	It("checks for intersections", func() {
		Expect(subject.IntersectsWith(util.NewSet("b", "d"))).To(BeTrue())
		Expect(subject.IntersectsWith(util.NewSet("c", "a"))).To(BeTrue())
		Expect(subject.IntersectsWith(util.NewSet("x", "y"))).To(BeFalse())
		Expect(subject.IntersectsWith(util.NewSet())).To(BeFalse())
	})

	It("constructs unions", func() {
		union := util.NewUnion(subject, util.NewSet("b", "x"))
		Expect(subject.Len()).To(Equal(3))
		Expect(union.Slice()).To(Equal([]string{"a", "b", "c", "x"}))
	})

	It("marshals/unmarshals", func() {
		Expect(json.Marshal(subject)).To(MatchJSON(`["a", "b", "c"]`))

		var s1 util.Set
		Expect(json.Unmarshal([]byte(`["b", "a", "c"]`), &s1)).To(Succeed())
		Expect(s1).To(Equal(subject))

		var s2 util.Set
		Expect(json.Unmarshal([]byte(`{"b": {}, "a": {}, "c": {}}`), &s2)).To(Succeed())
		Expect(s1).To(Equal(subject))
	})
})
