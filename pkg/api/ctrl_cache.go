package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"
)

func setCacheHeaders(h http.Header, r *http.Request, modTime riposo.Epoch) {
	h.Set("Last-Modified", modTime.HTTPFormat())
	h.Set("Etag", modTime.ETag())
	if m := r.Method; m == http.MethodGet || m == http.MethodHead {
		h.Set("Cache-Control", `no-cache, no-store, no-transform, must-revalidate, private, max-age=0`)
	}
}

func renderConditional(h http.Header, r *http.Request, epoch riposo.Epoch, obj *schema.Object) *schema.Error {
	setCacheHeaders(h, r, epoch)
	return condStatus(r, epoch, obj)
}

func condStatus(r *http.Request, epoch riposo.Epoch, obj *schema.Object) *schema.Error {
	etag := epoch.ETag()
	mtime := epoch.Time()

	// This function carefully follows RFC 7232 section 6.
	ch := condIfMatch(r, etag)
	if ch == condNone {
		ch = condIfUnmodifiedSince(r, mtime)
	}
	if ch == condFalse {
		return schema.ModifiedMeanwhile(obj)
	}

	ch = condIfNoneMatch(r, etag)
	if ch == condNone {
		ch = checkIfModifiedSince(r, mtime)
	}
	if ch == condFalse {
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			return schema.NotModified
		}
		return schema.ModifiedMeanwhile(obj)
	}

	return nil
}

// condResult is the result of an HTTP request precondition check.
// See https://tools.ietf.org/html/rfc7232 section 3.
type condResult uint8

const (
	condNone condResult = iota
	condTrue
	condFalse
)

func condIfMatch(r *http.Request, etag string) condResult {
	if val := r.Header.Get("If-Match"); val == "" {
		return condNone
	} else if val == "*" || strings.Contains(val, etag) {
		return condTrue
	}
	return condFalse
}

func condIfNoneMatch(r *http.Request, etag string) condResult {
	if val := r.Header.Get("If-None-Match"); val == "" {
		return condNone
	} else if val == "*" || strings.Contains(val, etag) {
		return condFalse
	}
	return condTrue
}

func condIfUnmodifiedSince(r *http.Request, modTime time.Time) condResult {
	val := r.Header.Get("If-Unmodified-Since")
	if val == "" || modTime.Unix() < 1 {
		return condNone
	}

	if t, err := http.ParseTime(val); err == nil {
		// The Date-Modified header truncates sub-second precision, so
		// use mtime < t+1s instead of mtime <= t to check for unmodified.
		if modTime.Before(t.Add(1 * time.Second)) {
			return condTrue
		}
		return condFalse
	}
	return condNone
}

func checkIfModifiedSince(r *http.Request, modTime time.Time) condResult {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return condNone
	}

	val := r.Header.Get("If-Modified-Since")
	if val == "" || modTime.Unix() < 1 {
		return condNone
	}

	t, err := http.ParseTime(val)
	if err != nil {
		return condNone
	}

	// The Date-Modified header truncates sub-second precision, so
	// use mtime < t+1s instead of mtime <= t to check for unmodified.
	if modTime.Before(t.Add(1 * time.Second)) {
		return condFalse
	}
	return condTrue
}
