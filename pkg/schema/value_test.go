package schema_test

import (
	"encoding/json"

	"github.com/riposo/riposo/pkg/schema"
	"github.com/tidwall/gjson"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Value", func() {
	It("(un-)marshals", func() {
		var val schema.Value
		Expect(json.Unmarshal([]byte(`"xx"`), &val)).To(Succeed())
		Expect(val).To(Equal(schema.Value{
			Type: gjson.String, Raw: `"xx"`, Str: "xx",
		}))

		Expect(json.Marshal(val)).To(MatchJSON(`"xx"`))

		Expect(json.Unmarshal([]byte("33"), &val)).To(Succeed())
		Expect(val).To(Equal(schema.Value{
			Type: gjson.Number, Raw: `33`, Num: 33,
		}))
	})
})
