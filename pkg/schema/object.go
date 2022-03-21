package schema

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"strconv"

	"github.com/riposo/riposo/pkg/riposo"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var newLine = []byte("\n")

var (
	errValNotString = errors.New("value is not a string")
	errValNotNumber = errors.New("value is not a number")
	errValNotBool   = errors.New("value is not a boolean")
)

// Object is a stored object.
type Object struct {
	ID      string
	ModTime riposo.Epoch
	Deleted bool
	Extra   []byte
}

type objectCore struct {
	ID      string       `json:"id"`            // the object ID
	ModTime riposo.Epoch `json:"last_modified"` // the last modified epoch
	Deleted bool         `json:"deleted,omitempty"`
}

// Get returns the value of field.
func (o *Object) Get(field string) Value {
	switch field {
	case "id":
		return Value{Type: gjson.String, Raw: strconv.Quote(o.ID), Str: o.ID}
	case "last_modified":
		return Value{Type: gjson.Number, Raw: strconv.FormatInt(int64(o.ModTime), 10), Num: float64(o.ModTime)}
	case "deleted":
		if o.Deleted {
			return Value{Type: gjson.True, Raw: "true"}
		}
		return Value{Type: gjson.False, Raw: "false"}
	default:
		return Value(gjson.GetBytes(o.Extra, field))
	}
}

// Set sets a field value.
func (o *Object) Set(field string, value interface{}) error {
	switch field {
	case "id":
		switch val := value.(type) {
		case string:
			o.ID = val
		default:
			return errValNotString
		}
	case "last_modified":
		switch val := value.(type) {
		case int:
			o.ModTime = riposo.Epoch(val)
		case int64:
			o.ModTime = riposo.Epoch(val)
		case riposo.Epoch:
			o.ModTime = val
		default:
			return errValNotNumber
		}
	case "deleted":
		switch val := value.(type) {
		case bool:
			o.Deleted = val
		default:
			return errValNotBool
		}
	default:
		bin, err := sjson.SetBytes(o.Extra, field, value)
		if err != nil {
			return err
		}
		o.Extra = bin
	}
	return nil
}

// DecodeExtra unmarshals extra Extra into a value.
func (o *Object) DecodeExtra(v interface{}) error {
	o.Norm()

	if err := json.Unmarshal(o.Extra, v); err != nil {
		if se := new(json.UnmarshalTypeError); errors.As(err, &se) {
			if se.Field == "" {
				se.Field = "data"
			} else {
				se.Field = "data." + se.Field
			}
		}
		return err
	}

	return nil
}

// EncodeExtra marshals v into extra Extra.
func (o *Object) EncodeExtra(v interface{}) error {
	buf := bytes.NewBuffer(o.Extra[:0])
	if err := json.NewEncoder(buf).Encode(v); err != nil {
		return err
	}

	o.Extra = bytes.TrimSuffix(buf.Bytes(), newLine)
	return nil
}

// MarshalJSON implements json.Marshaler interface.
func (o *Object) MarshalJSON() ([]byte, error) {
	core, err := json.Marshal(objectCore{
		ID:      o.ID,
		ModTime: o.ModTime,
		Deleted: o.Deleted,
	})
	if err != nil {
		return nil, err
	}

	if o.Deleted || len(o.Extra) < 3 {
		return core, nil
	}

	core[len(core)-1] = ','
	return append(core, o.Extra[1:]...), nil
}

// UnmarshalJSON implements custom JSON unmarshaler.
func (o *Object) UnmarshalJSON(p []byte) error {
	b := bytes.NewBuffer(p[:0])
	err := json.Compact(b, p)
	if err != nil {
		return err
	}
	p = b.Bytes()

	id := gjson.GetBytes(p, "id")
	if id.Raw != "" {
		if p, err = sjson.DeleteBytes(p, "id"); err != nil {
			return err
		}
	}

	modTime := gjson.GetBytes(p, "last_modified")
	if modTime.Raw != "" {
		if p, err = sjson.DeleteBytes(p, "last_modified"); err != nil {
			return err
		}
	}

	deleted := gjson.GetBytes(p, "deleted")
	if deleted.Raw != "" {
		if p, err = sjson.DeleteBytes(p, "deleted"); err != nil {
			return err
		}
	}

	if deleted.Bool() {
		p = nil
	}

	*o = Object{
		ID:      id.String(),
		ModTime: riposo.Epoch(modTime.Int()),
		Deleted: deleted.Bool(),
		Extra:   p,
	}
	return nil
}

// Copy creates a copy of the object.
func (o *Object) Copy() *Object {
	if o == nil {
		return nil
	}

	extra := make([]byte, len(o.Extra))
	copy(extra, o.Extra)

	return &Object{
		ID:      o.ID,
		ModTime: o.ModTime,
		Deleted: o.Deleted,
		Extra:   extra,
	}
}

// Update uses values of x to update o.
func (o *Object) Update(x *Object) {
	if x.ID != "" {
		o.ID = x.ID
	}
	if x.ModTime > o.ModTime {
		o.ModTime = x.ModTime
	}
	if x.Deleted {
		o.Deleted = x.Deleted
	}
	if len(x.Extra) != 0 {
		o.Extra = append(o.Extra[:0], x.Extra...)
	} else {
		o.Norm()
	}
}

// Patch merges attributes of x into o.
func (o *Object) Patch(x *Object) error {
	if len(x.Extra) < 3 {
		o.Norm()
		return nil
	}

	var m1, m2 map[string]interface{}
	if err := json.Unmarshal(x.Extra, &m2); err != nil {
		return err
	}

	if len(o.Extra) < 3 {
		m1 = m2
	} else {
		if err := json.Unmarshal(o.Extra, &m1); err != nil {
			return err
		}
		recPatch(m1, m2)
	}
	return o.EncodeExtra(m1)
}

// Norm normalises the object.
func (o *Object) Norm() {
	if len(o.Extra) == 0 {
		o.Extra = append(o.Extra, '{', '}')
	}
}

// String implements Stringer interface.
func (o *Object) String() string {
	bin, _ := o.MarshalJSON()
	return string(bin)
}

// ByteSize returns the size of the JSON encoded object in bytes.
func (o *Object) ByteSize() int64 {
	if o == nil {
		return 4
	}

	n := 2
	n += 7 + len(o.ID)               // "id":""
	n += 17 + sizeOfEpoch(o.ModTime) // ,"last_modified":
	if o.Deleted {
		n += 15 // ,"deleted":true
	}
	if m := len(o.Extra); m > 2 {
		n += m - 1
	}
	return int64(n)
}

func recPatch(o1, o2 map[string]interface{}) {
	for key, v2 := range o2 {
		if v2 == nil {
			continue
		}

		v1, ok := o1[key]
		if !ok || v1 == nil {
			o1[key] = v2
			continue
		}

		switch tv2 := v2.(type) {
		case map[string]interface{}:
			switch tv1 := v1.(type) {
			case map[string]interface{}:
				recPatch(tv1, tv2)
			default:
				o1[key] = tv2
			}
		default:
			o1[key] = tv2
		}
	}
}

func sizeOfEpoch(e riposo.Epoch) int {
	if e == 0 {
		return 1
	}

	n := 1
	if e < 0 {
		e = -e
		n++
	}
	return n + int(math.Log10(float64(e)))
}
