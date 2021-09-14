package batch_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/riposo/riposo/internal/batch"
)

func BenchmarkHandle(b *testing.B) {
	handler := Handler("/v1", mockMux)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := httptest.NewRequest(http.MethodPost, "/batch", strings.NewReader(mockValidRequest))
		w := httptest.NewRecorder()
		b.StartTimer()

		handler.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			b.Fatalf("expected %d, but received %d - %s", http.StatusOK, w.Code, w.Body.String())
		}
	}
}
