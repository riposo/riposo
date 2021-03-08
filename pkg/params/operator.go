package params

// Operator is an enum type.
type Operator uint8

// Operator enum values.
const (
	OperatorLT Operator = iota + 1
	OperatorMIN
	OperatorMAX
	OperatorNOT
	OperatorEQ
	OperatorGT
	OperatorIN
	OperatorEXCLUDE
	OperatorLIKE
	OperatorHAS
	OperatorContainsAny
	OperatorContains
)

var prefixMap = []struct {
	Operator
	Prefix string
}{
	{Operator: OperatorHAS, Prefix: "has_"},
	{Operator: OperatorMIN, Prefix: "min_"},
	{Operator: OperatorMAX, Prefix: "max_"},
	{Operator: OperatorLT, Prefix: "lt_"},
	{Operator: OperatorGT, Prefix: "gt_"},
	{Operator: OperatorEQ, Prefix: "eq_"},
	{Operator: OperatorNOT, Prefix: "not_"},
	{Operator: OperatorIN, Prefix: "in_"},
	{Operator: OperatorLIKE, Prefix: "like_"},
	{Operator: OperatorEXCLUDE, Prefix: "exclude_"},
	{Operator: OperatorContainsAny, Prefix: "contains_any_"},
	{Operator: OperatorContains, Prefix: "contains_"},
}
