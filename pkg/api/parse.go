package api

import (
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"

	"github.com/riposo/riposo/pkg/schema"
)

// Parse parses a request body into v.
func Parse(r *http.Request, v interface{}) error {
	// read request body
	var body io.ReadCloser
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		rc, err := gzip.NewReader(r.Body)
		if err != nil {
			return schema.BadRequest(err)
		}
		defer rc.Close()

		body = rc
	case "flate":
		rc := flate.NewReader(r.Body)
		defer rc.Close()

		body = rc
	default:
		body = r.Body
	}

	if err := json.NewDecoder(body).Decode(v); err != nil && err != io.EOF {
		return schema.BadRequest(err)
	}
	return nil
}
