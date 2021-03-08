package params

import "strings"

// SortOrder determines sorting order.
type SortOrder struct {
	Field      string
	Descending bool
}

// ParseSort parses a single sort parameter.
func ParseSort(value string) (sort []SortOrder) {
	for min, max := 0, 0; min < len(value); {
		if pos := strings.IndexByte(value[min:], ','); pos < 0 {
			max = len(value)
		} else {
			max = pos + min
		}

		if field := value[min:max]; len(field) != 0 && field[0] != '-' {
			sort = appendSO(sort, SortOrder{Field: field})
		} else if len(field) > 1 {
			sort = appendSO(sort, SortOrder{Field: field[1:], Descending: true})
		}
		min = max + 1
	}
	return
}

func appendSO(t []SortOrder, so SortOrder) []SortOrder {
	for _, x := range t {
		if x.Field == so.Field {
			return t
		}
	}
	return append(t, so)
}
