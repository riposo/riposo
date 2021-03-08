package storage

import (
	"strings"

	"github.com/bsm/minisql"
	"github.com/lib/pq"
	"github.com/riposo/riposo/pkg/conn/storage"
	"github.com/riposo/riposo/pkg/params"
	"github.com/riposo/riposo/pkg/util"
)

type queryBuilder struct {
	*minisql.Query
	hasWhere bool
}

func newQueryBuilder() *queryBuilder {
	return &queryBuilder{
		Query: minisql.Pooled(),
	}
}

func (b *queryBuilder) Release() {
	minisql.Release(b.Query)
}

func (b *queryBuilder) Limit(n int) {
	if n > 0 {
		b.AppendString(" LIMIT ")
		b.AppendInt(int64(n))
	}
}

func (b *queryBuilder) OrderBy(order []params.SortOrder) {
	if len(order) == 0 {
		return
	}

	b.AppendString(" ORDER BY ")
	for i, so := range order {
		if i != 0 {
			b.AppendString(", ")
		}

		switch so.Field {
		case "id", "last_modified":
			b.AppendString(so.Field)
		default:
			b.AppendString("data")
			util.SplitFunc(so.Field, ".", func(attr string) {
				b.AppendString("->")
				b.AppendValue(attr)
			})
		}
		if so.Descending {
			b.AppendString(` DESC`)
		} else {
			b.AppendString(` ASC`)
		}
	}
}

func (b *queryBuilder) Where(str string) {
	b.where()
	b.AppendString(str)
}

func (b *queryBuilder) InclusionFilter(status storage.Inclusion) {
	switch status {
	case storage.IncludeLive:
		b.Where(`NOT deleted`)
	}
}

func (b *queryBuilder) ConditionFilter(cond params.Condition) {
	if len(cond) == 0 {
		return
	}

	b.where()
	b.condition(cond)
}

func (b *queryBuilder) PaginationFilter(conds params.ConditionSet) {
	if conds = conds.Compact(); len(conds) == 0 {
		return
	}

	b.where()
	b.AppendString("( ")
	for i, cond := range conds {
		if i != 0 {
			b.AppendString(" OR ")
		}
		b.condition(cond)
	}
	b.AppendString(" )")
}

func (b *queryBuilder) where() {
	if b.hasWhere {
		b.AppendString(" AND ")
	} else {
		b.AppendString(" WHERE ")
		b.hasWhere = true
	}
}

func (b *queryBuilder) condition(cond params.Condition) {
	b.AppendString("( ")
	for i, flt := range cond {
		if i != 0 {
			b.AppendString(" AND ")
		}
		b.filter(flt)
	}
	b.AppendString(" )")
}

func (b *queryBuilder) filter(flt params.Filter) {
	var isNull bool
	for _, v := range flt.Values {
		if v.IsNull() {
			isNull = true
			break
		}
	}

	switch flt.Operator {
	case params.OperatorHAS:
		b.filterField(flt)
		if flt.Value(0).Bool() {
			b.AppendString(" IS NOT NULL")
		} else {
			b.AppendString(" IS NULL")
		}
	case params.OperatorEQ:
		if isNull {
			b.filterField(flt)
			b.AppendString(" IS NULL")
		} else {
			b.filterComparison(flt, " = ", true)
		}
	case params.OperatorNOT:
		if isNull {
			b.filterField(flt)
			b.AppendString(" IS NOT NULL")
		} else {
			b.filterComparison(flt, " != ", false)
		}
	case params.OperatorLIKE:
		if isNull || flt.Field == "last_modified" {
			b.AppendString("FALSE")
		} else {
			b.filterComparison(flt, " ILIKE ", true)
		}
	case params.OperatorGT:
		if isNull {
			b.AppendString("FALSE") // nothing can be > NULL
		} else {
			b.filterComparison(flt, " > ", false)
		}
	case params.OperatorLT:
		if isNull {
			b.filterField(flt)
			b.AppendString(" IS NOT NULL") // everything NOT NULL is < NULL
		} else {
			b.filterComparison(flt, " < ", true)
		}
	case params.OperatorMIN:
		if isNull {
			b.filterField(flt)
			b.AppendString(" IS NULL") // only NULL is <= NULL
		} else {
			b.filterComparison(flt, " >= ", false)
		}
	case params.OperatorMAX:
		if isNull {
			b.AppendString("TRUE") // everything is <= NULL
		} else {
			b.filterComparison(flt, " <= ", true)
		}
	case params.OperatorIN:
		if len(flt.Values) == 0 {
			b.AppendString("FALSE")
		} else {
			b.filterComparison(flt, " IN ", !isNull)
		}
	case params.OperatorEXCLUDE:
		if len(flt.Values) == 0 {
			b.AppendString("TRUE")
		} else {
			b.filterComparison(flt, " NOT IN ", isNull)
		}
	case params.OperatorContains:
		if isDataQuery(flt.Field) {
			b.filterComparison(flt, " @> ", true)
		} else {
			b.AppendString("FALSE")
		}
	case params.OperatorContainsAny:
		if isDataQuery(flt.Field) {
			b.filterComparison(flt, " && ", true)
		} else {
			b.AppendString("FALSE")
		}
	}
}

