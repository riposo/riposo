package schema_test

import (
	"encoding/json"

	"github.com/tidwall/gjson"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/pkg/schema"
)

var _ = Describe("Value", func() {
	It("parses", func() {
		Expect(ParseValue("")).To(Equal(Value{Type: gjson.Null, Raw: `null`}))
		Expect(ParseValue("PLAIN")).To(Equal(Value{Type: gjson.String, Raw: `"PLAIN"`, Str: "PLAIN"}))
		Expect(ParseValue("iz")).To(Equal(Value{Type: gjson.String, Raw: `"iz"`, Str: "iz"}))
		Expect(ParseValue("123")).To(Equal(Value{Type: gjson.Number, Raw: `123`, Num: 123}))
		Expect(ParseValue("true")).To(Equal(Value{Type: gjson.True, Raw: `true`}))
		Expect(ParseValue("1,2,3")).To(Equal(Value{Type: gjson.String, Raw: `"1,2,3"`, Str: "1,2,3"}))
	})

	It("(un-)marshals", func() {
		var val Value
		Expect(json.Unmarshal([]byte(`"xx"`), &val)).To(Succeed())
		Expect(val).To(Equal(Value{
			Type: gjson.String, Raw: `"xx"`, Str: "xx",
		}))

		Expect(json.Marshal(val)).To(MatchJSON(`"xx"`))

		Expect(json.Unmarshal([]byte("33"), &val)).To(Succeed())
		Expect(val).To(Equal(Value{
			Type: gjson.Number, Raw: `33`, Num: 33,
		}))
	})
})
