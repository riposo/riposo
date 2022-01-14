package server

import (
	"context"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/riposo/riposo/internal/config"
	"github.com/riposo/riposo/internal/model/group"
	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/auth"
	"github.com/riposo/riposo/pkg/conn"
	"github.com/riposo/riposo/pkg/plugin"
	"github.com/riposo/riposo/pkg/riposo"
	"go.uber.org/multierr"
)

// Server implements a HTTP server.
type Server struct {
	srv *http.Server
	cfg *config.Config
	cls []io.Closer
}

// New inits the server and binds it to an address.
func New(ctx context.Context, cfg *config.Config) (*Server, error) {
	// init helpers
	hlp, err := cfg.InitHelpers()
	if err != nil {
		return nil, err
	}

	// init routes, install resources
	rts := api.NewRoutes(cfg.APIConfig())
	rts.Resource("/buckets", nil)
	rts.Resource("/buckets/{bucket_id}/groups", group.Model{})
	rts.Resource("/buckets/{bucket_id}/collections", nil)
	rts.Resource("/buckets/{bucket_id}/collections/{collection_id}/records", nil)

	// init plugins
	plugins, err := plugin.Init(rts, hlp, cfg.Plugins)
	if err != nil {
		return nil, err
	}
	cfg.Capabilities = plugins

	// init auth
	auth, err := initAuth(hlp, cfg)
	if err != nil {
		_ = plugins.Close()
		return nil, err
	}

	cns, err := establishConns(ctx, hlp, cfg)
	if err != nil {
		_ = auth.Close()
		_ = plugins.Close()
		return nil, err
	}

	mux := newMux(rts, hlp, cns, auth, cfg)
	cls := []io.Closer{cns, auth, plugins}

	return &Server{
		srv: &http.Server{
			Handler:           mux,
			Addr:              cfg.Server.Address,
			ReadHeaderTimeout: time.Second,
			ReadTimeout:       cfg.Server.ReadTimeout,
			WriteTimeout:      cfg.Server.WriteTimeout,
			BaseContext:       func(_ net.Listener) context.Context { return ctx },
		},
		cfg: cfg,
		cls: cls,
	}, nil
}

// ListenAndServe starts the server.
func (s *Server) ListenAndServe() error {
	riposo.Logger.Println("starting server on", s.srv.Addr)
	return s.srv.ListenAndServe()
}

// Close stops the server and releases all resources.
func (s *Server) Close() error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(s.cfg.Server.ShutdownTimeout))
	defer cancel()

	err := s.srv.Shutdown(ctx)
	for _, c := range s.cls {
		err = multierr.Append(err, c.Close())
	}
	return s.srv.Close()
}

// --------------------------------------------------------------------

func initAuth(hlp riposo.Helpers, cfg *config.Config) (auth.Method, error) {
	sub := make([]auth.Method, 0, len(cfg.Auth.Methods))
	for _, m := range cfg.Auth.Methods {
		factory, err := auth.Get(m)
		if err != nil {
			return nil, err
		}

		method, err := factory(hlp)
		if err != nil {
			return nil, err
		}

		sub = append(sub, method)
	}
	return auth.OneOf(sub...), nil
}

func establishConns(ctx context.Context, hlp riposo.Helpers, cfg *config.Config) (*conn.Set, error) {
	return conn.Connect(
		ctx,
		cfg.Storage.URL,
		cfg.Permission.URL,
		cfg.Cache.URL,
		hlp,
	)
}
