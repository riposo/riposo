package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/riposo"
)

// Method implementations are responsible for authenticating users
// from HTTP requests.
type Method interface {
	// Authenticate parses an API request and returns a user. If a user cannot authenticated it must
	// return an ErrUnauthenticated compatible error.
	Authenticate(*http.Request) (*api.User, error)

	// Close may release all resources held by the Method implementation.
	Close() error
}

// Factory initializes a Method at runtime.
type Factory func(context.Context, riposo.Helpers) (Method, error)

// --------------------------------------------------------------------

// ErrUnauthenticated defines the basic unauthenticated error.
var ErrUnauthenticated = errors.New("unauthenticated")

type unauthenticated struct{ error }

// Errorf indicates ErrUnauthenticated with a custom message.
func Errorf(message string, args ...interface{}) error {
	return &unauthenticated{error: fmt.Errorf(message, args...)}
}

// WrapError indicates ErrUnauthenticated and wraps an internal error.
func WrapError(err error) error {
	return &unauthenticated{error: err}
}

// Error implements error interface.
func (e *unauthenticated) Error() string { return e.error.Error() }

// Is implements errors interface.
func (e *unauthenticated) Is(err error) bool { return err == ErrUnauthenticated }

// --------------------------------------------------------------------

var (
	registry   = make(map[string]Factory)
	registryMu sync.RWMutex
)

// Register registers a new factory by name.
// It will panic if multiple factories are registered under the same name.
func Register(name string, factory Factory) {
	registryMu.Lock()
	defer registryMu.Unlock()

	_, ok := registry[name]
	if ok {
		panic("factory " + name + " is already registered")
	}
	registry[name] = factory
}

// Get returns a factory by name. Returns an error if not found.
func Get(name string) (Factory, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	if factory, ok := registry[name]; ok {
		return factory, nil
	}
	return nil, fmt.Errorf("unknown auth factory %q", name)
}
