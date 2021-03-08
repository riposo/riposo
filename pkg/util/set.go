package util

import (
	"encoding/json"
	"sort"
)

// Set is a simple string set.
type Set map[string]struct{}

// NewSet inits a new set with optional vals.
func NewSet(vals ...string) Set {
	s := make(Set, len(vals))
	for _, v := range vals {
		s.Add(v)
	}
	return s
}

// NewUnion creates a union of sets.
func NewUnion(sets ...Set) Set {
	n := 0
	for _, s := range sets {
		if m := len(s); m > n {
			n = m
		}
	}

	us := make(Set, n)
	for _, s := range sets {
		us.Merge(s)
	}
	return us
}

// Copy returns a copy of the Set.
func (s Set) Copy() Set {
	t := make(Set, len(s))
	t.Merge(s)
	return t
}

// Len returns set length.
func (s Set) Len() int {
	return len(s)
}

// Slice converts the set to a string slice.
func (s Set) Slice() []string {
	if len(s) == 0 {
		return nil
	}

	vv := make([]string, 0, len(s))
	for v := range s {
		vv = append(vv, v)
	}
	sort.Strings(vv)
	return vv
}

// Merge adds all values of t to s.
func (s Set) Merge(t Set) {
	for v := range t {
		s.Add(v)
	}
}

// MergeSlice adds all values of t to s.
func (s Set) MergeSlice(t []string) {
	for _, v := range t {
		s.Add(v)
	}
}

// Add adds a single value.
func (s Set) Add(v string) {
	s[v] = struct{}{}
}

// Remove removes a single value.
func (s Set) Remove(v string) {
	delete(s, v)
}

// Has checks for inclusion.
func (s Set) Has(v string) bool {
	_, ok := s[v]
	return ok
}

// HasAny checks for inclusion of any vals.
func (s Set) HasAny(vals ...string) bool {
	for _, v := range vals {
		if s.Has(v) {
			return true
		}
	}
	return false
}

// IntersectsWith checks is s intersects with t.
func (s Set) IntersectsWith(t Set) bool {
	if len(t) < len(s) {
		return t.IntersectsWith(s)
	}

	for v := range s {
		if t.Has(v) {
			return true
		}
	}
	return false
}

// MarshalJSON marshals Set to JSON.
func (s Set) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Slice())
}

// UnmarshalJSON unmarshals JSON to a Set.
func (s *Set) UnmarshalJSON(p []byte) error {
	if len(p) != 0 && p[0] == '{' {
		err := json.Unmarshal(p, (*map[string]struct{})(s))
		return err
	}

	var vals []string
	if err := json.Unmarshal(p, &vals); err != nil {
		return err
	}
	*s = NewSet(vals...)
	return nil
}
