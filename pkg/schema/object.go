package schema

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/riposo/riposo/pkg/riposo"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var newLine = []byte("\n")

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

// DecodeExtra unmarshals extra Extra into a value.
func (o *Object) DecodeExtra(v interface{}) error {
	if err := json.Unmarshal(o.Extra, v); err != nil {
		switch se := err.(type) {
		case *json.UnmarshalTypeError:
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
	}
}

// Patch merges attributes of x into o.
func (o *Object) Patch(x *Object) error {
	if len(x.Extra) < 3 {
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
