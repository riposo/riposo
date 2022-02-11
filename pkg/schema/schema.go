package schema

import "sort"

// PermissionSet contain a list of principals by type of access.
type PermissionSet map[string][]string

// Add adds a single principal to the set.
func (p PermissionSet) Add(perm, principal string) {
	perms := p[perm]
	for _, s := range perms {
		if s == principal {
			return
		}
	}

	perms = append(perms, principal)
	sort.Strings(perms)
	p[perm] = perms
}

// ByteSize returns the size of the JSON encoded permission set in bytes.
func (p PermissionSet) ByteSize() int64 {
	if p == nil {
		return 4
	}

	n := 2
	if m := len(p) - 1; m > 0 {
		n += m
	}

	for key, vals := range p {
		n += len(key) + 5

		if m := len(vals) - 1; m > 0 {
			n += m
		}

		for _, val := range vals {
			n += len(val) + 2
		}
	}
	return int64(n)
}

// Objects contains a slice of objects.
type Objects struct {
	Data []*Object `json:"data"`
}

// Resource contain a combination of object and permissions.
type Resource struct {
	StatusCode  int           `json:"-"`
	Data        *Object       `json:"data,omitempty"`
	Permissions PermissionSet `json:"permissions,omitempty"`
}

// HTTPStatus returns the http status code.
func (r *Resource) HTTPStatus() int { return r.StatusCode }

// ByteSize returns the size of the JSON encoded resource in bytes.
func (r *Resource) ByteSize() int64 {
	if r == nil {
		return 4
	}

	n := int64(2)
	if r.Data != nil {
		n += 7 + r.Data.ByteSize()
	}
	if r.Permissions != nil {
		n += 14 + r.Permissions.ByteSize()
	}
	if r.Data != nil && r.Permissions != nil {
		n++
	}
	return n
}
