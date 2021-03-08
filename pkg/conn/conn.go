package conn

import (
	"context"

	"github.com/riposo/riposo/pkg/conn/cache"
	"github.com/riposo/riposo/pkg/conn/permission"
	"github.com/riposo/riposo/pkg/conn/storage"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"
	"go.uber.org/multierr"
)

// Set exposes connections.
type Set struct {
	store storage.Backend
	perms permission.Backend
	cache cache.Backend
}

// Use wraps existing backend connections.
func Use(store storage.Backend, perms permission.Backend, cache cache.Backend) *Set {
	return &Set{
		store: store,
		perms: perms,
		cache: cache,
	}
}

// Connect connects to all backends.
func Connect(ctx context.Context, storeURL, permsURL, cacheURL string, hlp *riposo.Helpers) (*Set, error) {
	store, err := storage.Connect(ctx, storeURL, hlp)
	if err != nil {
		return nil, err
	}

	perms, err := permission.Connect(ctx, permsURL, hlp)
	if err != nil {
		_ = store.Close()
		return nil, err
	}

	cache, err := cache.Connect(ctx, cacheURL, hlp)
	if err != nil {
		_ = store.Close()
		_ = perms.Close()
		return nil, err
	}

	return &Set{
		store: store,
		perms: perms,
		cache: cache,
	}, nil
}

// Store returns the main storage backend.
func (s *Set) Store() storage.Backend { return s.store }

// Perms returns the permission backend.
func (s *Set) Perms() permission.Backend { return s.perms }

// Cache returns the cache backend.
func (s *Set) Cache() cache.Backend { return s.cache }

// Close closes all connections
func (s *Set) Close() error {
	return multierr.Combine(
		s.store.Close(),
		s.perms.Close(),
		s.cache.Close(),
	)
}

// Heartbeat returns heartbeat info.
func (s *Set) Heartbeat(ctx context.Context) *schema.Heartbeat {
	return &schema.Heartbeat{
		Storage:    s.store.Ping(ctx) == nil,
		Permission: s.perms.Ping(ctx) == nil,
		Cache:      s.cache.Ping(ctx) == nil,
	}
}
