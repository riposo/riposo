package storage

import (
	"regexp"
	"strings"

	"github.com/riposo/riposo/pkg/params"
	"github.com/riposo/riposo/pkg/schema"
	"github.com/tidwall/gjson"
)

type objectSlice struct {
	Slice []*schema.Object
	Sort  []params.SortOrder
}

func (s *objectSlice) Len() int      { return len(s.Slice) }
func (s *objectSlice) Swap(i, j int) { s.Slice[i], s.Slice[j] = s.Slice[j], s.Slice[i] }
func (s *objectSlice) Less(i, j int) bool {
	o1, o2 := s.Slice[i], s.Slice[j]
	for _, so := range s.Sort {
		if x := compare(o1.Get(so.Field), o2.Get(so.Field)); x != 0 {
			if so.Descending {
				return x > 0
			}
			return x < 0
		}
	}
	return false
}

func compare(v1, v2 schema.Value) int {
	if r1, r2 := rank(v1.Type), rank(v2.Type); r1 < r2 {
		return -1
	} else if r1 > r2 {
		return 1
	}

	switch v1.Type {
	case gjson.Null:
		return 0
	case gjson.String:
		return strings.Compare(v1.String(), v2.String())
	case gjson.Number:
		if f1, f2 := v1.Float(), v2.Float(); f1 < f2 {
			return -1
		} else if f1 > f2 {
			return 1
		} else {
			return 0
		}
	default:
		return strings.Compare(v1.Raw, v2.Raw)
	}
}

func isEqual(v1, v2 schema.Value, strictNULL bool) bool {
	if v1.Type != v2.Type {
		return false
	}

	switch v1.Type {
	case gjson.Number:
		return v1.Num == v2.Num
	case gjson.Null:
		return !strictNULL || v1.Raw == v2.Raw
	default:
		return v1.Raw == v2.Raw
	}
}

func rank(t gjson.Type) int {
	switch t {
	case gjson.Null:
		return 5
	case gjson.True:
		return 4
	case gjson.False:
		return 3
	case gjson.Number:
		return 2
	case gjson.String:
		return 1
	default:
		return 0
	}
}

// --------------------------------------------------------------------

func match(o *schema.Object, f params.Filter) bool {
	val := o.Get(f.Field)

	switch f.Operator {
	case params.OperatorGT:
		return compare(val, f.Value(0)) > 0
	case params.OperatorMIN:
		return compare(val, f.Value(0)) > -1
	case params.OperatorLT:
		return compare(val, f.Value(0)) < 0
	case params.OperatorMAX:
		return compare(val, f.Value(0)) < 1
	case params.OperatorEQ:
		return isEqual(val, f.Value(0), false)
	case params.OperatorNOT:
		return !isEqual(val, f.Value(0), false)
	case params.OperatorIN:
		for i := range f.Values {
			if isEqual(val, f.Value(i), false) {
				return true
			}
		}
		return false
	case params.OperatorEXCLUDE:
		for i := range f.Values {
			if isEqual(val, f.Value(i), false) {
				return false
			}
		}
		return true
	case params.OperatorHAS:
		exp, act := f.Value(0).Bool(), val.Exists()
		return (exp && act) || (!exp && !act)
	case params.OperatorLIKE:
		if f.Field == "last_modified" {
			return false
		}

		pat := strings.ReplaceAll(regexp.QuoteMeta(f.Value(0).String()), "\\*", ".*")
		if pat == "" {
			break
		}

		exp, err := regexp.Compile(pat)
		if err != nil {
			break
		}

		switch val.Type {
		case gjson.String:
			return exp.MatchString(val.String())
		case gjson.Number, gjson.True, gjson.False, gjson.JSON:
			return exp.MatchString(val.Raw)
		}
	}
	return false
}

func conditionMatch(o *schema.Object, cnd params.Condition) bool {
	for _, f := range cnd {
		if !match(o, f) {
			return false
		}
	}
	return true
}

func paginationMatch(o *schema.Object, cnds params.ConditionSet) bool {
	for _, cs := range cnds {
		if conditionMatch(o, cs) {
			return true
		}
	}
	return len(cnds) == 0
}

func paginationFilter(objs []*schema.Object, cnds params.ConditionSet) []*schema.Object {
	if cnds = cnds.Compact(); len(cnds) == 0 {
		return objs
	}

	res := objs[:0]
	for _, o := range objs {
		if paginationMatch(o, cnds) {
			res = append(res, o)
		}
	}
	return res
}
