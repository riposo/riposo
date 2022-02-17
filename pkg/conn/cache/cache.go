// Package cache contains abstractions and implementations
// for a cache backend.
package cache

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/riposo/riposo/pkg/riposo"
)

var (
	// ErrNotFound is returned when an key is not found.
	ErrNotFound = errors.New("key not found")

	// ErrTxDone is returned when a transaction has expired.
	ErrTxDone = errors.New("transaction has already been committed or rolled back")

	// errInvalidKey is returned when an key is invalid.
	errInvalidKey = errors.New("key is invalid")
)

// Backend defines the abstract backend.
type Backend interface {
	// Ping returns an error if offline.
	Ping(context.Context) error

	// Begin starts a new Transation.
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

	// Flush deletes all stored data.
	Flush() error

	// Get retrieves a key. May return ErrNotFound.
	Get(key string) ([]byte, error)
	// Set sets a value for key.
	Set(key string, val []byte, exp time.Time) error
	// Del deletes a key. May return ErrNotFound.
	Del(key string) error
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
		return nil, fmt.Errorf("invalid cache URL %q", urlString)
	}

	factory, ok := registry[u.Scheme]
	if !ok {
		return nil, fmt.Errorf("unknown cache type %q", u.Scheme)
	}

	return factory(ctx, u, hlp)
}

// ValidateKey returns an error if key is invalid.
func ValidateKey(key string) error {
	if n := utf8.RuneCountInString(key); n == 0 || n > 256 {
		return errInvalidKey
	}
	return nil
}
