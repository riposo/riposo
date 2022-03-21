package schema_test

import (
	"encoding/json"

	"github.com/tidwall/gjson"

	. "github.com/bsm/ginkgo/v2"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/pkg/schema"
)

var _ = Describe("Object", func() {
	var subject *Object

	BeforeEach(func() {
		subject = &Object{
			ID:      "EPR.ID",
			ModTime: 1567815678988,
			Extra:   []byte(`{"meta": true, "nested": {"num": 33}}`),
		}
	})

	It("gets values", func() {
		Expect(subject.Get("id")).To(Equal(Value{Type: gjson.String, Raw: `"EPR.ID"`, Str: "EPR.ID"}))
		Expect(subject.Get("last_modified")).To(Equal(Value{Type: gjson.Number, Raw: `1567815678988`, Num: 1567815678988}))
		Expect(subject.Get("deleted")).To(Equal(Value{Type: gjson.False, Raw: `false`}))
		Expect(subject.Get("meta")).To(Equal(Value{Type: gjson.True, Raw: `true`, Index: 9}))
		Expect(subject.Get("nested")).To(Equal(Value{Type: gjson.JSON, Raw: `{"num": 33}`, Index: 25}))
		Expect(subject.Get("nested.num")).To(Equal(Value{Type: gjson.Number, Raw: `33`, Num: 33, Index: 33}))
		Expect(subject.Get("unknown")).To(Equal(Value{}))
		Expect(subject.Get("nested.unknown")).To(Equal(Value{}))
	})

	It("sets values", func() {
		Expect(subject.Set("id", 123)).To(MatchError(`value is not a string`))
		Expect(subject.Set("last_modified", "foo")).To(MatchError(`value is not a number`))
		Expect(subject.Set("deleted", "bar")).To(MatchError(`value is not a boolean`))

		Expect(subject.Set("id", "foo")).To(Succeed())
		Expect(subject.Set("last_modified", 1515151515_000)).To(Succeed())
		Expect(subject.Set("deleted", true)).To(Succeed())
		Expect(subject.Set("meta", "bar")).To(Succeed())
		Expect(subject.Set("nested.num", 34)).To(Succeed())
		Expect(subject.Set("other", true)).To(Succeed())

		Expect(subject).To(Equal(&Object{
			ID:      "foo",
			ModTime: 1515151515_000,
			Deleted: true,
			Extra:   []byte(`{"meta": "bar", "nested": {"num": 34},"other":true}`),
		}))
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

		Expect(json.Marshal(new(Object))).To(MatchJSON(`{
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

	It("copies objects", func() {
		o2 := subject.Copy()
		o2.Extra[2] = 'b'
		Expect(o2.String()).To(MatchJSON(`{"id": "EPR.ID", "last_modified": 1567815678988, "beta": true, "nested": { "num": 33 }}`))
		Expect(subject.String()).To(MatchJSON(`{"id": "EPR.ID", "last_modified": 1567815678988, "meta": true, "nested": { "num": 33 }}`))
	})

	It("updates objects", func() {
		o2 := &Object{ID: "ITR.ID", ModTime: 1567815679000, Extra: []byte(`{"a": 1}`)}
		subject.Update(o2)
		Expect(subject).To(Equal(o2))

		subject.Update(&Object{})
		Expect(subject).To(Equal(o2))

		o2.Extra = append(o2.Extra[:6], '2', '}')
		Expect(subject.Extra).To(MatchJSON(`{"a": 1}`))

		o3 := &Object{}
		o3.Update(&Object{})
		Expect(o3.Extra).To(MatchJSON(`{}`))
	})

	It("patches objects", func() {
		o1 := &Object{}
		o2 := &Object{}
		Expect(o1.Patch(o2)).To(Succeed())
		Expect(o1.Extra).To(MatchJSON(`{}`))

		o2 = &Object{Extra: []byte(`{
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

		o1 = &Object{Extra: []byte(`{
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

		o1 = &Object{Extra: []byte(`{"a": true}`)}
		o2 = &Object{Extra: []byte(`{"a": null}`)}
		Expect(o1.Patch(o2)).To(Succeed())
		Expect(o1.Extra).To(MatchJSON(`{"a": true}`))
	})

	It("norms", func() {
		subject.Extra = nil
		subject.Norm()
		Expect(subject.Extra).To(MatchJSON(`{}`))
	})

	DescribeTable("calculates size",
		func(o *Object) {
			bin, _ := json.Marshal(o)
			Expect(o.ByteSize()).To(Equal(int64(len(bin))))
		},
		Entry("null", (*Object)(nil)),
		Entry("blank", &Object{}),
		Entry("only ID", &Object{ID: "EPR.ID"}),
		Entry("no extra", &Object{ID: "EPR.ID", ModTime: 1567815678988}),
		Entry("negative epoch", &Object{ID: "EPR.ID", ModTime: -101}),
		Entry("blank extra", &Object{
			ID:      "EPR.ID",
			ModTime: 1567815678988,
			Extra:   []byte(`{}`),
		}),
		Entry("full", &Object{
			ID:      "EPR.ID",
			ModTime: 1567815678988,
			Extra:   []byte(`{"meta":true,"nested":{"num":33}}`),
		}),
		Entry("deleted", &Object{
			ID:      "EPR.ID",
			ModTime: 1567815678988,
			Deleted: true,
		}),
	)
})
