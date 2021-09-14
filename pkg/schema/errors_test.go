package schema_test

import (
	"encoding/json"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/pkg/schema"
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
		Expect(json.Marshal(BadRequest(err))).To(MatchJSON(`{
			"code": 400,
			"errno": 107,
			"error": "Invalid parameters",
			"message": "body: Invalid JSON",
			"details": [
				{ "location": "body", "description": "Invalid JSON" }
			]
		}`))

		err = json.Unmarshal([]byte(`NOT JSON`), &data)
		Expect(json.Marshal(BadRequest(err).Details)).To(MatchJSON(`[
			{ "location": "body", "description": "Invalid JSON" }
		]`))

		err = json.Unmarshal([]byte(`"BAD"`), &data)
		Expect(json.Marshal(BadRequest(err).Details)).To(MatchJSON(`[
			{ "location": "body", "description": "Invalid JSON" }
		]`))

		err = json.Unmarshal([]byte(`{"x": {"y": 33}}`), &data)
		Expect(json.Marshal(BadRequest(err).Details)).To(MatchJSON(`[
			{ "location": "body", "description": "Invalid type", "name": "x.y" }
		]`))
	})
})
