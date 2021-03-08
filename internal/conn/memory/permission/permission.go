package permission

import (
	"context"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/riposo/riposo/pkg/conn/permission"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"
	"github.com/riposo/riposo/pkg/util"
)

func init() {
	permission.Register("memory", func(context.Context, *url.URL, *riposo.Helpers) (permission.Backend, error) {
		return New(), nil
	})
}

type backend struct {
	users map[string]util.Set
	perms map[riposo.Path]map[string]util.Set
	mu    sync.RWMutex
}

// New inits a new in-memory permission backend. Please use for development and testing only!
func New() permission.Backend {
	return &backend{
		users: make(map[string]util.Set),
		perms: make(map[riposo.Path]map[string]util.Set),
	}
}

// Ping implements permission.Backend interface.
func (*backend) Ping(_ context.Context) error {
	return nil
}

// Begin implements permission.Backend interface.
func (b *backend) Begin(_ context.Context) (permission.Transaction, error) {
	return b, nil
}

// Close implements permission.Backend interface.
func (*backend) Close() error {
	return nil
}

// Commit implements permission.Transaction interface.
func (*backend) Commit() error { return nil }

// Rollback implements permission.Transaction interface.
func (*backend) Rollback() error { return nil }

// Flush implements permission.Transaction interface.
func (b *backend) Flush() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.users = make(map[string]util.Set)
	b.perms = make(map[riposo.Path]map[string]util.Set)
	return nil
}

// GetUserPrincipals implements permission.Transaction interface.
func (b *backend) GetUserPrincipals(userID string) ([]string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	set := b.users[userID].Copy()
	set.Add(userID)

	switch userID {
	case riposo.Everyone:
		// pass-through
	case riposo.Authenticated:
		set.Add(riposo.Everyone)
		set.Merge(b.users[riposo.Everyone])
	default:
		set.Add(riposo.Authenticated)
		set.Add(riposo.Everyone)
		set.Merge(b.users[riposo.Authenticated])
		set.Merge(b.users[riposo.Everyone])
	}
	return set.Slice(), nil
}

// AddUserPrincipal implements permission.Transaction interface.
func (b *backend) AddUserPrincipal(principal string, userIDs []string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, userID := range userIDs {
		set, ok := b.users[userID]
		if !ok {
			set := util.NewSet(principal)
			b.users[userID] = set
		} else {
			set.Add(principal)
		}
	}
	return nil
}

// RemoveUserPrincipal implements permission.Transaction interface.
func (b *backend) RemoveUserPrincipal(principal string, userIDs []string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, userID := range userIDs {
		if set, ok := b.users[userID]; ok {
			set.Remove(principal)
		}
	}
	return nil
}

// PurgeUserPrincipals implements permission.Transaction interface.
func (b *backend) PurgeUserPrincipals(principals ...string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, principal := range principals {
		for _, set := range b.users {
			set.Remove(principal)
		}
	}
	return nil
}

// GetACEPrincipals implements permission.Transaction interface.
func (b *backend) GetACEPrincipals(ent permission.ACE) ([]string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if perms, ok := b.perms[ent.Path]; ok {
		if set, ok := perms[ent.Perm]; ok {
			return set.Slice(), nil
		}
	}
	return nil, nil
}

// AddACEPrincipal implements permission.Transaction interface.
func (b *backend) AddACEPrincipal(principal string, ent permission.ACE) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	perms, ok := b.perms[ent.Path]
	if !ok {
		perms = make(map[string]util.Set)
		b.perms[ent.Path] = perms
	}

	set, ok := perms[ent.Perm]
	if !ok {
		perms[ent.Perm] = util.NewSet(principal)
	} else {
		set.Add(principal)
	}
	return nil
}

// RemoveACEPrincipal implements permission.Transaction interface.
func (b *backend) RemoveACEPrincipal(principal string, ent permission.ACE) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if perms, ok := b.perms[ent.Path]; ok {
		if set, ok := perms[ent.Perm]; ok {
			set.Remove(principal)
		}
	}
	return nil
}

// GetAllACEPrincipals implements permission.Transaction interface.
func (b *backend) GetAllACEPrincipals(ents []permission.ACE) ([]string, error) {
	if len(ents) == 0 {
		return nil, nil
	}

	res := util.NewSet()

	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, ent := range ents {
		if perms, ok := b.perms[ent.Path]; ok {
			if set, ok := perms[ent.Perm]; ok {
				res.Merge(set)
			}
		}
	}
	return res.Slice(), nil
}

// GetPermissions implements permission.Transaction interface.
func (b *backend) GetPermissions(path riposo.Path) (schema.PermissionSet, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	perms := make(schema.PermissionSet, len(b.perms[path]))
	for perm, set := range b.perms[path] {
		perms[perm] = set.Slice()
	}
	return perms, nil
}

// CreatePermissions implements permission.Transaction interface.
func (b *backend) CreatePermissions(path riposo.Path, set schema.PermissionSet) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	perms, ok := b.perms[path]
	if !ok {
		perms = make(map[string]util.Set, len(set))
		b.perms[path] = perms
	}

	for perm, principals := range set {
		pset, ok := perms[perm]
		if !ok {
			pset = util.NewSet()
			perms[perm] = pset
		}
		for _, principal := range principals {
			pset.Add(principal)
		}
	}
	return nil
}

// MergePermissions implements permission.Transaction interface.
func (b *backend) MergePermissions(path riposo.Path, set schema.PermissionSet) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	perms, ok := b.perms[path]
	if !ok {
		perms = make(map[string]util.Set, len(set))
		b.perms[path] = perms
	}
	for perm, principals := range set {
		if len(principals) == 0 {
			delete(perms, perm)
		} else {
			perms[perm] = util.NewSet(principals...)
		}
	}
	return nil
}

// DeletePermissions implements permission.Transaction interface.
func (b *backend) DeletePermissions(paths ...riposo.Path) error {
	if len(paths) == 0 {
		return nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for path := range b.perms {
		for _, pattern := range paths {
			if pattern == path || strings.HasPrefix(path.String(), pattern.String()+"/") {
				delete(b.perms, path)
			}
		}
	}
	return nil
}

// GetAccessiblePaths implements permission.Transaction interface.
func (b *backend) GetAccessiblePaths(dst []riposo.Path, principals []string, ents []permission.ACE) ([]riposo.Path, error) {
	if len(principals) == 0 || len(ents) == 0 {
		return nil, nil
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for path, perms := range b.perms {
		for perm, allowed := range perms {
			if match(ents, path, perm) && allowed.HasAny(principals...) {
				dst = append(dst, path)
				break
			}
		}
	}

	sort.Slice(dst, func(i, j int) bool { return dst[i] < dst[j] })
	return dst, nil
}

func match(ents []permission.ACE, path riposo.Path, perm string) bool {
	for _, ent := range ents {
		if ent.Perm == perm && ent.Path.Contains(path) {
			return true
		}
	}
	return false
}
