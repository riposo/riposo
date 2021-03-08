package batch_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/riposo/riposo/internal/batch"
	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/schema"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler", func() {
	var subject http.Handler
	var w *httptest.ResponseRecorder

	BeforeEach(func() {
		subject = batch.Handler("/v1", mockMux)
		w = httptest.NewRecorder()
	})

	It("handles empty requests", func() {
		r := httptest.NewRequest(http.MethodPost, "/batch", strings.NewReader(`{"requests":[]}`))
		r.SetBasicAuth("alice", "")
		subject.ServeHTTP(w, r)

		Expect(w.Code).To(Equal(http.StatusOK))
		Expect(w.Header()).To(Equal(http.Header{
			"Content-Type": {"application/json; charset=utf-8"},
		}))
		Expect(w.Body.Bytes()).To(MatchJSON(`{"responses":[]}`))
	})

	It("handles bad requests", func() {
		r := httptest.NewRequest(http.MethodPost, "/batch", strings.NewReader(`[]`))
		r.SetBasicAuth("alice", "")
		subject.ServeHTTP(w, r)

		Expect(w.Code).To(Equal(http.StatusBadRequest))
		Expect(w.Body.Bytes()).To(MatchJSON(`{
			"code": 400,
			"errno": 107,
			"error": "Invalid parameters",
			"message": "body: Invalid JSON",
			"details": [
				{
					"location": "body",
					"description": "Invalid JSON"
				}
			]
		}`))
	})

	It("handles sub-requests", func() {
		r := httptest.NewRequest(http.MethodPost, "/batch", strings.NewReader(mockValidRequest))
		r.SetBasicAuth("alice", "")
		subject.ServeHTTP(w, r)

		Expect(w.Code).To(Equal(http.StatusOK))
		Expect(w.Header()).To(Equal(http.Header{
			"Content-Type": {"application/json; charset=utf-8"},
		}))
		Expect(w.Body.Bytes()).To(MatchJSON(`{
			"responses": [
				{
					"status": 200,
					"path": "/v1/books",
					"body": {
						"data": [
							{"id": 451, "title": "Fahrenheit 451", "author": "Ray Bradbury"}
						]
					},
					"headers": {
						"Content-Type": "application/json; charset=utf-8",
						"X-Custom": "batch"
					}
				},
				{
					"status": 200,
					"path": "/v1/books/451",
					"body": {
						"data": { "id": 451, "title": "Fahrenheit 451", "author": "Ray Bradbury" }
					},
					"headers": {
						"Content-Type": "application/json; charset=utf-8",
						"X-Custom": "batch"
					}
				},
				{
					"status": 304,
					"path": "/v1/books/451",
					"headers": {
						"X-Custom": "batch"
					}
				},
				{
					"status": 200,
					"path": "/v1/books",
					"body": {
						"data": {	"title": "Children of Men" }
					},
					"headers": {
						"Content-Type": "application/json; charset=utf-8",
						"X-Custom": "batch"
					}
				}
			]
		}`))
	})

	It("prevents recursion", func() {
		r := httptest.NewRequest(http.MethodPost, "/batch", strings.NewReader(`{
			"requests": [
				{ "method": "GET", "path": "/" },
				{ "method": "POST", "path": "/batch" }
			]
		}`))
		r.SetBasicAuth("alice", "")
		subject.ServeHTTP(w, r)

		Expect(w.Code).To(Equal(http.StatusBadRequest))
		Expect(w.Body.Bytes()).To(MatchJSON(`{
			"code": 400,
			"errno": 107,
			"error": "Invalid parameters",
			"message": "requests in body: Recursive call on /batch endpoint is forbidden.",
			"details": [
				{
					"location": "body",
					"name": "requests",
					"description": "Recursive call on /batch endpoint is forbidden."
				}
			]
		}`))
	})

	It("handles sub-request errors", func() {
		r := httptest.NewRequest(http.MethodPost, "/batch", strings.NewReader(`{
			"requests": [
				{ "method": "POST", "path": "/books/451" },
				{ "method": "GET", "path": "/missing" }
			]
		}`))
		r.SetBasicAuth("alice", "")
		subject.ServeHTTP(w, r)

		Expect(w.Code).To(Equal(http.StatusOK))
		Expect(w.Body.Bytes()).To(MatchJSON(`{
			"responses": [
				{
					"status": 405,
					"path": "/v1/books/451",
					"body": {
						"code": 405,
						"errno": 115,
						"error": "Method Not Allowed",
						"message": "Method not allowed on this endpoint."
					},
					"headers": {
						"Content-Type": "application/json; charset=utf-8",
						"X-Custom": "batch"
					}
				},
				{
					"status": 404,
					"path": "/v1/missing",
					"body": {
						"code": 404,
						"errno": 111,
						"error": "Not Found",
						"message": "The resource you are looking for could not be found."
					},
					"headers": {
						"Content-Type": "application/json; charset=utf-8",
						"X-Custom": "batch"
					}
				}
			]
		}`))
	})

	It("aborts on bad sub-requests", func() {
		r := httptest.NewRequest(http.MethodPost, "/batch", strings.NewReader(`{
			"requests": [
				{ "method": "GET", "path": "/" },
				{ "method": "BAD", "path": "/books/451" }
			]
		}`))
		r.SetBasicAuth("alice", "")
		subject.ServeHTTP(w, r)

		Expect(w.Code).To(Equal(http.StatusBadRequest))
		Expect(w.Body.Bytes()).To(MatchJSON(`{
			"code": 400,
			"errno": 107,
			"error": "Invalid parameters",
			"message": "requests.1 in body: invalid method \"BAD\"",
			"details": [{ "name": "requests.1", "location": "body", "description": "invalid method \"BAD\"" }]
		}`))

		r = httptest.NewRequest(http.MethodPost, "/batch", strings.NewReader(`{
			"requests": [
				{ "method": "GET", "path": "/" },
				{ "method": "GET", "path": "invalid\tpath" }
			]
		}`))
		r.SetBasicAuth("alice", "")
		w = httptest.NewRecorder()
		subject.ServeHTTP(w, r)

		Expect(w.Code).To(Equal(http.StatusBadRequest))
		Expect(w.Body.Bytes()).To(MatchJSON(`{
			"code": 400,
			"errno": 107,
			"error": "Invalid parameters",
			"message": "requests.1 in body: invalid path \"invalid\\tpath\"",
			"details": [{ "name": "requests.1", "location": "body", "description": "invalid path \"invalid\\tpath\"" }]
		}`))
	})

	It("delegates some headers", func() {
		r := httptest.NewRequest(http.MethodPost, "/batch", strings.NewReader(`{
			"requests": [
				{ "method": "GET", "path": "/books" }
			]
		}`))
		r.SetBasicAuth("bob", "")
		subject.ServeHTTP(w, r)

		Expect(w.Code).To(Equal(http.StatusOK))
		Expect(w.Body.Bytes()).To(MatchJSON(`{
			"responses": [
				{
					"status": 401,
					"path": "/v1/books",
					"headers": {
						"Www-Authenticate": "Basic realm=\"MOCK\"",
						"X-Custom": "batch"
					}
				}
			]
	}`))
	})
})

