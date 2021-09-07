package params_test

import (
	"github.com/riposo/riposo/pkg/params"
	"github.com/riposo/riposo/pkg/schema"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Filter", func() {
	It("parses", func() {
		Expect(params.ParseFilter("field", "value")).To(Equal(params.Filter{
			Field:    "field",
			Operator: params.OperatorEQ,
			Values:   []schema.Value{{Type: gjson.String, Raw: `"value"`, Str: "value"}},
		}))

		Expect(params.ParseFilter("field", `"value"`)).To(Equal(params.Filter{
			Field:    "field",
			Operator: params.OperatorEQ,
			Values:   []schema.Value{{Type: gjson.String, Raw: `"value"`, Str: "value"}},
		}))

		Expect(params.ParseFilter("field", `null`)).To(Equal(params.Filter{
			Field:    "field",
			Operator: params.OperatorEQ,
			Values:   []schema.Value{{Type: gjson.Null, Raw: `null`}},
		}))

		Expect(params.ParseFilter("field", `1,2,3`)).To(Equal(params.Filter{
			Field:    "field",
			Operator: params.OperatorEQ,
			Values:   []schema.Value{{Type: gjson.Number, Raw: `1`, Num: 1}},
		}))

		Expect(params.ParseFilter("field", ``)).To(Equal(params.Filter{
			Field:    "field",
			Operator: params.OperatorEQ,
			Values:   []schema.Value{{Type: gjson.Null, Raw: `null`}},
		}))

		Expect(params.ParseFilter("not_field", `true`)).To(Equal(params.Filter{
			Field:    "field",
			Operator: params.OperatorNOT,
			Values:   []schema.Value{{Type: gjson.True, Raw: "true"}},
		}))

		Expect(params.ParseFilter("has_field", `true`)).To(Equal(params.Filter{
			Field:    "field",
			Operator: params.OperatorHAS,
			Values:   []schema.Value{{Type: gjson.True, Raw: "true"}},
		}))

		Expect(params.ParseFilter("gt_field.sub", `5`)).To(Equal(params.Filter{
			Field:    "field.sub",
			Operator: params.OperatorGT,
			Values:   []schema.Value{{Type: gjson.Number, Raw: "5", Num: 5}},
		}))

		Expect(params.ParseFilter("in_field", `a,null,b,,c`)).To(Equal(params.Filter{
			Field:    "field",
			Operator: params.OperatorIN,
			Values: []schema.Value{
				{Type: gjson.String, Raw: `"a"`, Str: "a"},
				{Type: gjson.Null, Raw: `null`},
				{Type: gjson.String, Raw: `"b"`, Str: "b"},
				{Type: gjson.Null, Raw: `null`},
				{Type: gjson.String, Raw: `"c"`, Str: "c"},
			},
		}))

		Expect(params.ParseFilter("exclude_field", `1,2,true`)).To(Equal(params.Filter{
			Field:    "field",
			Operator: params.OperatorEXCLUDE,
			Values: []schema.Value{
				{Type: gjson.Number, Raw: `1`, Num: 1},
				{Type: gjson.Number, Raw: `2`, Num: 2},
				{Type: gjson.True, Raw: `true`},
			},
		}))

		Expect(params.ParseFilter("contains_any_field", `x,y,z`)).To(Equal(params.Filter{
			Field:    "field",
			Operator: params.OperatorContainsAny,
			Values: []schema.Value{
				{Type: gjson.String, Raw: `"x"`, Str: "x"},
				{Type: gjson.String, Raw: `"y"`, Str: "y"},
				{Type: gjson.String, Raw: `"z"`, Str: "z"},
			},
		}))
	})
})
