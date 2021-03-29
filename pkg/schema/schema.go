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
