package batch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

var skipHeaders = map[string]struct{}{
	"Accept-Encoding":   {},
	"Connection":        {},
	"Content-Encoding":  {},
	"Keep-Alive":        {},
	"Transfer-Encoding": {},
	"Upgrade":           {},
}

// Request is a batch request.
type Request struct {
	Defaults *RequestPart
	Requests []RequestPart
}

func (r *Request) containsRecursive() bool {
	if r.Defaults != nil && r.Defaults.isRecursive() {
		return true
	}

	for _, part := range r.Requests {
		if part.isRecursive() {
			return true
		}
	}
	return false
}

// RequestPart contains sub-request information of a batch request.
type RequestPart struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Body    json.RawMessage   `json:"body"`
	Headers map[string]string `json:"headers"`
}

// Norm merges defaults and normalizes part.
func (p *RequestPart) Norm(namespace string, defaults *RequestPart) {
	if defaults == nil {
		return
	}

	if p.Method == "" {
		p.Method = defaults.Method
	}
	if p.Path == "" {
		p.Path = defaults.Path
	}
	if len(p.Body) == 0 {
		p.Body = append(p.Body[:0], defaults.Body...)
	}
	if n := len(defaults.Headers); n != 0 {
		if p.Headers == nil {
			p.Headers = make(map[string]string, n)
		}
		for key, val := range defaults.Headers {
			if _, ok := p.Headers[key]; !ok {
				p.Headers[key] = val
			}
		}
	}

	if strings.HasPrefix(p.Path, namespace+"/") {
		p.Path = strings.TrimPrefix(p.Path, namespace)
	}
}

func (p *RequestPart) isRecursive() bool {
	return strings.HasPrefix(p.Path, "/batch")
}

func (p *RequestPart) httpRequest(ctx context.Context, parent http.Header) (*http.Request, error) {
	switch strings.ToUpper(p.Method) {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		// OK
	default:
		return nil, fmt.Errorf("invalid method %q", p.Method)
	}

	hr, err := http.NewRequestWithContext(ctx, p.Method, p.Path, bytes.NewReader(p.Body))
	if err != nil {
		if uerr := new(url.Error); errors.As(err, &uerr) {
			return nil, fmt.Errorf("invalid path %q", p.Path)
		}
		return nil, err
	}

	for key := range parent {
		if _, ok := skipHeaders[key]; !ok {
			hr.Header.Set(key, parent.Get(key))
		}
	}
	for key, val := range p.Headers {
		hr.Header.Set(key, val)
	}
	return hr, nil
}
