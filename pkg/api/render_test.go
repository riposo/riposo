package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/schema"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
)

var _ = Describe("Rebder", func() {
	var w *httptest.ResponseRecorder

	BeforeEach(func() {
		w = httptest.NewRecorder()
	})

	It("renders structs", func() {
		api.Render(w, &mockRenderType{Name: "foo"})
		Expect(w.Code).To(Equal(http.StatusOK))
		Expect(w.Header()).To(Equal(http.Header{
			"Content-Type": {"application/json; charset=utf-8"},
		}))
		Expect(w.Body.String()).To(MatchJSON(`{"name":"foo"}`))
	})

	It("renders schema errors", func() {
		api.Render(w, schema.Forbidden)
		Expect(w.Code).To(Equal(http.StatusForbidden))
		Expect(w.Body.String()).To(MatchJSON(`{
			"code": 403,
			"errno": 121,
			"error": "Forbidden",
			"message": "This user cannot access this resource."
		}`))
	})

	It("renders other errors", func() {
		api.Render(w, fmt.Errorf("doh!"))
		Expect(w.Code).To(Equal(http.StatusInternalServerError))
		Expect(w.Body.String()).To(MatchJSON(`{
			"code": 500,
			"errno": 999,
			"error": "Internal Server Error",
			"message": "doh!"
		}`))
	})

	It("renders nil", func() {
		api.Render(w, nil)
		Expect(w.Code).To(Equal(http.StatusOK))
		Expect(w.Body.Len()).To(Equal(0))
	})

	It("renders custom status", func() {
		api.Render(w, &customStatusType{Status: http.StatusTeapot})
		Expect(w.Code).To(Equal(http.StatusTeapot))
		Expect(w.Body.String()).To(MatchJSON(`{ "code": 418 }`))

		w = httptest.NewRecorder()
		api.Render(w, &customStatusType{})
		Expect(w.Code).To(Equal(http.StatusOK))
		Expect(w.Body.String()).To(MatchJSON(`{ "code": 0 }`))
	})

	It("fails gracefully", func() {
		api.Render(w, &nonMarshalableType{})
		Expect(w.Code).To(Equal(http.StatusInternalServerError))
		Expect(w.Body.String()).To(MatchJSON(`{
			"code": 500,
			"errno": 999,
			"error": "Internal Server Error",
			"message": "json: error calling MarshalJSON for type *api_test.nonMarshalableType: cannot marshal type"
		}`))
	})
})

type mockRenderType struct {
	Name string `json:"name"`
}

type customStatusType struct {
	Status int `json:"code"`
}

func (s *customStatusType) HTTPStatus() int { return s.Status }

type nonMarshalableType struct{}

func (*nonMarshalableType) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal type")
}
