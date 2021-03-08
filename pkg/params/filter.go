package params

import (
	"strings"

	"github.com/riposo/riposo/pkg/schema"
	"github.com/riposo/riposo/pkg/util"
)

// Condition is a set of filters that form a logical conjunction,
// i.e. ALL filters match match to qualify.
type Condition []Filter

// ConditionSet is a set of conditions that form a logical disjunction,
// i.e. ANY of the conditions must match to qualify.
type ConditionSet []Condition

// Compact removes empty conditions from the set.
func (cs ConditionSet) Compact() ConditionSet {
	rs := cs[:0]
	for _, c := range cs {
		if len(c) != 0 {
			rs = append(rs, c)
		}
	}
	return rs
}

// Filter expresses a filterable condition.
type Filter struct {
	Field    string         // the field name
	Operator Operator       // the comparison operator
	Values   []schema.Value // slice of parsed values
}

// ParseFilter parses a filter from a field-value string pair.
func ParseFilter(field, value string) Filter {
	operator := OperatorEQ
	for _, ent := range prefixMap {
		if strings.HasPrefix(field, ent.Prefix) {
			field = strings.TrimPrefix(field, ent.Prefix)
			operator = ent.Operator
			break
		}
	}

	values := make([]schema.Value, 0, 1)
	switch operator {
	case OperatorIN, OperatorEXCLUDE, OperatorContainsAny:
		util.SplitFunc(value, ",", func(val string) {
			values = append(values, schema.ParseValue(val))
		})
	default:
		values = append(values, schema.ParseValue(value))
	}

	return Filter{Field: field, Operator: operator, Values: values}
}

func (f Filter) isValid() bool {
	return f.Field != "" && len(f.Values) > 0
}

// Value returns the value at index.
func (f Filter) Value(index int) schema.Value {
	if index < len(f.Values) {
		return f.Values[index]
	}
	return schema.Value{}
}
