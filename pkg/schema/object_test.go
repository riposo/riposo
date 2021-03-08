package schema_test

import (
	"encoding/json"

	"github.com/riposo/riposo/pkg/schema"
	"github.com/tidwall/gjson"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Object", func() {
	var subject *schema.Object

	BeforeEach(func() {
		subject = &schema.Object{
			ID:      "EPR.ID",
			ModTime: 1567815678988,
			Extra:   []byte(`{"meta": true, "nested": {"num": 33}}`),
		}
	})

	It("gets values", func() {
		Expect(subject.Get("id")).To(Equal(schema.Value{Type: gjson.String, Raw: `"EPR.ID"`, Str: "EPR.ID"}))
		Expect(subject.Get("last_modified")).To(Equal(schema.Value{Type: gjson.Number, Raw: `1567815678988`, Num: 1567815678988}))
		Expect(subject.Get("deleted")).To(Equal(schema.Value{Type: gjson.False, Raw: `false`}))
		Expect(subject.Get("meta")).To(Equal(schema.Value{Type: gjson.True, Raw: `true`, Index: 9}))
		Expect(subject.Get("nested")).To(Equal(schema.Value{Type: gjson.JSON, Raw: `{"num": 33}`, Index: 25}))
		Expect(subject.Get("nested.num")).To(Equal(schema.Value{Type: gjson.Number, Raw: `33`, Num: 33, Index: 33}))
		Expect(subject.Get("unknown")).To(Equal(schema.Value{}))
		Expect(subject.Get("nested.unknown")).To(Equal(schema.Value{}))
	})

	It("marshals to JSON", func() {
		Expect(json.Marshal(subject)).To(MatchJSON(`{
			"id": "EPR.ID",
			"last_modified": 1567815678988,
			"meta": true,
			"nested": {"num": 33}
		}`))

		subject.Deleted = true
		Expect(json.Marshal(subject)).To(MatchJSON(`{
			"id": "EPR.ID",
			"last_modified": 1567815678988,
			"deleted": true
		}`))

		Expect(json.Marshal(new(schema.Object))).To(MatchJSON(`{
			"id": "",
			"last_modified": 0
		}`))
	})

	It("decodes/encodes extra", func() {
		var x struct {
			Meta bool `json:"meta"`
		}
		Expect(subject.DecodeExtra(&x)).To(Succeed())

		x.Meta = false
		Expect(subject.EncodeExtra(&x)).To(Succeed())
		Expect(string(subject.Extra)).To(Equal(`{"meta":false}`))

		var y struct {
			Meta string `json:"meta"`
		}
		Expect(subject.DecodeExtra(&y).(*json.UnmarshalTypeError).Field).To(Equal("data.meta"))
	})

	It("updates objects", func() {
		o2 := &schema.Object{ID: "ITR.ID", ModTime: 1567815679000, Extra: []byte(`{"a": 1}`)}
		subject.Update(o2)
		Expect(subject).To(Equal(o2))

		subject.Update(&schema.Object{})
		Expect(subject).To(Equal(o2))

		o2.Extra = append(o2.Extra[:6], '2', '}')
		Expect(subject.Extra).To(MatchJSON(`{"a": 1}`))
	})

	It("patches objects", func() {
		o1 := &schema.Object{}
		o2 := &schema.Object{}
		Expect(o1.Patch(o2)).To(Succeed())
		Expect(o1.Extra).To(BeNil())

		o2 = &schema.Object{Extra: []byte(`{
			"a": "ok",
			"d": [false, 8],
			"e": {"y": 2, "z": 3},
			"f": "extra"
		}`)}
		Expect(o1.Patch(o2)).To(Succeed())
		Expect(o1.Extra).To(MatchJSON(`{
			"a": "ok",
			"d": [false, 8],
			"e": {"y": 2, "z": 3},
			"f": "extra"
		}`))

		o1 = &schema.Object{Extra: []byte(`{
			"a": true,
			"b": 33,
			"c": "str",
			"d": [9, "o"],
			"e": {"x": 1, "y": 3}
		}`)}
		Expect(o1.Patch(o2)).To(Succeed())
		Expect(o1.Extra).To(MatchJSON(`{
			"a": "ok",
			"b": 33,
			"c": "str",
			"d": [false, 8],
			"e": {"x": 1, "y": 2, "z": 3},
			"f": "extra"
		}`))

		o1 = &schema.Object{Extra: []byte(`{"a": true}`)}
		o2 = &schema.Object{Extra: []byte(`{"a": null}`)}
		Expect(o1.Patch(o2)).To(Succeed())
		Expect(o1.Extra).To(MatchJSON(`{"a": true}`))
	})
})
