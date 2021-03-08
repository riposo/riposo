package schema

import (
	"encoding/json"

	"github.com/tidwall/gjson"
)

// Value is a parameter/object value.
type Value gjson.Result

// StringValue converts a string literal to a value.
func StringValue(s string) Value {
	for i := 0; i < len(s); i++ {
		if s[i] < ' ' || s[i] > 0x7f || s[i] == '"' || s[i] == '\\' {
			raw, _ := json.Marshal(s)
			return Value{Type: gjson.String, Raw: string(raw), Str: s}
		}
	}

	return Value{Type: gjson.String, Raw: `"` + s + `"`, Str: s}
}

// ParseValue parses a value.
func ParseValue(s string) Value {
	if s == "" {
		return Value{Raw: "null"}
	}

	if res := gjson.Parse(s); res.Raw != "" {
		return Value(res)
	}

	return StringValue(s)
}

// IsNull returns true if value is null.
func (v Value) IsNull() bool {
	return v.Type == gjson.Null
}

// String returns the string representation of the value.
func (v Value) String() string {
	return gjson.Result(v).String()
}

// Int returns the integer representation of the value.
func (v Value) Int() int64 {
	return gjson.Result(v).Int()
}

// Float returns the numeric representation of the value.
func (v Value) Float() float64 {
	return gjson.Result(v).Float()
}

// Bool returns the boolean representation of the value.
func (v Value) Bool() bool {
	return gjson.Result(v).Bool()
}

// Exists returns true it the value exists.
func (v Value) Exists() bool {
	return gjson.Result(v).Exists()
}

// Value returns a the value converted to the appropriate type.
func (v Value) Value() interface{} {
	return gjson.Result(v).Value()
}

// MarshalJSON implement custom JSON marshaler.
func (v Value) MarshalJSON() ([]byte, error) {
	if v.Raw == "" {
		return []byte("null"), nil
	}
	return []byte(v.Raw), nil
}

// UnmarshalJSON implement custom JSON unmarshaler.
func (v *Value) UnmarshalJSON(p []byte) error {
	if len(p) == 0 {
		*v = Value{Raw: "null"}
	} else if res := gjson.ParseBytes(p); res.Raw != "" {
		*v = Value(res)
	}
	return nil
}
