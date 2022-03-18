package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/riposo/riposo/internal/config"
	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/auth"
	"github.com/riposo/riposo/pkg/mock"
	"github.com/riposo/riposo/pkg/plugin"
	"github.com/riposo/riposo/pkg/schema"
)

// NewMux inits a new handler for tests.
func NewMux() http.Handler {
	hlp := mock.Helpers()
	cns := mock.Conns(hlp)

	cfg := new(config.Config)
	cfg.Project.Docs = "http://example.com"
	cfg.Project.Name = "Example Project"
	cfg.Project.Version = "0.11.2"
	cfg.Batch.MaxRequests = 25
	cfg.Pagination.MaxLimit = 3
	cfg.Pagination.TokenValidity = time.Hour
	cfg.EOS.Time = time.Unix(2424242424, 0)
	cfg.Capabilities = new(plugin.Set)
	cfg.Backoff.Duration = 60 * time.Second
	cfg.RetryAfter = 30 * time.Second

	rts := api.NewRoutes(cfg.APIConfig())
	rts.Resource("/buckets", nil)
	rts.Handle("/failure", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api.Render(w, schema.InternalError(fmt.Errorf("doh")))
	}))

	return newMux(rts, hlp, cns, mockAuth{}, cfg)
}

type mockAuth struct{}

func (mockAuth) Authenticate(r *http.Request) (*api.User, error) {
	if user, _, ok := r.BasicAuth(); ok {
		return mock.User("account:" + user), nil
	}
	return nil, auth.Errorf("missing credentials")
}

func (mockAuth) Close() error { return nil }
