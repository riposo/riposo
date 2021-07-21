package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type muxKey struct{}

// GetMux extracts the API mux from the request.
func GetMux(req *http.Request) http.Handler {
	if v := req.Context().Value(muxKey{}); v != nil {
		return v.(http.Handler)
	}
	return nil
}

// Middleware represents a middleware handler function.
type Middleware func(http.Handler) http.Handler

// Routes contains the main API route defintions.
type Routes struct {
	mux   *chi.Mux
	hooks *Hooks
	cfg   *Config
}

// NewRoutes inits a new routes instance.
func NewRoutes(cfg *Config) *Routes {
	mux := chi.NewMux()
	mux.Use(middleware.WithValue(muxKey{}, mux))

	return &Routes{
		mux:   mux,
		hooks: new(Hooks),
		cfg:   cfg.norm(),
	}
}

// Use registers a new middleware.
func (r *Routes) Use(middleware Middleware) {
	r.mux.Use(middleware)
}

// Method registers a new HTTP handler under a particular HTTP method.
func (r *Routes) Method(method, pattern string, handler http.Handler) {
	r.mux.Method(method, pattern, handler)
}

// Handle registers a new API handler.
func (r *Routes) Handle(pattern string, handler http.Handler) {
	r.mux.Handle(pattern, handler)
}

// Hook registers resource callbacks.
func (r *Routes) Hook(globs []string, callbacks Hook) {
	r.hooks.Register(globs, callbacks)
}

// Resource registers a new resource under a prefix.
func (r *Routes) Resource(prefix string, model Model) {
	if model == nil {
		model = DefaultModel{}
	}

	c := &controller{
		mod:   model,
		hooks: r.hooks,
		cfg:   r.cfg,
	}

	r.mux.Route(prefix, func(ns chi.Router) {
		ns.Method(http.MethodGet, "/", HandlerFunc(c.List))
		ns.Method(http.MethodHead, "/", HandlerFunc(c.Count))
		ns.Method(http.MethodDelete, "/", HandlerFunc(c.DeleteBulk))

		ns.Method(http.MethodGet, "/{id}", HandlerFunc(c.Get))
		ns.Method(http.MethodPost, "/", HandlerFunc(c.Create))
		ns.Method(http.MethodPut, "/{id}", HandlerFunc(c.Update))
		ns.Method(http.MethodPatch, "/{id}", HandlerFunc(c.Patch))
		ns.Method(http.MethodDelete, "/{id}", HandlerFunc(c.Delete))
	})
}

// Mux returns the mountable API mux.
func (r *Routes) Mux() http.Handler {
	return r.mux
}
