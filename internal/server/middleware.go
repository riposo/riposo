package server

import (
	"errors"
	"net/http"

	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/auth"
	"github.com/riposo/riposo/pkg/conn"
	"github.com/riposo/riposo/pkg/riposo"
)

// Combined middleware to create transactions and authenticate users.
func transactional(cns *conn.Set, hlp *riposo.Helpers, am auth.Method) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// init transaction
			txn, err := api.NewTxn(r.Context(), cns, hlp)
			if err != nil {
				api.Render(w, err)
				return
			}
			defer txn.Abort()

			// embed txn into request context
			r = r.WithContext(api.WithTxn(r.Context(), txn))

			// authenticate user
			if user, err := am.Authenticate(r); errors.Is(err, auth.ErrUnauthenticated) {
				// pass-through
			} else if err != nil {
				api.Render(w, err)
				return
			} else if user != nil {
				txn.User = user
			}

			// update principals
			txn.User.Principals, err = txn.Perms.GetUserPrincipals(txn.User.ID)
			if err != nil {
				api.Render(w, err)
				return
			}

			// propagate downstream
			rw := &responseWriter{ResponseWriter: w, txn: txn}
			next.ServeHTTP(rw, r)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	txn  *api.Txn
	done bool
}

func (w *responseWriter) WriteHeader(code int) {
	if w.done {
		return
	}
	w.done = true

	if code < 500 {
		if err := w.txn.Commit(); err != nil {
			w.ResponseWriter.WriteHeader(http.StatusServiceUnavailable)
			return
		}
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(buf []byte) (int, error) {
	w.WriteHeader(http.StatusOK)
	return w.ResponseWriter.Write(buf)
}
