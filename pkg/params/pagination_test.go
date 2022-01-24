package params_test

import (
	"github.com/riposo/riposo/pkg/schema"
	"github.com/tidwall/gjson"

	. "github.com/bsm/ginkgo/v2"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/pkg/params"
)

var _ = Describe("Pagination", func() {
	It("parses blank", func() {
		Expect(ParseToken("")).To(BeNil())
	})

	It("encodes/parses", func() {
		var t *Pagination
		Expect(t.Encode()).To(Equal(""))

		t = &Pagination{
			Nonce:   "x",
			LastObj: map[string]schema.Value{"field": {Type: gjson.Number, Raw: "33", Num: 33.0}},
		}
		s, err := t.Encode()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(s)).To(BeNumerically("~", 54, 10))
		Expect(ParseToken(s)).To(Equal(t))
	})

	It("generates conditions", func() {
		var t *Pagination
		Expect(t.Conditions()).To(BeNil())

		t = &Pagination{}
		Expect(t.Conditions()).To(BeNil())

		t = &Pagination{
			LastObj: map[string]schema.Value{
				"field": schema.ParseValue("33"),
				"other": schema.StringValue("foo"),
			},
		}

		Expect(t.Conditions()).To(ConsistOf(
			ConsistOf(
				Filter{Field: "field", Operator: OperatorGT, Values: []schema.Value{schema.ParseValue("33")}},
				Filter{Field: "other", Operator: OperatorEQ, Values: []schema.Value{schema.StringValue("foo")}},
			),
			ConsistOf(
				Filter{Field: "field", Operator: OperatorEQ, Values: []schema.Value{schema.ParseValue("33")}},
				Filter{Field: "other", Operator: OperatorGT, Values: []schema.Value{schema.StringValue("foo")}},
			),
		))
	})
})
