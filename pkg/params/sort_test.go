package params_test

import (
	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/pkg/params"
)

var _ = Describe("ParseSort", func() {
	It("parses", func() {
		Expect(ParseSort("")).To(BeNil())
		Expect(ParseSort("field,-nested.field,other")).To(Equal([]SortOrder{
			{Field: "field"},
			{Field: "nested.field", Descending: true},
			{Field: "other"},
		}))
		Expect(ParseSort(",,-,,,")).To(BeNil())
		Expect(ParseSort(",,field,,")).To(Equal([]SortOrder{
			{Field: "field"},
		}))
		Expect(ParseSort("field,-field,other,field")).To(Equal([]SortOrder{
			{Field: "field"},
			{Field: "other"},
		}))
	})
})
