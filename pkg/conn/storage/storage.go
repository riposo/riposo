// Package storage contains abstractions and implementations
// for an object backend.
package storage

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/riposo/riposo/pkg/params"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"
)

var (
	// ErrNotFound is returned when an object is not found.
	ErrNotFound = errors.New("object not found")
	// ErrObjectExists is returned when trying to create an object that already exists.
	ErrObjectExists = errors.New("object already exists")
	// ErrInvalidPath is returned when an invalid path is used for the method.
	ErrInvalidPath = errors.New("invalid path")
	// ErrTxDone is returned when a transaction has expired.
	ErrTxDone = errors.New("transaction has already been committed or rolled back")
)

// Backend defines the abstract storage interface.
type Backend interface {
	// Ping returns an error if offline.
	Ping(context.Context) error

	// Begin starts a new Transaction.
	Begin(context.Context) (Transaction, error)

	// Close closes the backend.
	Close() error
}

// Transaction is a transaction. Please note that transactions are not
// guaranteed to be thread-safe and must not be used across multiple goroutines.
type Transaction interface {
	// Commit commits the transaction.
	Commit() error
	// Rollback rolls back the transaction.
	Rollback() error

	// Flush removes every object from this backend.
	Flush() error
	// Purge purges all deleted objects olderThan epoch.
	Purge(olderThan riposo.Epoch) (int64, error)

	// ModTime returns the maximum epoch of the given path.
	ModTime(path riposo.Path) (riposo.Epoch, error)
	// ListAll appends matching objects within a path and returns the resulting slice.
	ListAll(objs []*schema.Object, path riposo.Path, opt ListOptions) ([]*schema.Object, error)
	// CountAll counts all matching objects within a path and returns the resulting number.
	CountAll(path riposo.Path, opt CountOptions) (int64, error)
	// DeleteAll recursively deletes given paths and returns the
	// maximum modTime and paths of the deleted objects.
	DeleteAll(paths []riposo.Path) (riposo.Epoch, []riposo.Path, error)

	// Exists returns true if a path exists.
	// Accepts elementary paths only.
	Exists(path riposo.Path) (bool, error)
	// Get returns a stored object. May return ErrNotFound.
	// Accepts elementary paths only.
	Get(path riposo.Path) (*schema.Object, error)
	// GetForUpdate returns a stored object with a lock. May return ErrNotFound.
	GetForUpdate(path riposo.Path) (*schema.Object, error)

	// Create stores a new object under a path.
	Create(path riposo.Path, obj *schema.Object) error
	// Update updates an existing object.
	Update(path riposo.Path, obj *schema.Object) error
	// Delete deletes an existing path and returns the affected object.
	Delete(path riposo.Path) (*schema.Object, error)
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
		return nil, fmt.Errorf("invalid storage URL %q", urlString)
	}

	factory, ok := registry[u.Scheme]
	if !ok {
		return nil, fmt.Errorf("unknown storage type %q", u.Scheme)
	}

	return factory(ctx, u, hlp)
}

// --------------------------------------------------------------------

// Inclusion allows to include objects by state.
type Inclusion uint8

// Inclusion enum.
const (
	IncludeLive Inclusion = iota
	IncludeAll
)

// CountOptions contain options for bulk-counting.
type CountOptions struct {
	// Condition are AND'ed.
	Condition params.Condition
}

// ListOptions contain options for bulk-listing.
type ListOptions struct {
	// Condition are AND'ed.
	Condition params.Condition
	// Pagination are are OR'd.
	Pagination params.ConditionSet
	// Include live/all objects in the output.
	Include Inclusion
	// Sort are applied sequentially.
	Sort []params.SortOrder
	// Limits the number of objects returned.
	Limit int
}