// --------------------------------------------------------------------

const mockValidRequest = `{
	"defaults": {
		"method": "GET",
		"path": "/books"
	},
	"requests": [
		{},
		{ "path": "/v1/books/451" },
		{
			"path": "/books/451",
			"headers": { "if-none-match": "451" }
		},
		{
			"method": "POST",
			"body": { "title": "Children of Men" }
		}
	]
}`

var mockMux = func() *chi.Mux {
	m := chi.NewMux()
	m.Use(middleware.SetHeader("X-Custom", "batch"))
	m.Use(middleware.BasicAuth("MOCK", map[string]string{"alice": ""}))

	m.NotFound(func(w http.ResponseWriter, r *http.Request) {
		api.Render(w, schema.NotFound)
	})
	m.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		api.Render(w, schema.MethodNotAllowed)
	})

	m.Get("/books", func(w http.ResponseWriter, _ *http.Request) {
		api.Render(w, json.RawMessage(`{
				"data": [
					{"id": 451, "title": "Fahrenheit 451", "author": "Ray Bradbury"}
				]
			}`))
	})

	m.Get("/books/451", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") == `451` {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		api.Render(w, json.RawMessage(`{
				"data": {
					"id": 451,
					"title": "Fahrenheit 451",
					"author": "Ray Bradbury"
				}
			}`))
	})

	m.Post("/books", func(w http.ResponseWriter, r *http.Request) {
		var data json.RawMessage
		if err := api.Parse(r, &data); err != nil {
			api.Render(w, err)
			return
		}
		api.Render(w, json.RawMessage(`{"data": `+string(data)+`}`))
	})

	return m
}()

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "internal/batch")
}
