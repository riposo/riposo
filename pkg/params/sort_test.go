package params_test

import (
	"github.com/riposo/riposo/pkg/params"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ParseSort", func() {
	It("parses", func() {
		Expect(params.ParseSort("")).To(BeNil())
		Expect(params.ParseSort("field,-nested.field,other")).To(Equal([]params.SortOrder{
			{Field: "field"},
			{Field: "nested.field", Descending: true},
			{Field: "other"},
		}))
		Expect(params.ParseSort(",,-,,,")).To(BeNil())
		Expect(params.ParseSort(",,field,,")).To(Equal([]params.SortOrder{
			{Field: "field"},
		}))
		Expect(params.ParseSort("field,-field,other,field")).To(Equal([]params.SortOrder{
			{Field: "field"},
			{Field: "other"},
		}))
	})
})
