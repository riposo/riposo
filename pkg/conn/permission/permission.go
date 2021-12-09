// Package permission contains abstractions and implementations
// for a permission backend.
package permission

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"
)

// ACE is a permission/path tuple.
type ACE struct {
	Perm string      // the permission name
	Path riposo.Path // the object path
}

// Backend defines the abstract backend.
type Backend interface {
	// Ping returns an error if offline.
	Ping(context.Context) error

	// Begin starts a new Transation.
	Begin(context.Context) (Transaction, error)

	// Close closes the backend.
	Close() error
}

// Transaction is a transaction.
type Transaction interface {
	// Commit commits the transaction.
	Commit() error
	// Rollback aborts the transaction.
	Rollback() error

	// Flush deletes all stored data.
	Flush() error

	// GetUserPrincipals returns all principals assigned to a user.
	GetUserPrincipals(userID string) ([]string, error)
	// AddUserPrincipal adds a principal to users.
	AddUserPrincipal(principal string, userIDs []string) error
	// RemoveUserPrincipal removes a principal from users.
	RemoveUserPrincipal(principal string, userIDs []string) error
	// PurgeUserPrincipals removes principals from every user.
	PurgeUserPrincipals(principals []string) error

	// GetACEPrincipals returns a list of principals for an Access Control Entry.
	GetACEPrincipals(ent ACE) ([]string, error)
	// AddACEPrincipal adds an additional principal to an Access Control Entry.
	AddACEPrincipal(principal string, ent ACE) error
	// RemoveACEPrincipal deletes a principal from an Access Control Entry.
	RemoveACEPrincipal(principal string, ent ACE) error
	// GetAllACEPrincipals returns principals with access to the requested ents.
	GetAllACEPrincipals(ents []ACE) ([]string, error)

	// GetPermissions gets all permissions for a single path.
	GetPermissions(path riposo.Path) (schema.PermissionSet, error)
	// CreatePermissions creates permissions of a single path.
	CreatePermissions(path riposo.Path, set schema.PermissionSet) error
	// MergePermissions merges permissions of a single path.
	MergePermissions(path riposo.Path, set schema.PermissionSet) error
	// DeletePermissions recursively deletes for the given paths.
	DeletePermissions(paths []riposo.Path) error

	// GetAccessiblePaths appends paths to dst that are accessible by principals within ents.
	// ACE paths may contains wildcards.
	//
	// Example: get all readable or writable paths by "account:alice" matching "/buckets/foo/collections/*":
	// 	backend.GetAccessiblePaths(ctx, nil, []string{"account:alice"},	[]permission.ACE{
	// 		{Perm: "read", Path: "/buckets/foo/collections/*"},
	// 		{Perm: "write", Path: "/buckets/foo/collections/*"},
	// 	})
	GetAccessiblePaths(dst []riposo.Path, principals []string, ents []ACE) ([]riposo.Path, error)
}

var (
	registry   = make(map[string]Factory)
	registryMu sync.RWMutex
)

// Factory initializes a new backend at runtime.
type Factory func(context.Context, *url.URL, riposo.Helpers) (Backend, error)

// Register registers a new backend by scheme.
// It will panic if multiple backends are registered under the same scheme.
func Register(scheme string, factory Factory) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if _, ok := registry[scheme]; ok {
		panic("scheme " + scheme + " is already registered")
	}
	registry[scheme] = factory
}

// Connect connects a backend via URL.
func Connect(ctx context.Context, urlString string, hlp riposo.Helpers) (Backend, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	u, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("invalid permission URL %q", urlString)
	}

	factory, ok := registry[u.Scheme]
	if !ok {
		return nil, fmt.Errorf("unknown permission type %q", u.Scheme)
	}

	return factory(ctx, u, hlp)
}
