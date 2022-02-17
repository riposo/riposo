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
	permission.Register("memory", func(context.Context, *url.URL, riposo.Helpers) (permission.Backend, error) {
		return New(), nil
	})
}

type backend struct {
	users map[string]util.Set
	perms map[riposo.Path]map[string]util.Set

	mu sync.Mutex
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
	b.mu.Lock()
	return &transaction{b: b}, nil
}

// Close implements permission.Backend interface.
func (*backend) Close() error {
	return nil
}

// --------------------------------------------------------------------

type transaction struct {
	b *backend

	xusers map[string]util.Set
	xperms map[riposo.Path]map[string]util.Set

	done, flushed bool
}

// Commit implements permission.Transaction interface.
func (t *transaction) Commit() error {
	if t.done {
		return permission.ErrTxDone
	}
	t.done = true
	defer t.b.mu.Unlock()

	return nil
}

// Rollback implements permission.Transaction interface.
func (t *transaction) Rollback() error {
	if t.done {
		return permission.ErrTxDone
	}
	t.done = true
	defer t.b.mu.Unlock()

	if t.flushed {
		t.b.users, t.b.perms = t.xusers, t.xperms
		return nil
	}

	for k, v := range t.xusers {
		if v == nil {
			delete(t.b.users, k)
		} else {
			t.b.users[k] = v
		}
	}
	for k, v := range t.xperms {
		t.b.perms[k] = v
	}

	return nil
}

// Flush implements permission.Transaction interface.
func (t *transaction) Flush() error {
	if t.done {
		return permission.ErrTxDone
	}

	t.xusers, t.xperms = t.b.users, t.b.perms
	t.flushed = true

	t.b.users = make(map[string]util.Set)
	t.b.perms = make(map[riposo.Path]map[string]util.Set)
	return nil
}

// GetUserPrincipals implements permission.Transaction interface.
func (t *transaction) GetUserPrincipals(userID string) ([]string, error) {
	if t.done {
		return nil, permission.ErrTxDone
	}

	set := t.b.users[userID].Copy()
	set.Add(userID)

	switch userID {
	case riposo.Everyone:
		// pass-through
	case riposo.Authenticated:
		set.Add(riposo.Everyone)
		set.Merge(t.b.users[riposo.Everyone])
	default:
		set.Add(riposo.Authenticated)
		set.Add(riposo.Everyone)
		set.Merge(t.b.users[riposo.Authenticated])
		set.Merge(t.b.users[riposo.Everyone])
	}
	return set.Slice(), nil
}

// AddUserPrincipal implements permission.Transaction interface.
func (t *transaction) AddUserPrincipal(principal string, userIDs []string) error {
	if t.done {
		return permission.ErrTxDone
	}

	for _, userID := range userIDs {
		t.backupUser(userID)

		set, ok := t.b.users[userID]
		if !ok {
			set := util.NewSet(principal)
			t.b.users[userID] = set
		} else {
			set.Add(principal)
		}
	}
	return nil
}

// RemoveUserPrincipal implements permission.Transaction interface.
func (t *transaction) RemoveUserPrincipal(principal string, userIDs []string) error {
	if t.done {
		return permission.ErrTxDone
	}

	for _, userID := range userIDs {
		if set, ok := t.b.users[userID]; ok {
			t.backupUser(userID)

			set.Remove(principal)
		}
	}
	return nil
}

// PurgeUserPrincipals implements permission.Transaction interface.
func (t *transaction) PurgeUserPrincipals(principals []string) error {
	if t.done {
		return permission.ErrTxDone
	}

	for _, principal := range principals {
		for userID, set := range t.b.users {
			t.backupUser(userID)

			set.Remove(principal)
		}
	}
	return nil
}

// GetACEPrincipals implements permission.Transaction interface.
func (t *transaction) GetACEPrincipals(ent permission.ACE) ([]string, error) {
	if t.done {
		return nil, permission.ErrTxDone
	}

	if perms, ok := t.b.perms[ent.Path]; ok {
		if set, ok := perms[ent.Perm]; ok {
			return set.Slice(), nil
		}
	}
	return nil, nil
}

