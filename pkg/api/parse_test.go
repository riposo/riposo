package api_test

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/pkg/api"
)

var _ = Describe("Parse", func() {
	type target struct {
		Status string
	}

	It("parses plain", func() {
		v := new(target)
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"status":"PLAIN OK"}`))
		Expect(Parse(r, v)).To(Succeed())
		Expect(v).To(Equal(&target{Status: "PLAIN OK"}))
	})

	It("decodes gzip", func() {
		b := new(bytes.Buffer)
		z := gzip.NewWriter(b)
		Expect(z.Write([]byte(`{"status":"GZIP OK"}`))).To(Equal(20))
		Expect(z.Close()).To(Succeed())

		v := new(target)
		r := httptest.NewRequest(http.MethodPost, "/", b)
		r.Header.Set("Content-Encoding", "gzip")
		Expect(Parse(r, v)).To(Succeed())
		Expect(v).To(Equal(&target{Status: "GZIP OK"}))
	})

	It("decodes flate", func() {
		b := new(bytes.Buffer)
		z, err := flate.NewWriter(b, flate.DefaultCompression)
		Expect(err).NotTo(HaveOccurred())
		Expect(z.Write([]byte(`{"status":"FLATE OK"}`))).To(Equal(21))
		Expect(z.Close()).To(Succeed())

		v := new(target)
		r := httptest.NewRequest(http.MethodPost, "/", b)
		r.Header.Set("Content-Encoding", "flate")
		Expect(Parse(r, v)).To(Succeed())
		Expect(v).To(Equal(&target{Status: "FLATE OK"}))
	})

	It("may return schema compatible errors", func() {
		v := new(target)
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`not even JSON`))
		Expect(Parse(r, v)).To(MatchError(`body: Invalid JSON`))

		r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`not a GZIP encoded payload`))
		r.Header.Set("Content-Encoding", "gzip")
		Expect(Parse(r, v)).To(MatchError(`body: Invalid gzip encoding`))
	})
})
