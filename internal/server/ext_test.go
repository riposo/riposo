package server

import (
	"net/http"
	"time"

	"github.com/riposo/riposo/internal/config"
	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/auth"
	"github.com/riposo/riposo/pkg/mock"
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
	cfg.Capabilities = make(map[string]map[string]interface{})

	rts := api.NewRoutes(cfg.APIConfig())
	rts.Resource("/buckets", api.StdModel())

	return newMux(rts, cns, hlp, cfg, mockAuth{})
}

type mockAuth struct{}

func (mockAuth) Authenticate(r *http.Request) (*api.User, error) {
	if user, _, ok := r.BasicAuth(); ok {
		return mock.User("account:" + user), nil
	}
	return nil, auth.Errorf("missing credentials")
}

func (mockAuth) Close() error { return nil }
