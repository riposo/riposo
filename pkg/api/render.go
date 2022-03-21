package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/riposo/riposo/pkg/bufferpool"
	"github.com/riposo/riposo/pkg/schema"
)

type customStatus interface {
	HTTPStatus() int
}

// Render renders any value as JSON.
func Render(w http.ResponseWriter, v interface{}) {
	switch vv := v.(type) {
	case nil:
		w.WriteHeader(http.StatusOK)
	case *schema.Error:
		renderError(w, vv)
	case error:
		renderError(w, vv)
	case customStatus:
		if err := render(w, vv.HTTPStatus(), vv); err != nil {
			renderError(w, err)
		}
	default:
		if err := render(w, http.StatusOK, vv); err != nil {
			renderError(w, err)
		}
	}
}

// renderError responds with an error.
func renderError(w http.ResponseWriter, err error) {
	resp := new(schema.Error)
	if !errors.As(err, &resp) {
		resp = schema.InternalError(err)
	}

	switch resp.StatusCode {
	case http.StatusNoContent, http.StatusNotModified:
		w.WriteHeader(resp.StatusCode)
	default:
		_ = render(w, resp.StatusCode, resp) // ignore errors
	}
}

func render(w http.ResponseWriter, code int, v interface{}) error {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(v); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if code != 0 {
		w.WriteHeader(code)
	}
	_, _ = buf.WriteTo(w) // ignore errors, header already written
	return nil
}
