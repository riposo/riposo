package server

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
)

type logger struct {
	*log.Logger
}

func (l *logger) NewLogEntry(r *http.Request) middleware.LogEntry {
	remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)

	e := &logEntry{parent: l}
	if remoteIP != "" {
		e.printf(cRed, "host=%s", remoteIP)
		e.buf.WriteByte(' ')
	}
	e.printf(cYellow, "method=%s", r.Method)
	e.buf.WriteByte(' ')
	e.printf(nil, "path=%s", r.RequestURI)
	e.buf.WriteByte(' ')
	e.printf(cBlue, "proto=%s", r.Proto)
	return e
}

type logEntry struct {
	parent *logger
	buf    bytes.Buffer
}

func (e *logEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	e.buf.WriteByte(' ')
	e.printf(cGreen, "status=%03d", status)

	e.buf.WriteByte(' ')
	e.printf(cMagenta, "bytes=%d", bytes)

	e.buf.WriteByte(' ')
	e.printf(cCyan, "taken=%.3f", elapsed.Seconds())

	e.parent.Println(e.buf.String())
}

func (*logEntry) Panic(v interface{}, stack []byte) {
	middleware.PrintPrettyStack(v)
}

func (e *logEntry) printf(color []byte, s string, args ...interface{}) {
	if middleware.IsTTY && len(color) != 0 {
		e.buf.Write(color)
	}
	fmt.Fprintf(&e.buf, s, args...)
	if middleware.IsTTY && len(color) != 0 {
		e.buf.Write(cReset)
	}
}

var (
	cRed     = []byte{'\033', '[', '3', '1', ';', '1', 'm'}
	cGreen   = []byte{'\033', '[', '3', '2', ';', '1', 'm'}
	cYellow  = []byte{'\033', '[', '3', '3', ';', '1', 'm'}
	cBlue    = []byte{'\033', '[', '3', '4', ';', '1', 'm'}
	cMagenta = []byte{'\033', '[', '3', '5', ';', '1', 'm'}
	cCyan    = []byte{'\033', '[', '3', '6', ';', '1', 'm'}
	cReset   = []byte{'\033', '[', '0', 'm'}
)
