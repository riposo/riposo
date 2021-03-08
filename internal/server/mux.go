package server

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	chimw "github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/riposo/riposo/internal/batch"
	"github.com/riposo/riposo/internal/config"
	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/auth"
	"github.com/riposo/riposo/pkg/conn"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"
)

type mux struct {
	*chi.Mux
	cns *conn.Set
	hlp *riposo.Helpers
	cfg *config.Config
}

func newMux(rts *api.Routes, cns *conn.Set, hlp *riposo.Helpers, cfg *config.Config, auth auth.Method) http.Handler {
	m := &mux{
		Mux: chi.NewMux(),
		cns: cns,
		hlp: hlp,
		cfg: cfg,
	}

	m.Use(chimw.RealIP)
	m.Use(chimw.RequestLogger(&logger{Logger: riposo.Logger}))
	m.Use(chimw.Recoverer)
	m.Use(chimw.Compress(3))
	m.Use(chimw.SetHeader("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none';"))
	m.Use(chimw.SetHeader("X-Content-Type-Options", "nosniff"))
	m.Use(cors.Handler(configCORS(m.cfg)))

	// custom error handlers
	m.NotFound(func(w http.ResponseWriter, r *http.Request) {
		api.Render(w, schema.NotFound)
	})
	m.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		api.Render(w, schema.MethodNotAllowed)
	})

	// redirect everything non-v1 to /v1
	m.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/v1"+r.URL.Path, http.StatusTemporaryRedirect)
	})

	// within /v1 scope
	m.Route("/v1", func(r chi.Router) {
		r.Get("/", m.Hello)
		r.Get("/__heartbeat__", m.Heartbeat)
		r.Get("/__lbheartbeat__", m.HeartbeatLB)

		r.Group(func(r chi.Router) {
			r.Use(chimw.StripSlashes)
			r.Use(transactional(m.cns, m.hlp, auth))

			r.Mount("/", rts.Mux())
			r.Method(http.MethodPost, "/batch", batch.Handler("/v1", rts.Mux()))
		})
	})

	return m
}

func (m *mux) Heartbeat(w http.ResponseWriter, r *http.Request) {
	api.Render(w, m.cns.Heartbeat(r.Context()))
}

func (*mux) HeartbeatLB(w http.ResponseWriter, _ *http.Request) {
	api.Render(w, struct{}{})
}

func (m *mux) Hello(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "/") {
		http.Redirect(w, r, "/v1/", http.StatusTemporaryRedirect)
		return
	}

	resp := schema.Hello{
		ProjectName:    m.cfg.Project.Name,
		ProjectDocs:    m.cfg.Project.Docs,
		ProjectVersion: m.cfg.Project.Version,
		HTTPAPIVersion: riposo.APIVersion,
		URL:            r.URL.String(),
		Capabilities:   m.cfg.Capabilities,
	}
	if !m.cfg.EOS.Time.IsZero() {
		resp.EOS = m.cfg.EOS.Time.Format("2006-01-02")
	}
	resp.Settings.BatchMaxRequests = m.cfg.Batch.MaxRequests
	api.Render(w, &resp)
}

// --------------------------------------------------------------------

func configCORS(c *config.Config) cors.Options {
	return cors.Options{
		AllowedOrigins: c.CORS.Origins,
		MaxAge:         int(c.CORS.MaxAge.Seconds()),
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
	}
}
