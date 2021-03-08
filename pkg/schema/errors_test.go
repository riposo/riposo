package schema_test

import (
	"encoding/json"

	"github.com/riposo/riposo/pkg/schema"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BadRequest", func() {
	It("generates error messages", func() {
		var data struct {
			X struct {
				Y struct {
					Z string `json:"z"`
				} `json:"y"`
			} `json:"x"`
		}

		err := json.Unmarshal([]byte{}, &data)
		Expect(json.Marshal(schema.BadRequest(err))).To(MatchJSON(`{
			"code": 400,
			"errno": 107,
			"error": "Invalid parameters",
			"message": "body: Invalid JSON",
			"details": [
				{ "location": "body", "description": "Invalid JSON" }
			]
		}`))

		err = json.Unmarshal([]byte(`NOT JSON`), &data)
		Expect(json.Marshal(schema.BadRequest(err).Details)).To(MatchJSON(`[
			{ "location": "body", "description": "Invalid JSON" }
		]`))

		err = json.Unmarshal([]byte(`"BAD"`), &data)
		Expect(json.Marshal(schema.BadRequest(err).Details)).To(MatchJSON(`[
			{ "location": "body", "description": "Invalid JSON" }
		]`))

		err = json.Unmarshal([]byte(`{"x": {"y": 33}}`), &data)
		Expect(json.Marshal(schema.BadRequest(err).Details)).To(MatchJSON(`[
			{ "location": "body", "description": "Invalid type", "name": "x.y" }
		]`))
	})
})