func (b *queryBuilder) filterComparison(flt params.Filter, opstr string, required bool) {
	if isDataQuery(flt.Field) {
		b.AppendString("(")
	}

	if flt.Operator == params.OperatorContainsAny {
		b.AppendString("jsonb_typeof(")
		b.filterField(flt)
		b.AppendString(") = 'array' AND (SELECT array_agg(elems) FROM jsonb_array_elements(")
		b.filterField(flt)
		b.AppendString(") elems)")
	} else {
		b.filterField(flt)
	}

	b.AppendString(opstr)
	b.filterValue(flt)

	if isDataQuery(flt.Field) {
		if required {
			b.AppendString(" AND ")
			b.filterField(flt)
			b.AppendString(" IS NOT NULL)")
		} else {
			b.AppendString(" OR ")
			b.filterField(flt)
			b.AppendString(" IS NULL)")
		}
	}
}

func (b *queryBuilder) filterField(flt params.Filter) {
	switch flt.Field {
	case "id", "last_modified":
		b.AppendString(flt.Field)
	default:
		b.AppendString("data")
		num := strings.Count(flt.Field, ".")
		util.SplitFunc(flt.Field, ".", func(attr string) {
			if num--; num == -1 && flt.Operator == params.OperatorLIKE {
				b.AppendString("->>")
			} else {
				b.AppendString("->")
			}
			b.AppendValue(attr)
		})
	}
}

func (b *queryBuilder) filterValue(flt params.Filter) {
	switch flt.Operator {
	case params.OperatorLIKE:
		if s := flt.Value(0).String(); strings.ContainsRune(s, '*') {
			b.AppendValue(strings.ReplaceAll(s, "*", "%"))
		} else {
			b.AppendValue("%" + s + "%")
		}
		return
	case params.OperatorContainsAny:
		strs := make([]string, len(flt.Values))
		for i, v := range flt.Values {
			strs[i] = v.Raw
		}
		b.AppendString("(")
		b.AppendValue(pq.StringArray(strs))
		b.AppendString(")::jsonb[]")
		return
	case params.OperatorIN, params.OperatorEXCLUDE:
		b.AppendString("(")
	}

	for i := range flt.Values {
		if i != 0 {
			b.AppendString(", ")
		}

		switch flt.Field {
		case "id":
			b.AppendValue(flt.Value(i).String())
		case "last_modified":
			b.AppendValue(flt.Value(i).Int())
		default:
			b.AppendValue(flt.Value(i).Raw)
		}
	}

	switch flt.Operator {
	case params.OperatorIN, params.OperatorEXCLUDE:
		b.AppendString(")")
	}
}

func isDataQuery(field string) bool {
	return field != "id" && field != "last_modified"
}
