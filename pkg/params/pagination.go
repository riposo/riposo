package params

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/riposo/riposo/pkg/schema"
)

var ErrNoToken = errors.New("no pagination token")

// Pagination is a decoded token.
type Pagination struct {
	Nonce   string                  `json:"nonce,omitempty"`
	LastObj map[string]schema.Value `json:"last_object,omitempty"`
}

// ParseToken parses a token from a string.
// It may return ErrNoToken if there is no pagination token.
func ParseToken(s string) (*Pagination, error) {
	if s == "" {
		return nil, ErrNoToken
	}

	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	var t *Pagination
	if err := json.Unmarshal(raw, &t); err != nil {
		return nil, err
	}

	return t, nil
}

// Encode encodes a pagination token as an URL-safe base64 string.
func (t *Pagination) Encode() (string, error) {
	if t == nil {
		return "", nil
	}

	raw, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// Conditions constructs a ConditionSet from the received
// Pagination token.
func (t *Pagination) Conditions() ConditionSet {
	if t == nil || len(t.LastObj) == 0 {
		return nil
	}

	fields := make([]field, 0, len(t.LastObj))
	for name, val := range t.LastObj {
		fields = append(fields, field{Name: name, Value: val})
	}

	conds := make(ConditionSet, 0, len(fields))
	for p := 0; p < len(fields); p++ {
		cond := make(Condition, 0, len(fields))
		for i, fv := range fields {
			op := OperatorEQ
			if i == p {
				op = OperatorGT
			}
			cond = append(cond, Filter{
				Field:    fv.Name,
				Operator: op,
				Values:   []schema.Value{fv.Value},
			})
		}
		conds = append(conds, cond)
	}
	return conds
}

func newPagination(nonce string, lastObj *schema.Object, sort []SortOrder) *Pagination {
	t := &Pagination{
		Nonce:   nonce,
		LastObj: make(map[string]schema.Value, len(sort)),
	}
	for _, s := range sort {
		if v := lastObj.Get(s.Field); !v.IsNull() {
			t.LastObj[s.Field] = v
		}
	}
	return t
}
