package server_test

import (
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/internal/server"
)

var _ = Describe("Muxer", func() {
	var subject http.Handler

	serve := func(r *http.Request) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		subject.ServeHTTP(w, r)
		return w
	}

	BeforeEach(func() {
		subject = NewMux()
	})

	It("redirects to /v1 scope", func() {
		w := serve(httptest.NewRequest(http.MethodGet, "/", nil))
		Expect(w.Code).To(Equal(http.StatusTemporaryRedirect))
		Expect(w.Header().Get("Location")).To(Equal("/v1/"))

		w = serve(httptest.NewRequest(http.MethodGet, "/unknown", nil))
		Expect(w.Code).To(Equal(http.StatusTemporaryRedirect))
		Expect(w.Header().Get("Location")).To(Equal("/v1/unknown"))
	})

	It("exposes default headers", func() {
		w := serve(httptest.NewRequest(http.MethodGet, "/v1/__lbheartbeat__", nil))
		Expect(w.Header()).To(Equal(http.Header{
			"Content-Type":            []string{"application/json; charset=utf-8"},
			"Content-Security-Policy": []string{"default-src 'none'; frame-ancestors 'none'; base-uri 'none';"},
			"X-Content-Type-Options":  []string{"nosniff"},
			"Vary":                    []string{"Origin"},
		}))
	})

	It("responds with NotFound if not found", func() {
		w := serve(httptest.NewRequest(http.MethodGet, "/v1/unknown", nil))
		Expect(w.Code).To(Equal(http.StatusNotFound))
		Expect(w.Body.String()).To(MatchJSON(`{
			"code": 404,
			"errno": 111,
			"error": "Not Found",
			"message": "The resource you are looking for could not be found."
		}`))
	})

	Describe("GET /v1/", func() {
		It("responds", func() {
			w := serve(httptest.NewRequest(http.MethodGet, "/v1/", nil))
			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Body.Bytes()).To(MatchJSON(`{
				"project_name": "Example Project",
				"project_docs": "http://example.com",
				"project_version": "0.11.2",
				"http_api_version": "1.22",
				"url": "/v1/",
				"eos": "2046-10-27",
				"settings": {
					"batch_max_requests": 25,
					"readonly": false
				},
				"capabilities": {}
			}`))
		})

		It("redirects without a trailing slash", func() {
			w := serve(httptest.NewRequest(http.MethodGet, "/v1", nil))
			Expect(w.Code).To(Equal(http.StatusTemporaryRedirect))
			Expect(w.Header().Get("Location")).To(Equal(`/v1/`))
		})
	})

	Describe("GET /v1/__heartbeat__", func() {
		It("responds", func() {
			w := serve(httptest.NewRequest(http.MethodGet, "/v1/__heartbeat__", nil))
			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Body.Bytes()).To(MatchJSON(`{
				"storage":true,
				"permission":true,
				"cache":true
			}`))
		})
	})

	Describe("GET /v1/__lbheartbeat__", func() {
		It("responds", func() {
			w := serve(httptest.NewRequest(http.MethodGet, "/v1/__lbheartbeat__", nil))
			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Body.Bytes()).To(MatchJSON(`{}`))
		})
	})

	Describe("GET /v1/buckets", func() {
		It("responds", func() {
			r := httptest.NewRequest(http.MethodGet, "/v1/buckets", nil)
			r.SetBasicAuth("alice", "")

			w := serve(r)
			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Body.Bytes()).To(MatchJSON(`{"data": []}`))
		})
	})

	Describe("POST /v1/batch", func() {
		It("responds", func() {
			r := httptest.NewRequest(http.MethodPost, "/v1/batch", strings.NewReader(`{
				"requests": [
					{ "method": "GET", "path": "/buckets" },
					{ "method": "GET", "path": "/buckets/foo" }
				]
			}`))
			r.SetBasicAuth("alice", "")
			r.Header.Set("Origin", "http://test.host")

			w := serve(r)
			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Body.Bytes()).To(MatchJSON(`{
				"responses": [
					{
						"status": 200,
						"path": "/v1/buckets",
						"body": {
							"data": []
						},
						"headers": {
							"Cache-Control": "no-cache, no-store, no-transform, must-revalidate, private, max-age=0",
							"Content-Type": "application/json; charset=utf-8",
							"Etag": "\"0\"",
							"Last-Modified": "Thu, 01 Jan 1970 00:00:00 GMT"
						}
					},
					{
						"status": 403,
						"path": "/v1/buckets/foo",
						"body": {
							"code": 403,
							"errno": 121,
							"error": "Forbidden",
							"message": "This user cannot access this resource."
						},
						"headers": {
							"Content-Type": "application/json; charset=utf-8"
						}
					}
				]
			}`))
		})
	})
})