// AddACEPrincipal implements permission.Transaction interface.
func (t *transaction) AddACEPrincipal(principal string, ent permission.ACE) error {
	if t.done {
		return permission.ErrTxDone
	}

	t.backupPerms(ent.Path)

	perms, ok := t.b.perms[ent.Path]
	if !ok {
		perms = make(map[string]util.Set)
		t.b.perms[ent.Path] = perms
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
func (t *transaction) RemoveACEPrincipal(principal string, ent permission.ACE) error {
	if t.done {
		return permission.ErrTxDone
	}

	if perms, ok := t.b.perms[ent.Path]; ok {
		t.backupPerms(ent.Path)

		if set, ok := perms[ent.Perm]; ok {
			set.Remove(principal)
		}
	}
	return nil
}

// GetAllACEPrincipals implements permission.Transaction interface.
func (t *transaction) GetAllACEPrincipals(ents []permission.ACE) ([]string, error) {
	if t.done {
		return nil, permission.ErrTxDone
	}

	if len(ents) == 0 {
		return nil, nil
	}

	res := util.NewSet()
	for _, ent := range ents {
		if perms, ok := t.b.perms[ent.Path]; ok {
			if set, ok := perms[ent.Perm]; ok {
				res.Merge(set)
			}
		}
	}
	return res.Slice(), nil
}

// GetPermissions implements permission.Transaction interface.
func (t *transaction) GetPermissions(path riposo.Path) (schema.PermissionSet, error) {
	if t.done {
		return nil, permission.ErrTxDone
	}

	perms := make(schema.PermissionSet, len(t.b.perms[path]))
	for perm, set := range t.b.perms[path] {
		perms[perm] = set.Slice()
	}
	return perms, nil
}

// CreatePermissions implements permission.Transaction interface.
func (t *transaction) CreatePermissions(path riposo.Path, set schema.PermissionSet) error {
	if t.done {
		return permission.ErrTxDone
	}

	t.backupPerms(path)

	perms, ok := t.b.perms[path]
	if !ok {
		perms = make(map[string]util.Set, len(set))
		t.b.perms[path] = perms
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
func (t *transaction) MergePermissions(path riposo.Path, set schema.PermissionSet) error {
	if t.done {
		return permission.ErrTxDone
	}

	t.backupPerms(path)

	perms, ok := t.b.perms[path]
	if !ok {
		perms = make(map[string]util.Set, len(set))
		t.b.perms[path] = perms
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
func (t *transaction) DeletePermissions(paths []riposo.Path) error {
	if t.done {
		return permission.ErrTxDone
	}

	if len(paths) == 0 {
		return nil
	}

	for path := range t.b.perms {
		for _, pattern := range paths {
			if pattern == path || strings.HasPrefix(path.String(), pattern.String()+"/") {
				t.backupPerms(path)
				delete(t.b.perms, path)
			}
		}
	}
	return nil
}

// GetAccessiblePaths implements permission.Transaction interface.
func (t *transaction) GetAccessiblePaths(dst []riposo.Path, principals []string, ents []permission.ACE) ([]riposo.Path, error) {
	if t.done {
		return nil, permission.ErrTxDone
	}

	if len(principals) == 0 || len(ents) == 0 {
		return nil, nil
	}

	for path, perms := range t.b.perms {
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

func (t *transaction) backupUser(userID string) {
	if t.flushed {
		return
	}

	if t.xusers == nil {
		t.xusers = make(map[string]util.Set)
	}

	if _, ok := t.xusers[userID]; !ok {
		t.xusers[userID] = t.b.users[userID].Copy()
	}
}

func (t *transaction) backupPerms(path riposo.Path) {
	if t.flushed {
		return
	}

	if t.xperms == nil {
		t.xperms = make(map[riposo.Path]map[string]util.Set)
	}

	if _, ok := t.xperms[path]; !ok {
		pristine := t.b.perms[path]
		t.xperms[path] = make(map[string]util.Set, len(pristine))
		for k, v := range pristine {
			t.xperms[path][k] = v.Copy()
		}
	}
}

func match(ents []permission.ACE, path riposo.Path, perm string) bool {
	for _, ent := range ents {
		if ent.Perm == perm && ent.Path.Contains(path) {
			return true
		}
	}
	return false
}
