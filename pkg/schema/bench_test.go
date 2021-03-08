package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/riposo/riposo/pkg/schema"
)

func BenchmarkObject_MarshalJSON(b *testing.B) {
	o := &schema.Object{
		ID:      "EPR.ID",
		ModTime: 1567815678988,
		Extra:   []byte(`{"meta": true}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(o); err != nil {
			b.Fatal("unexpected error", err)
		}
	}
}

func BenchmarkObject_UnmarshalJSON(b *testing.B) {
	raw := `{
		"id": "EPR.ID",
		"last_modified": 1567815678988,
		"meta": true
	}`
	var data []byte
	var o schema.Object

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data = append(data[:0], raw...)
		if err := json.Unmarshal(data, &o); err != nil {
			b.Fatal("unexpected error", err)
		}
	}
}

func BenchmarkObject_Patch(b *testing.B) {
	d1 := []byte(`{
		"a": true,
		"b": 33,
		"c": "str",
		"d": [9, "o"],
		"e": {"x": 1, "y": 3}
	}`)
	o1 := &schema.Object{Extra: append([]byte{}, d1...)}
	o2 := &schema.Object{Extra: []byte(`{
		"a": "ok",
		"d": [false, 8],
		"e": {"y": 2, "z": 3},
		"f": "extra"
	}`)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		o1.Extra = append(o1.Extra[:0], d1...)
		if err := o1.Patch(o2); err != nil {
			b.Fatal("unexpected error", err)
		}
	}
}
