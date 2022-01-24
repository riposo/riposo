package params_test

import (
	"github.com/riposo/riposo/pkg/schema"
	"github.com/tidwall/gjson"

	. "github.com/bsm/ginkgo/v2"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/pkg/params"
)

var _ = Describe("Filter", func() {
	It("parses", func() {
		Expect(ParseFilter("field", "value")).To(Equal(Filter{
			Field:    "field",
			Operator: OperatorEQ,
			Values:   []schema.Value{{Type: gjson.String, Raw: `"value"`, Str: "value"}},
		}))

		Expect(ParseFilter("field", `"value"`)).To(Equal(Filter{
			Field:    "field",
			Operator: OperatorEQ,
			Values:   []schema.Value{{Type: gjson.String, Raw: `"value"`, Str: "value"}},
		}))

		Expect(ParseFilter("field", `null`)).To(Equal(Filter{
			Field:    "field",
			Operator: OperatorEQ,
			Values:   []schema.Value{{Type: gjson.Null, Raw: `null`}},
		}))

		Expect(ParseFilter("field", `1,2,3`)).To(Equal(Filter{
			Field:    "field",
			Operator: OperatorEQ,
			Values:   []schema.Value{{Type: gjson.String, Raw: `"1,2,3"`, Str: "1,2,3"}},
		}))

		Expect(ParseFilter("field", ``)).To(Equal(Filter{
			Field:    "field",
			Operator: OperatorEQ,
			Values:   []schema.Value{{Type: gjson.Null, Raw: `null`}},
		}))

		Expect(ParseFilter("not_field", `true`)).To(Equal(Filter{
			Field:    "field",
			Operator: OperatorNOT,
			Values:   []schema.Value{{Type: gjson.True, Raw: "true"}},
		}))

		Expect(ParseFilter("has_field", `true`)).To(Equal(Filter{
			Field:    "field",
			Operator: OperatorHAS,
			Values:   []schema.Value{{Type: gjson.True, Raw: "true"}},
		}))

		Expect(ParseFilter("gt_field.sub", `5`)).To(Equal(Filter{
			Field:    "field.sub",
			Operator: OperatorGT,
			Values:   []schema.Value{{Type: gjson.Number, Raw: "5", Num: 5}},
		}))

		Expect(ParseFilter("in_field", `a,null,b,,c`)).To(Equal(Filter{
			Field:    "field",
			Operator: OperatorIN,
			Values: []schema.Value{
				{Type: gjson.String, Raw: `"a"`, Str: "a"},
				{Type: gjson.Null, Raw: `null`},
				{Type: gjson.String, Raw: `"b"`, Str: "b"},
				{Type: gjson.Null, Raw: `null`},
				{Type: gjson.String, Raw: `"c"`, Str: "c"},
			},
		}))

		Expect(ParseFilter("exclude_field", `1,2,true`)).To(Equal(Filter{
			Field:    "field",
			Operator: OperatorEXCLUDE,
			Values: []schema.Value{
				{Type: gjson.Number, Raw: `1`, Num: 1},
				{Type: gjson.Number, Raw: `2`, Num: 2},
				{Type: gjson.True, Raw: `true`},
			},
		}))

		Expect(ParseFilter("contains_any_field", `x,y,z`)).To(Equal(Filter{
			Field:    "field",
			Operator: OperatorContainsAny,
			Values: []schema.Value{
				{Type: gjson.String, Raw: `"x"`, Str: "x"},
				{Type: gjson.String, Raw: `"y"`, Str: "y"},
				{Type: gjson.String, Raw: `"z"`, Str: "z"},
			},
		}))
	})
})
