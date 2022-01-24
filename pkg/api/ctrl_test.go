package api_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/bsm/gomega/types"
	"github.com/riposo/riposo/pkg/conn/permission"
	"github.com/riposo/riposo/pkg/mock"
	"github.com/riposo/riposo/pkg/riposo"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/pkg/api"
)

var _ = Describe("Routes.Resource", func() {
	var subject *Routes
	var txn *Txn

	var (
		alice  = mock.User("account:alice", "principal:team")
		bob    = mock.User("account:bob")
		claire = mock.User("account:claire", "principal:team")
	)

	newRequest := func(method, path, payload string) *http.Request {
		var body io.Reader
		if payload != "" {
			body = strings.NewReader(payload)
		}
		return mock.Request(txn, method, path, body)
	}

	serve := func(r *http.Request) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		subject.Mux().ServeHTTP(w, r)
		return w
	}

	handle := func(method, path, payload string) *httptest.ResponseRecorder {
		return serve(newRequest(method, path, payload))
	}

	seedThree := func() {
		ExpectWithOffset(1, handle(http.MethodPost, "/resources", `{"data": {"id":"alpha"}}`).Code).To(Equal(http.StatusCreated))
		ExpectWithOffset(1, handle(http.MethodPost, "/resources", `{"data": {"id":"beta"}}`).Code).To(Equal(http.StatusCreated))
		ExpectWithOffset(1, handle(http.MethodPost, "/resources", `{"data": {"id":"gamma"}}`).Code).To(Equal(http.StatusCreated))
	}

	BeforeEach(func() {
		txn = mock.Txn()
		txn.User = alice

		// init config
		cfg := &Config{
			// grant resource:create to team members
			Authz: Authz{"resource:create": {"principal:team"}},
		}
		cfg.Pagination.MaxLimit = 3

		// setup routes and compile
		subject = NewRoutes(cfg)
		subject.Resource("/resources", nil)
		subject.Resource("/resources/{resourceID}/nested", nil)
	})

	AfterEach(func() {
		Expect(txn.Abort()).To(Succeed())
	})

	Describe("GET /resources", func() {
		const nextPageToken = "eyJsYXN0X29iamVjdCI6eyJsYXN0X21vZGlmaWVkIjoxNTE1MTUxNTE1Njc4fX0"
		const nextPageURL = "/resources?_limit=2&_sort=last_modified&_token=" + nextPageToken

		BeforeEach(seedThree)

		It("responds", func() {
			Expect(handle(http.MethodGet, "/resources?_sort=last_modified", "")).To(MatchResponse(http.StatusOK, map[string]string{
				"Cache-Control": "no-cache, no-store, no-transform, must-revalidate, private, max-age=0",
				"Content-Type":  "application/json; charset=utf-8",
				"Etag":          `"1515151515679"`,
				"Last-Modified": "Fri, 05 Jan 2018 11:25:15 GMT",
			}, `{
				"data": [
					{"id": "alpha", "last_modified": 1515151515677},
					{"id": "beta", "last_modified": 1515151515678},
					{"id": "gamma", "last_modified": 1515151515679}
				]
			}`))
		})

		It("paginates", func() {
			Expect(handle(http.MethodGet, "/resources?_sort=last_modified&_limit=2", "")).To(MatchResponse(http.StatusOK, map[string]string{
				"Next-Page": nextPageURL,
			}, `{
				"data": [
					{"id": "alpha", "last_modified": 1515151515677},
					{"id": "beta", "last_modified": 1515151515678}
				]
			}`))

			Expect(handle(http.MethodGet, nextPageURL, "")).To(MatchResponse(http.StatusOK, nil, `{
				"data": [
					{"id": "gamma", "last_modified": 1515151515679}
				]
			}`))
		})

		It("supports conditional rendering", func() {
			// If-None-Match
			r := newRequest(http.MethodGet, "/resources", ``)
			r.Header.Set("If-None-Match", `"1515151515679"`)
			w := serve(r)
			Expect(w.Code).To(Equal(http.StatusNotModified))
			Expect(w.Body.Len()).To(BeZero())

			// If-Match
			r = newRequest(http.MethodGet, "/resources", ``)
			r.Header.Set("If-Match", `"1616161616000"`)
			Expect(serve(r)).To(MatchResponse(http.StatusPreconditionFailed, nil, `{
				"code": 412,
				"errno": 114,
				"error": "Precondition Failed",
				"message": "Resource was modified meanwhile"
			}`))
		})
	})

	Describe("HEAD /resources", func() {
		BeforeEach(seedThree)

		It("authorizes", func() {
			// bob has no access to resources
			txn.User = bob
			Expect(handle(http.MethodHead, "/resources", ``)).To(MatchResponse(http.StatusOK, map[string]string{
				"Total-Objects": "0",
				"Total-Records": "0",
			}, ``))
		})

		It("responds", func() {
			Expect(handle(http.MethodHead, "/resources", ``)).To(MatchResponse(http.StatusOK, map[string]string{
				"Cache-Control": "no-cache, no-store, no-transform, must-revalidate, private, max-age=0",
				"Etag":          `"1515151515679"`,
				"Last-Modified": "Fri, 05 Jan 2018 11:25:15 GMT",
				"Total-Objects": "3",
				"Total-Records": "3",
			}, ""))
		})
	})

	Describe("DELETE /resources", func() {
		const nextPageToken = "eyJub25jZSI6InBhZ2luYXRpb24tdG9rZW4tRVBSLklEIiwibGFzdF9vYmplY3QiOnsibGFzdF9tb2RpZmllZCI6MTUxNTE1MTUxNTY3OH19"
		const nextPageURL = "/resources?_limit=2&_sort=last_modified&_token=" + nextPageToken

		BeforeEach(seedThree)

		It("authorizes", func() {
			// bob has no access to resources
			txn.User = bob
			Expect(handle(http.MethodDelete, "/resources", ``)).To(MatchResponse(http.StatusOK, nil, `{"data": []}`))
		})

		It("responds", func() {
			Expect(handle(http.MethodGet, "/resources/alpha", ``).Code).To(Equal(http.StatusOK))
			Expect(handle(http.MethodDelete, "/resources?_sort=last_modified", ``)).To(MatchResponse(http.StatusOK, map[string]string{
				"Content-Type":  "application/json; charset=utf-8",
				"Etag":          `"1515151515682"`,
				"Last-Modified": "Fri, 05 Jan 2018 11:25:15 GMT",
			}, `{
				"data": [
					{"id": "alpha", "last_modified": 1515151515682, "deleted": true},
					{"id": "beta", "last_modified": 1515151515682, "deleted": true},
					{"id": "gamma", "last_modified": 1515151515682, "deleted": true}
				]
			}`))
			Expect(handle(http.MethodGet, "/resources/alpha", ``).Code).To(Equal(http.StatusForbidden))
		})

		It("paginates", func() {
			Expect(handle(http.MethodDelete, "/resources?_sort=last_modified&_limit=2", ``)).To(MatchResponse(http.StatusOK, map[string]string{
				"Content-Type":  "application/json; charset=utf-8",
				"Etag":          `"1515151515681"`,
				"Last-Modified": "Fri, 05 Jan 2018 11:25:15 GMT",
				"Next-Page":     nextPageURL,
			}, `{
				"data": [
					{"id": "alpha", "last_modified": 1515151515681, "deleted": true},
					{"id": "beta", "last_modified": 1515151515681, "deleted": true}
				]
			}`))

			Expect(handle(http.MethodDelete, nextPageURL, ``)).To(MatchResponse(http.StatusOK, nil, `{
				"data": [
					{"id": "gamma", "last_modified": 1515151515682, "deleted": true}
				]
			}`))

			Expect(handle(http.MethodDelete, nextPageURL, ``)).To(MatchResponse(http.StatusBadRequest, nil, `{
				"code": 400,
				"errno": 107,
				"error": "Invalid parameters",
				"message": "querystring: _token was already used or has expired",
				"details": [
					{"location":"querystring", "description":"_token was already used or has expired"}
				]
			}`))
		})

		It("supports conditional rendering", func() {
			r := newRequest(http.MethodDelete, "/resources", ``)
			r.Header.Set("If-Match", `"1616161616000"`)
			Expect(serve(r)).To(MatchResponse(http.StatusPreconditionFailed, nil, `{
				"code": 412,
				"errno": 114,
				"error": "Precondition Failed",
				"message": "Resource was modified meanwhile"
			}`))
		})

		It("only deletes writable resources", func() {
			// seed delta, writable by bob
			ExpectWithOffset(1, handle(http.MethodPost, "/resources", `{
				"data": {"id":"delta"},
				"permissions": {"write":["account:bob"]}
			}`).Code).To(Equal(http.StatusCreated))

			// as bob
			txn.User = bob
			Expect(handle(http.MethodDelete, "/resources?_sort=last_modified", ``)).To(MatchResponse(http.StatusOK, map[string]string{
				"Content-Type":  "application/json; charset=utf-8",
				"Etag":          `"1515151515681"`,
				"Last-Modified": "Fri, 05 Jan 2018 11:25:15 GMT",
			}, `{
				"data": [
					{"id": "delta", "last_modified": 1515151515681, "deleted": true}
				]
			}`))
		})
	})

	Describe("GET /resource/ID", func() {
		BeforeEach(func() {
			// seed resource: alpha
			Expect(handle(http.MethodPost, "/resources", `{"data": {"id":"alpha", "meta":"data"}}`).Code).To(Equal(http.StatusCreated))
		})

		It("requires authentication", func() {
			txn.User = mock.User(riposo.Everyone)
			Expect(handle(http.MethodGet, "/resources/alpha", ``)).To(MatchResponse(http.StatusUnauthorized, nil, `{
				"code": 401,
				"errno": 104,
				"error": "Unauthorized",
				"message": "Please authenticate yourself to use this endpoint."
			}`))
		})

		It("validates object IDs", func() {
			Expect(handle(http.MethodGet, "/resources/invalid%20id", ``)).To(MatchResponse(http.StatusBadRequest, nil, `{
				"code": 400,
				"errno": 107,
				"error": "Invalid parameters",
				"message": "path: Invalid object id",
				"details": [
					{"location": "path", "description": "Invalid object id"}
				]
			}`))
		})

		It("authorizes", func() {
			// bob has no access to alpha
			txn.User = bob
			Expect(handle(http.MethodGet, "/resources/alpha", ``)).To(MatchResponse(http.StatusForbidden, nil, `{
				"code": 403,
				"errno": 121,
				"error": "Forbidden",
				"message": "This user cannot access this resource."
			}`))

			// even with conditions
			r := newRequest(http.MethodGet, "/resources/alpha", ``)
			r.Header.Set("If-Match", `"1616161616000"`)
			Expect(serve(r).Code).To(Equal(http.StatusForbidden))

			// alice has no access to beta (does not exist)
			txn.User = alice
			Expect(handle(http.MethodGet, "/resources/beta", ``)).To(MatchResponse(http.StatusForbidden, nil, `{
				"code": 403,
				"errno": 121,
				"error": "Forbidden",
				"message": "This user cannot access this resource."
			}`))
		})

		It("responds", func() {
			Expect(handle(http.MethodGet, "/resources/alpha", ``)).To(MatchResponse(http.StatusOK, map[string]string{
				"Cache-Control": "no-cache, no-store, no-transform, must-revalidate, private, max-age=0",
				"Content-Type":  "application/json; charset=utf-8",
				"Etag":          `"1515151515677"`,
				"Last-Modified": "Fri, 05 Jan 2018 11:25:15 GMT",
			}, `{
				"data": {
					"id": "alpha",
					"last_modified": 1515151515677,
					"meta": "data"
				},
				"permissions": {
					"write": ["account:alice"]
				}
			}`))
		})

		It("supports conditional rendering", func() {
			// If-None-Match
			r := newRequest(http.MethodGet, "/resources/alpha", ``)
			r.Header.Set("If-None-Match", `"1515151515677"`)
			w := serve(r)
			Expect(w.Code).To(Equal(http.StatusNotModified))
			Expect(w.Body.Len()).To(BeZero())

			// If-Match
			r = newRequest(http.MethodGet, "/resources/alpha", ``)
			r.Header.Set("If-Match", `"1616161616000"`)
			Expect(serve(r)).To(MatchResponse(http.StatusPreconditionFailed, nil, `{
				"code": 412,
				"errno": 114,
				"error": "Precondition Failed",
				"message": "Resource was modified meanwhile",
				"details": {
					"existing": {
						"id": "alpha",
						"last_modified": 1515151515677,
						"meta": "data"
					}
				}
			}`))
		})
	})

	Describe("POST /resources", func() {
		It("authorizes", func() {
			// bob cannot create resources
			txn.User = bob
			Expect(handle(http.MethodPost, "/resources", `{"data": {}}`)).To(MatchResponse(http.StatusForbidden, nil, `{
				"code": 403,
				"errno": 121,
				"error": "Forbidden",
				"message": "This user cannot access this resource."
			}`))

			// grant bob global write permissions
			Expect(txn.Perms.AddACEPrincipal("account:bob", permission.ACE{Perm: "write"})).To(Succeed())
			// seed resource: beta
			Expect(handle(http.MethodPost, "/resources", `{"data": {"id":"beta"}}`).Code).To(Equal(http.StatusCreated))

			// alice cannot read beta
			txn.User = alice
			Expect(handle(http.MethodPost, "/resources", `{"data": {"id": "beta"}}`)).To(MatchResponse(http.StatusForbidden, nil, `{
				"code": 403,
				"errno": 121,
				"error": "Forbidden",
				"message": "This user cannot access this resource."
			}`))
		})

		It("responds", func() {
			Expect(handle(http.MethodPost, "/resources", `{
				"permissions": {
					"sub:create": ["account:alice"],
					"write": ["system.Everyone"]
				}
			}`)).To(MatchResponse(http.StatusCreated, map[string]string{
				"Content-Type":  "application/json; charset=utf-8",
				"Etag":          `"1515151515677"`,
				"Last-Modified": "Fri, 05 Jan 2018 11:25:15 GMT",
			}, `{
				"data": {
					"id": "EPR.ID",
					"last_modified": 1515151515677
				},
				"permissions": {
					"sub:create": ["account:alice"],
					"write": ["account:alice",	"system.Everyone"]
				}
			}`))

			Expect(handle(http.MethodGet, "/resources/EPR.ID", ``)).To(MatchResponse(http.StatusOK, nil, `{
				"data": {
					"id": "EPR.ID",
					"last_modified": 1515151515677
				},
				"permissions": {
					"sub:create": ["account:alice"],
					"write": ["account:alice", "system.Everyone"]
				}
			}`))
		})

		It("accepts custom IDs", func() {
			Expect(handle(http.MethodPost, "/resources", `{
				"data": {
					"id": "alpha",
					"meta": "too"
				}
			}`)).To(MatchResponse(http.StatusCreated, nil, `{
				"data": {
					"id": "alpha",
					"meta": "too",
					"last_modified": 1515151515677
				},
				"permissions": {
					"write": ["account:alice"]
				}
			}`))

			Expect(handle(http.MethodGet, "/resources/alpha", ``)).To(MatchResponse(http.StatusOK, nil, `{
				"data": {
					"id": "alpha",
					"last_modified": 1515151515677,
					"meta": "too"
				},
				"permissions": {
					"write": ["account:alice"]
				}
			}`))
		})

		It("validates IDs", func() {
			Expect(handle(http.MethodPost, "/resources", `{
				"data": {"id": "invalid id"}
			}`)).To(MatchResponse(http.StatusBadRequest, nil, `{
				"code": 400,
				"errno": 107,
				"error": "Invalid parameters",
				"message": "path: Invalid object id",
				"details": [
					{"location": "path", "description": "Invalid object id"}
				]
			}`))
		})

		It("returns existing unmodified", func() {
			// seed resource: alpha
			Expect(handle(http.MethodPost, "/resources", `{"data": {"id":"alpha"}}`).Code).To(Equal(http.StatusCreated))
			Expect(handle(http.MethodPost, "/resources", `{
				"data": {
					"id": "alpha",
					"ignored": "update"
				},
				"permissions": {
					"write": ["system.Everyone"]
				}
			}`)).To(MatchResponse(http.StatusOK, map[string]string{
				"Content-Type":  "application/json; charset=utf-8",
				"Etag":          `"1515151515677"`,
				"Last-Modified": "Fri, 05 Jan 2018 11:25:15 GMT",
			}, `{
				"data": {
					"id": "alpha",
					"last_modified": 1515151515677
				},
				"permissions": {
					"write": ["account:alice"]
				}
			}`))
		})

		It("does not fail on blank body", func() {
			Expect(handle(http.MethodPost, "/resources", ``)).To(MatchResponse(http.StatusCreated, nil, `{
				"data": {
					"id": "EPR.ID",
					"last_modified": 1515151515677
				},
				"permissions": {
					"write": ["account:alice"]
				}
			}`))
		})
	})

	Describe("PUT /resources/{id}", func() {
		BeforeEach(func() {
			// seed resource: alpha
			Expect(handle(http.MethodPost, "/resources", `{
				"data": {"id":"alpha", "meta":"data"},
				"permissions": {"read":["account:bob"]}
			}`).Code).To(Equal(http.StatusCreated))
		})

		It("authorizes", func() {
			// bob cannot write alpha
			txn.User = bob
			Expect(handle(http.MethodPut, "/resources/alpha", `{"data":{}}`)).To(MatchResponse(http.StatusForbidden, nil, `{
				"code": 403,
				"errno": 121,
				"error": "Forbidden",
				"message": "This user cannot access this resource."
			}`))

			// even with conditions
			r := newRequest(http.MethodPut, "/resources/alpha", `{"data": {"extra": "value"}}`)
			r.Header.Set("If-Match", `"1616161616000"`)
			Expect(serve(r).Code).To(Equal(http.StatusForbidden))

			// alice can create beta via PUT
			txn.User = alice
			Expect(handle(http.MethodPut, "/resources/beta", `{"data":{}}`).Code).To(Equal(http.StatusCreated))
		})

		It("responds", func() {
			Expect(handle(http.MethodPut, "/resources/alpha", `{
				"data": {"extra": "value"},
				"permissions": {
					"write": ["system.Authenticated"],
					"read": null
				}
			}`)).To(MatchResponse(http.StatusOK, map[string]string{
				"Content-Type":  "application/json; charset=utf-8",
				"Etag":          `"1515151515678"`,
				"Last-Modified": "Fri, 05 Jan 2018 11:25:15 GMT",
			}, `{
				"data": {
					"id": "alpha",
					"extra": "value",
					"last_modified": 1515151515678
				},
				"permissions": {
					"write": ["account:alice", "system.Authenticated"]
				}
			}`))
		})

		It("allows to leave permissions unchanged", func() {
			Expect(handle(http.MethodPut, "/resources/alpha", `{
				"data": {"extra": "value"}
			}`)).To(MatchResponse(http.StatusOK, nil, `{
				"data": {
					"id": "alpha",
					"extra": "value",
					"last_modified": 1515151515678
				},
				"permissions": {
					"write": ["account:alice"],
					"read": ["account:bob"]
				}
			}`))
		})

		It("creates if not found", func() {
			Expect(handle(http.MethodPut, "/resources/beta", `{
				"data": {
					"id": "beta",
					"meta": "data"
				},
				"permissions": {
					"write": ["principal:team"]
				}
			}`)).To(MatchResponse(http.StatusCreated, nil, `{
				"data": {
					"id": "beta",
					"meta": "data",
					"last_modified": 1515151515678
				},
				"permissions": {
					"write": ["account:alice", "principal:team"]
				}
			}`))

			Expect(handle(http.MethodGet, "/resources/beta", ``)).To(MatchResponse(http.StatusOK, nil, `{
				"data": {
					"id": "beta",
					"last_modified": 1515151515678,
					"meta": "data"
				},
				"permissions": {
					"write": ["account:alice", "principal:team"]
				}
			}`))
		})

		It("creates if not found (empty body)", func() {
			Expect(handle(http.MethodPut, "/resources/beta", `{}`)).
				To(MatchResponse(http.StatusCreated, nil, `{
				"data": {
					"id": "beta",
					"last_modified": 1515151515678
				},
				"permissions": {
					"write": ["account:alice"]
				}
			}`))
		})

		It("creates if not found (only data)", func() {
			Expect(handle(http.MethodPut, "/resources/beta", `{
				"data": { "meta": "data" }
			}`)).To(MatchResponse(http.StatusCreated, nil, `{
				"data": {
					"id": "beta",
					"meta": "data",
					"last_modified": 1515151515678
				},
				"permissions": {
					"write": ["account:alice"]
				}
			}`))
		})

		It("rejects inconsistent IDs", func() {
			Expect(handle(http.MethodPut, "/resources/alpha", `{
				"data": {"id": "beta"}
			}`)).To(MatchResponse(http.StatusBadRequest, map[string]string{
				"Content-Type": "application/json; charset=utf-8",
			}, `{
				"code": 400,
				"errno": 107,
				"error": "Invalid parameters",
				"message": "data.id in body: Does not match requested object",
				"details": [
					{"name": "data.id", "location": "body", "description": "Does not match requested object"}
				]
			}`))
		})

		It("validates object IDs", func() {
			Expect(handle(http.MethodPut, "/resources/invalid%20id", ``)).To(MatchResponse(http.StatusBadRequest, nil, `{
				"code": 400,
				"errno": 107,
				"error": "Invalid parameters",
				"message": "path: Invalid object id",
				"details": [
					{"location": "path", "description": "Invalid object id"}
				]
			}`))
		})

		It("supports conditional rendering", func() {
			r := newRequest(http.MethodPut, "/resources/alpha", `{"data": {"extra": "value"}}`)
			r.Header.Set("If-Match", `"1616161616000"`)
			Expect(serve(r)).To(MatchResponse(http.StatusPreconditionFailed, nil, `{
				"code": 412,
				"errno": 114,
				"error": "Precondition Failed",
				"message": "Resource was modified meanwhile",
				"details": {
					"existing": {
						"id": "alpha",
						"last_modified": 1515151515677,
						"meta": "data"
					}
				}
			}`))
		})
	})

	Describe("PATCH /resources/{id}", func() {
		BeforeEach(func() {
			// seed resource: alpha
			Expect(handle(http.MethodPost, "/resources", `{
				"data": {"id":"alpha", "meta":"data"},
				"permissions": {"read":["account:bob"]}
			}`).Code).To(Equal(http.StatusCreated))
		})

		It("authorizes", func() {
			// bob cannot write alpha
			txn.User = bob
			Expect(handle(http.MethodPatch, "/resources/alpha", `{"data":{}}`)).To(MatchResponse(http.StatusForbidden, nil, `{
				"code": 403,
				"errno": 121,
				"error": "Forbidden",
				"message": "This user cannot access this resource."
			}`))

			// even with conditions
			r := newRequest(http.MethodPatch, "/resources/alpha", `{"data": {}}`)
			r.Header.Set("If-Match", `"1616161616000"`)
			Expect(serve(r).Code).To(Equal(http.StatusForbidden))
		})

		It("responds", func() {
			Expect(handle(http.MethodPatch, "/resources/alpha", `{
				"data": {"extra": "value"},
				"permissions": {
					"write": ["system.Everyone"],
					"read": null
				}
			}`)).To(MatchResponse(http.StatusOK, map[string]string{
				"Content-Type":  "application/json; charset=utf-8",
				"Etag":          `"1515151515678"`,
				"Last-Modified": "Fri, 05 Jan 2018 11:25:15 GMT",
			}, `{
				"data": {
					"id": "alpha",
					"last_modified": 1515151515678,
					"meta": "data",
					"extra": "value"
				},
				"permissions": {
					"write": ["account:alice", "system.Everyone"]
				}
			}`))
		})

		It("rejects inconsistent IDs", func() {
			Expect(handle(http.MethodPatch, "/resources/alpha", `{
				"data": {"id": "beta"}
			}`)).To(MatchResponse(http.StatusBadRequest, map[string]string{
				"Content-Type": "application/json; charset=utf-8",
			}, `{
				"code": 400,
				"errno": 107,
				"error": "Invalid parameters",
				"message": "data.id in body: Does not match requested object",
				"details": [
					{"name": "data.id", "location": "body", "description": "Does not match requested object"}
				]
			}`))
		})

		It("validates object IDs", func() {
			Expect(handle(http.MethodPatch, "/resources/invalid%20id", ``)).To(MatchResponse(http.StatusBadRequest, nil, `{
				"code": 400,
				"errno": 107,
				"error": "Invalid parameters",
				"message": "path: Invalid object id",
				"details": [
					{"location": "path", "description": "Invalid object id"}
				]
			}`))
		})

		It("requires data or permissions", func() {
			Expect(handle(http.MethodPatch, "/resources/alpha", `{}`)).To(MatchResponse(http.StatusBadRequest, nil, `{
				"code": 400,
				"errno": 107,
				"error": "Invalid parameters",
				"message": "body: Provide at least one of data or permissions",
				"details": [
					{"location": "body", "description": "Provide at least one of data or permissions"}
				]
			}`))
		})

		It("allows to leave permissions unchanged", func() {
			Expect(handle(http.MethodPatch, "/resources/alpha", `{
				"data": {"extra": "value"}
			}`)).To(MatchResponse(http.StatusOK, nil, `{
				"data": {
					"id": "alpha",
					"last_modified": 1515151515678,
					"meta": "data",
					"extra": "value"
				},
				"permissions": {
					"write": ["account:alice"],
					"read": ["account:bob"]
				}
			}`))
		})

		It("allows to only update permissions", func() {
			Expect(handle(http.MethodPatch, "/resources/alpha", `{
				"permissions": {"write": ["system.Everyone"]}
			}`)).To(MatchResponse(http.StatusOK, nil, `{
				"data": {
					"id": "alpha",
					"last_modified": 1515151515678,
					"meta": "data"
				},
				"permissions": {
					"write": ["account:alice", "system.Everyone"],
					"read": ["account:bob"]
				}
			}`))
		})

		It("supports conditional rendering", func() {
			r := newRequest(http.MethodPatch, "/resources/alpha", `{"data": {"extra": "value"}}`)
			r.Header.Set("If-Match", `"1616161616000"`)
			Expect(serve(r)).To(MatchResponse(http.StatusPreconditionFailed, nil, `{
				"code": 412,
				"errno": 114,
				"error": "Precondition Failed",
				"message": "Resource was modified meanwhile",
				"details": {
					"existing": {
						"id": "alpha",
						"last_modified": 1515151515677,
						"meta": "data"
					}
				}
			}`))
		})
	})

	Describe("DELETE /resources/{id}", func() {
		BeforeEach(func() {
			// seed resource: alpha
			Expect(handle(http.MethodPost, "/resources", `{"data": {"id":"alpha"}}`).Code).To(Equal(http.StatusCreated))
		})

		It("authorizes", func() {
			// bob cannot write alpha
			txn.User = bob
			Expect(handle(http.MethodDelete, "/resources/alpha", `{"data":{}}`)).To(MatchResponse(http.StatusForbidden, nil, `{
				"code": 403,
				"errno": 121,
				"error": "Forbidden",
				"message": "This user cannot access this resource."
			}`))

			// even with conditions
			r := newRequest(http.MethodDelete, "/resources/alpha", `{"data": {}}`)
			r.Header.Set("If-Match", `"1616161616000"`)
			Expect(serve(r).Code).To(Equal(http.StatusForbidden))
		})

		It("responds", func() {
			Expect(handle(http.MethodGet, "/resources/alpha", ``).Code).To(Equal(http.StatusOK))
			Expect(handle(http.MethodDelete, "/resources/alpha", ``)).To(MatchResponse(http.StatusOK, map[string]string{
				"Content-Type":  "application/json; charset=utf-8",
				"Etag":          `"1515151515678"`,
				"Last-Modified": "Fri, 05 Jan 2018 11:25:15 GMT",
			}, `{
				"data": {
					"id": "alpha",
					"deleted": true,
					"last_modified": 1515151515678
				}
			}`))
			Expect(handle(http.MethodGet, "/resources/alpha", ``).Code).To(Equal(http.StatusForbidden))
		})

		It("handles not-found", func() {
			// seed permission: alice could write beta
			Expect(txn.Perms.AddACEPrincipal("account:alice", permission.ACE{Perm: "write", Path: "/resources/beta"})).To(Succeed())
			Expect(handle(http.MethodDelete, "/resources/beta", ``)).To(MatchResponse(http.StatusNotFound, nil, `{
				"code": 404,
				"errno": 110,
				"error": "Not Found",
				"details": {"id": "beta", "resource_name": "resource"}
			}`))
		})

		It("validates object IDs", func() {
			Expect(handle(http.MethodDelete, "/resources/invalid%20id", ``)).To(MatchResponse(http.StatusBadRequest, nil, `{
				"code": 400,
				"errno": 107,
				"error": "Invalid parameters",
				"message": "path: Invalid object id",
				"details": [
					{"location": "path", "description": "Invalid object id"}
				]
			}`))
		})

		It("supports conditional rendering", func() {
			r := newRequest(http.MethodDelete, "/resources/alpha", ``)
			r.Header.Set("If-Match", `"1616161616000"`)
			Expect(serve(r)).To(MatchResponse(http.StatusPreconditionFailed, nil, `{
				"code": 412,
				"errno": 114,
				"error": "Precondition Failed",
				"message": "Resource was modified meanwhile",
				"details": {
					"existing": {
						"id": "alpha",
						"last_modified": 1515151515677
					}
				}
			}`))
		})
	})

	Describe("nested", func() {
		BeforeEach(func() {
			// seed resource: alpha
			Expect(handle(http.MethodPost, "/resources", `{
				"data": {"id":"alpha"},
				"permissions": {
					"nested:create":["account:bob", "principal:team"]
				}
			}`).Code).To(Equal(http.StatusCreated))

			// seed nested: omega
			txn.User = bob
			Expect(handle(http.MethodPost, "/resources/alpha/nested", `{
				"data": {"id":"omega"},
				"permissions": {"read":["principal:team"]}
			}`).Code).To(Equal(http.StatusCreated))

			// reset user
			txn.User = alice
		})

		Describe("GET /resource/RESID/nested", func() {
			It("authorizes", func() {
				// alice has access to /resources/alpha
				Expect(handle(http.MethodGet, "/resources/alpha/nested", ``).Code).To(Equal(http.StatusOK))

				// bob has no read access to /resources/alpha
				txn.User = bob
				Expect(handle(http.MethodGet, "/resources/alpha/nested", ``).Code).To(Equal(http.StatusForbidden))
			})

			It("fails when parent doesn't exist", func() {
				// without permission
				Expect(handle(http.MethodGet, "/resources/missing/nested", ``).Code).To(Equal(http.StatusForbidden))

				// with global read permission
				Expect(txn.Perms.AddACEPrincipal("principal:team", permission.ACE{Perm: "read"})).To(Succeed())
				Expect(handle(http.MethodGet, "/resources/missing/nested", ``)).To(MatchResponse(http.StatusNotFound, nil, `{
					"code": 404,
					"errno": 111,
					"error": "Not Found",
					"details": {
						"id": "missing",
						"resource_name": "resource"
					}
				}`))
			})
		})

		Describe("GET /resource/RESID/nested/ID", func() {
			It("authorizes", func() {
				// bob can write omega (direct)
				txn.User = bob
				Expect(handle(http.MethodGet, "/resources/alpha/nested/omega", ``).Code).To(Equal(http.StatusOK))

				// alice can write omega (inherited)
				txn.User = alice
				Expect(handle(http.MethodGet, "/resources/alpha/nested/omega", ``).Code).To(Equal(http.StatusOK))

				// claire can read omega as member of team
				txn.User = claire
				Expect(handle(http.MethodGet, "/resources/alpha/nested/omega", ``).Code).To(Equal(http.StatusOK))

				// daniel cannot access omega
				txn.User = mock.User("account:daniel")
				Expect(handle(http.MethodGet, "/resources/alpha/nested/omega", ``).Code).To(Equal(http.StatusForbidden))
			})

			It("responds", func() {
				Expect(handle(http.MethodGet, "/resources/alpha/nested/omega", ``)).To(MatchResponse(http.StatusOK, nil, `{
					"data": {
						"id": "omega",
						"last_modified": 1515151515677
					},
					"permissions": {
						"read": ["principal:team"],
						"write": ["account:bob"]
					}
				}`))
			})

			It("fails when parent doesn't exist", func() {
				// without permission
				Expect(handle(http.MethodGet, "/resources/missing/nested/omega", ``).Code).To(Equal(http.StatusForbidden))

				// with global read permission
				Expect(txn.Perms.AddACEPrincipal("principal:team", permission.ACE{Perm: "read"})).To(Succeed())
				Expect(handle(http.MethodGet, "/resources/missing/nested/omega", ``)).To(MatchResponse(http.StatusNotFound, nil, `{
					"code": 404,
					"errno": 111,
					"error": "Not Found",
					"details": {
						"id": "missing",
						"resource_name": "resource"
					}
				}`))
			})
		})

		Describe("POST /resource/RESID/nested", func() {
			It("authorizes", func() {
				// bob can create nested
				txn.User = bob
				Expect(handle(http.MethodPost, "/resources/alpha/nested", `{"data": {}}`).Code).To(Equal(http.StatusCreated))

				// alice can write alpha
				txn.User = alice
				Expect(handle(http.MethodPost, "/resources/alpha/nested", `{"data": {}}`).Code).To(Equal(http.StatusCreated))

				// claire can create as member of principal:team
				txn.User = claire
				Expect(handle(http.MethodPost, "/resources/alpha/nested", `{"data": {}}`).Code).To(Equal(http.StatusCreated))

				// daniel cannot create
				txn.User = mock.User("account:daniel")
				Expect(handle(http.MethodPost, "/resources/alpha/nested", `{"data": {}}`).Code).To(Equal(http.StatusForbidden))
			})

			It("responds", func() {
				Expect(handle(http.MethodPost, "/resources/alpha/nested", `{"data": {}}`)).To(MatchResponse(http.StatusCreated, nil, `{
					"data": {
						"id": "EPR.ID",
						"last_modified": 1515151515678
					},
					"permissions": {
						"write": ["account:alice"]
					}
				}`))
			})

			It("fails when parent doesn't exist", func() {
				// without permission
				Expect(handle(http.MethodPost, "/resources/missing/nested", `{"data":{}}`).Code).To(Equal(http.StatusForbidden))

				// with global write permission
				Expect(txn.Perms.AddACEPrincipal("principal:team", permission.ACE{Perm: "write"})).To(Succeed())
				Expect(handle(http.MethodPost, "/resources/missing/nested", `{"data":{}}`)).To(MatchResponse(http.StatusNotFound, nil, `{
					"code": 404,
					"errno": 111,
					"error": "Not Found",
					"details": {
						"id": "missing",
						"resource_name": "resource"
					}
				}`))
			})
		})

		Describe("PUT /resource/RESID/nested/ID", func() {
			It("authorizes", func() {
				// bob has direct write access to omega
				txn.User = bob
				Expect(handle(http.MethodPut, "/resources/alpha/nested/omega", `{"data":{}}`).Code).To(Equal(http.StatusOK))

				// alice has inherited write access to omega
				txn.User = alice
				Expect(handle(http.MethodPut, "/resources/alpha/nested/omega", `{"data":{}}`).Code).To(Equal(http.StatusOK))

				// claire cannot write to omega
				txn.User = claire
				Expect(handle(http.MethodPut, "/resources/alpha/nested/omega", `{"data":{}}`).Code).To(Equal(http.StatusForbidden))

				// daniel cannot access omega
				txn.User = mock.User("account:daniel")
				Expect(handle(http.MethodPut, "/resources/alpha/nested/omega", `{"data":{}}`).Code).To(Equal(http.StatusForbidden))
			})

			It("fails when parent doesn't exist", func() {
				// without permission
				Expect(handle(http.MethodPut, "/resources/missing/nested/omega", `{"data":{}}`).Code).To(Equal(http.StatusForbidden))

				// with global write permission
				Expect(txn.Perms.AddACEPrincipal("principal:team", permission.ACE{Perm: "write"})).To(Succeed())
				Expect(handle(http.MethodPut, "/resources/missing/nested/omega", `{"data":{}}`)).To(MatchResponse(http.StatusNotFound, nil, `{
					"code": 404,
					"errno": 111,
					"error": "Not Found",
					"details": {
						"id": "missing",
						"resource_name": "resource"
					}
				}`))
			})
		})

		Describe("PATCH /resource/RESID/nested/ID", func() {
			It("authorizes", func() {
				// bob has direct write access to omega
				txn.User = bob
				Expect(handle(http.MethodPatch, "/resources/alpha/nested/omega", `{"data":{}}`).Code).To(Equal(http.StatusOK))

				// alice has inherited write access to omega
				txn.User = alice
				Expect(handle(http.MethodPatch, "/resources/alpha/nested/omega", `{"data":{}}`).Code).To(Equal(http.StatusOK))

				// claire cannot write to omega
				txn.User = claire
				Expect(handle(http.MethodPatch, "/resources/alpha/nested/omega", `{"data":{}}`).Code).To(Equal(http.StatusForbidden))

				// daniel cannot access omega
				txn.User = mock.User("account:daniel")
				Expect(handle(http.MethodPatch, "/resources/alpha/nested/omega", `{"data":{}}`).Code).To(Equal(http.StatusForbidden))
			})

			It("fails when parent doesn't exist", func() {
				// without permission
				Expect(handle(http.MethodPatch, "/resources/missing/nested/omega", `{"data":{}}`).Code).To(Equal(http.StatusForbidden))

				// with global write permission
				Expect(txn.Perms.AddACEPrincipal("principal:team", permission.ACE{Perm: "write"})).To(Succeed())
				Expect(handle(http.MethodPatch, "/resources/missing/nested/omega", `{"data":{}}`)).To(MatchResponse(http.StatusNotFound, nil, `{
					"code": 404,
					"errno": 111,
					"error": "Not Found",
					"details": {
						"id": "missing",
						"resource_name": "resource"
					}
				}`))
			})
		})

		Describe("DELETE /resource/RESID/nested/ID", func() {
			It("authorizes", func() {
				// claire cannot write to omega
				txn.User = claire
				Expect(handle(http.MethodDelete, "/resources/alpha/nested/omega", ``).Code).To(Equal(http.StatusForbidden))

				// dan cannot access omega
				txn.User = mock.User("account:daniel")
				Expect(handle(http.MethodDelete, "/resources/alpha/nested/omega", ``).Code).To(Equal(http.StatusForbidden))

				// alice has inherited write access to omega
				txn.User = alice
				Expect(handle(http.MethodDelete, "/resources/alpha/nested/omega", ``).Code).To(Equal(http.StatusOK))
			})

			It("fails when parent doesn't exist", func() {
				// without permission
				Expect(handle(http.MethodDelete, "/resources/missing/nested/omega", ``).Code).To(Equal(http.StatusForbidden))

				// with global write permission
				Expect(txn.Perms.AddACEPrincipal("principal:team", permission.ACE{Perm: "write"})).To(Succeed())
				Expect(handle(http.MethodDelete, "/resources/missing/nested/omega", ``)).To(MatchResponse(http.StatusNotFound, nil, `{
					"code": 404,
					"errno": 111,
					"error": "Not Found",
					"details": {
						"id": "missing",
						"resource_name": "resource"
					}
				}`))
			})
		})
	})
})

// --------------------------------------------------------------------

func ContainHeaders(headers map[string]string) types.GomegaMatcher {
	matchers := make([]types.GomegaMatcher, 0, len(headers))
	for k, v := range headers {
		matchers = append(matchers, HaveKeyWithValue(k, v))
	}
	return WithTransform(func(actual http.Header) map[string]string {
		hm := make(map[string]string, len(actual))
		for k := range actual {
			hm[k] = actual.Get(k)
		}
		return hm
	}, And(matchers...))
}

func MatchBody(body string) types.GomegaMatcher {
	if body == "" {
		return BeEmpty()
	}
	return MatchJSON(body)
}

func MatchResponse(code int, headers map[string]string, body string) types.GomegaMatcher {
	return And(
		WithTransform(func(w *httptest.ResponseRecorder) int { return w.Code }, Equal(code)),
		WithTransform(func(w *httptest.ResponseRecorder) http.Header { return w.Header() }, ContainHeaders(headers)),
		WithTransform(func(w *httptest.ResponseRecorder) []byte { return w.Body.Bytes() }, MatchBody(body)),
	)
}
