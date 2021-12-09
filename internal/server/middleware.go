package server

import (
	"errors"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/auth"
	"github.com/riposo/riposo/pkg/conn"
	"github.com/riposo/riposo/pkg/riposo"
)

// Backoff and Retry-After header middleware.
func backoff(backoff time.Duration, backoffPct int, retryAfter time.Duration) func(http.Handler) http.Handler {
	var backoffVal string
	if sec := int64(backoff.Seconds()); sec >= 1 {
		backoffVal = strconv.FormatInt(sec, 10)
	} else {
		backoffPct = 0
	}
	backoffInc := new(uint32)

	var retryAfterVal string
	if sec := int64(retryAfter.Seconds()); sec >= 1 {
		retryAfterVal = strconv.FormatInt(sec, 10)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(&backoffWrapper{
				ResponseWriter: w,
				backoffVal:     backoffVal,
				backoffPct:     uint32(backoffPct),
				backoffInc:     backoffInc,
				retryAfterVal:  retryAfterVal,
			}, r)
		})
	}
}

type backoffWrapper struct {
	http.ResponseWriter
	wroteHeader bool

	backoffVal, retryAfterVal string
	backoffPct                uint32
	backoffInc                *uint32
}

func (w *backoffWrapper) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	if code >= http.StatusOK && code < http.StatusBadRequest {
		if w.showBackoff() {
			w.Header().Set("Backoff", w.backoffVal)
		}
	} else if code >= http.StatusInternalServerError {
		if w.showRetryAfter() {
			w.Header().Set("Retry-After", w.retryAfterVal)
		}
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *backoffWrapper) Write(buf []byte) (int, error) {
	w.WriteHeader(http.StatusOK)
	return w.ResponseWriter.Write(buf)
}

func (w *backoffWrapper) showBackoff() bool {
	if max := uint32(100); w.backoffPct > 0 && w.backoffPct < max {
		inc := atomic.AddUint32(w.backoffInc, 1) % max
		return (w.backoffPct*inc)%max+w.backoffPct >= max
	}
	return w.backoffVal != ""
}

func (w *backoffWrapper) showRetryAfter() bool {
	return w.retryAfterVal != ""
}

// ----------------------------------------------------------------------------

// Combined middleware to create transactions and authenticate users.
func transactional(cns *conn.Set, hlp riposo.Helpers, am auth.Method) func(http.Handler) http.Handler {
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
			rw := &transactionalWrapper{ResponseWriter: w, txn: txn}
			next.ServeHTTP(rw, r)
		})
	}
}

type transactionalWrapper struct {
	http.ResponseWriter
	txn         *api.Txn
	wroteHeader bool
}

func (w *transactionalWrapper) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	if code < http.StatusInternalServerError {
		if err := w.txn.Commit(); err != nil {
			w.ResponseWriter.WriteHeader(http.StatusServiceUnavailable)
			return
		}
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *transactionalWrapper) Write(buf []byte) (int, error) {
	w.WriteHeader(http.StatusOK)
	return w.ResponseWriter.Write(buf)
}
