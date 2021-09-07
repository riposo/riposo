package riposo

import "strings"

// Path represents an object path.
type Path string

// NormPath parses a path from a URL path.
func NormPath(path string) Path {
	path = trimPathPrefix(path)
	n := 0
	for _, r := range path {
		if r == '/' {
			n++
		}
	}

	if n%2 == 0 {
		return Path(path)
	}
	return Path(path + "/*")
}

// JoinPath joins namespace and object ID.
func JoinPath(namespace, objID string) Path {
	return NormPath(namespace + "/" + objID)
}

// Parent returns the parent path.
func (p Path) Parent() Path {
	var xx bool
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			if xx {
				return p[:i]
			}
			xx = true
		}
	}
	return ""
}

// WithObjectID replaces the objectID of the path and returns
// the result.
func (p Path) WithObjectID(objectID string) Path {
	s := string(p)
	if i := strings.LastIndexByte(s, '/'); i > -1 {
		s = s[:i]
	}
	return Path(s + "/" + objectID)
}

// Split splits the path into a namespace and object ID.
func (p Path) Split() (string, string) {
	s := string(p)
	if i := strings.LastIndexByte(s, '/'); i > -1 {
		return s[:i], s[i+1:]
	}
	return "", ""
}

// ObjectID extracts the object ID.
func (p Path) ObjectID() string {
	_, objID := p.Split()
	return objID
}

// ResourceName extracts the resource name.
func (p Path) ResourceName() string {
	s := p.namespace()
	if i := strings.LastIndexByte(s, '/'); i > -1 {
		s = s[i+1:]
	}
	return strings.TrimSuffix(s, "s")
}

// String returns the path as a plain string.
func (p Path) String() string {
	return string(p)
}

// Contains returns true if path contains other.
func (p Path) Contains(other Path) bool {
	if p.IsNode() {
		return p.namespace() == other.namespace()
	}
	return p == other
}

// Traverse iterates over path and its parents.
func (p Path) Traverse(fn func(Path)) {
	for {
		fn(p)
		if p == "" {
			break
		}
		p = p.Parent()
	}
}

// IsNode returns true if the path addresses multiple resources.
func (p Path) IsNode() bool {
	return strings.HasSuffix(string(p), "*")
}

func (p Path) namespace() string {
	ns, _ := p.Split()
	return ns
}

func trimPathPrefix(path string) string {
	if !strings.HasPrefix(path, "/v") {
		return path
	}

	for i, r := range path[2:] {
		if r == '/' {
			return path[2+i:]
		} else if r < '0' || r > '9' {
			return path
		}
	}
	return ""
}
