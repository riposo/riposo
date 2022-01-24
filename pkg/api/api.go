package api

import (
	"context"
	"net/http"
	"time"

	"github.com/riposo/riposo/pkg/riposo"
)

// GetPath extracts the resource path from the request.
func GetPath(req *http.Request) riposo.Path {
	return riposo.NormPath(req.URL.Path)
}

// User is the authenticated API user.
type User struct {
	ID         string
	Principals []string // extra principals
}

// IsAuthenticated reports true if the User is an authenticated user.
func (u *User) IsAuthenticated() bool {
	if u.ID != riposo.Everyone {
		for _, x := range u.Principals {
			if x == riposo.Authenticated {
				return true
			}
		}
	}
	return false
}

// Config holds API configuration values.
type Config struct {
	// Authz is a custom authorization guard.
	Authz Authz

	// Pagination configures resource pagination options.
	Pagination struct {
		TokenValidity time.Duration
		MaxLimit      int
	}
}

func (c *Config) norm() *Config {
	if c == nil {
		c = new(Config)
	}
	if c.Authz == nil {
		c.Authz = make(Authz)
	}
	if c.Pagination.TokenValidity == 0 {
		c.Pagination.TokenValidity = 10 * time.Minute
	}
	if c.Pagination.MaxLimit == 0 {
		c.Pagination.MaxLimit = 10_000
	}
	return c
}

// HandlerFunc is a simplified API handler function.
type HandlerFunc func(out http.Header, req *http.Request) interface{}

// ServeHTTP implements http.Handler interface.
func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res := f(w.Header(), r)
	Render(w, res)
}

// Component represents a component that can be initialised on boot and closed
// on shutdown.
type Component interface {
	Init(context.Context, *Routes, riposo.Helpers) error
	Close() error
}
