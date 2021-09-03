package params_test

import (
	"github.com/riposo/riposo/pkg/params"
	"github.com/riposo/riposo/pkg/schema"
	"github.com/tidwall/gjson"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
)

var _ = Describe("Pagination", func() {
	It("parses blank", func() {
		Expect(params.ParseToken("")).To(BeNil())
	})

	It("encodes/parses", func() {
		var t *params.Pagination
		Expect(t.Encode()).To(Equal(""))

		t = &params.Pagination{
			Nonce:   "x",
			LastObj: map[string]schema.Value{"field": {Type: gjson.Number, Raw: "33", Num: 33.0}},
		}
		s, err := t.Encode()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(s)).To(BeNumerically("~", 54, 10))
		Expect(params.ParseToken(s)).To(Equal(t))
	})

	It("generates conditions", func() {
		var t *params.Pagination
		Expect(t.Conditions()).To(BeNil())

		t = &params.Pagination{}
		Expect(t.Conditions()).To(BeNil())

		t = &params.Pagination{
			LastObj: map[string]schema.Value{
				"field": schema.ParseValue("33"),
				"other": schema.StringValue("foo"),
			},
		}

		Expect(t.Conditions()).To(ConsistOf(
			ConsistOf(
				params.Filter{Field: "field", Operator: params.OperatorGT, Values: []schema.Value{schema.ParseValue("33")}},
				params.Filter{Field: "other", Operator: params.OperatorEQ, Values: []schema.Value{schema.StringValue("foo")}},
			),
			ConsistOf(
				params.Filter{Field: "field", Operator: params.OperatorEQ, Values: []schema.Value{schema.ParseValue("33")}},
				params.Filter{Field: "other", Operator: params.OperatorGT, Values: []schema.Value{schema.StringValue("foo")}},
			),
		))
	})
})
