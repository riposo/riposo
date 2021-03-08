package batch

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/schema"
)

// Handler returns a new handler.
func Handler(namespace string, mux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// parse request
		var req Request
		if err := api.Parse(r, &req); err != nil {
			api.Render(w, err)
			return
		}
		if req.containsRecursive() {
			api.Render(w, schema.InvalidBody("requests", "Recursive call on /batch endpoint is forbidden."))
			return
		}

		// unset the batch route context
		ctx := context.WithValue(r.Context(), chi.RouteCtxKey, nil)

		// parse/collect sub-requests
		sub := make([]*http.Request, 0, len(req.Requests))
		for pos, part := range req.Requests {
			part.Norm(namespace, req.Defaults)

			sreq, err := part.httpRequest(ctx, r.Header)
			if err != nil {
				api.Render(w, schema.InvalidBody("requests."+strconv.Itoa(pos), err.Error()))
				return
			}
			sub = append(sub, sreq)
		}

		// record/collect responses
		rec := poolRecorder()
		defer releaseRecorder(rec)

		res := &Response{Responses: make([]ResponsePart, 0, len(req.Requests))}
		for _, sr := range sub {
			mux.ServeHTTP(rec, sr)
			rec.appendTo(res, namespace+sr.URL.String())
		}

		api.Render(w, res)
	})
}
