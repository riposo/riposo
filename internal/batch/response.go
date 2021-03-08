package batch

import (
	"encoding/json"
	"net/http"
)

// Response is a batch response.
type Response struct {
	Responses []ResponsePart `json:"responses"`
}

// ResponsePart contains sub-response information of a batch response.
type ResponsePart struct {
	Status  int               `json:"status"`
	Path    string            `json:"path"`
	Body    json.RawMessage   `json:"body,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// ResponseRecorder records responses.
type ResponseRecorder struct {
	code   int
	header http.Header
	offset int
	body   []byte
}

// Reset resets the recorder.
func (w *ResponseRecorder) Reset() {
	*w = ResponseRecorder{body: w.body[:0]}
}

// WriteHeader implements http.ResponseWriter interface.
func (w *ResponseRecorder) WriteHeader(code int) {
	w.code = code
	w.offset = len(w.body)
}

// Header implements http.ResponseWriter interface.
func (w *ResponseRecorder) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

// Write implements http.ResponseWriter interface.
func (w *ResponseRecorder) Write(p []byte) (int, error) {
	if w.code == 0 {
		w.WriteHeader(http.StatusOK)
	}
	w.body = append(w.body, p...)
	return len(p), nil
}

func (w *ResponseRecorder) appendTo(r *Response, path string) {
	code := w.code
	w.code = 0

	headers := make(map[string]string, len(w.header))
	for key := range w.header {
		headers[key] = w.header.Get(key)
		delete(w.header, key)
	}

	r.Responses = append(r.Responses, ResponsePart{
		Status:  code,
		Path:    path,
		Body:    w.body[w.offset:len(w.body)],
		Headers: headers,
	})
}
