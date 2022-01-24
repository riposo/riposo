package params_test

import (
	"net/url"
	"testing"

	"github.com/riposo/riposo/pkg/schema"
	"github.com/tidwall/gjson"

	. "github.com/bsm/ginkgo/v2"
	. "github.com/bsm/gomega"
	. "github.com/riposo/riposo/pkg/params"
)

var _ = Describe("Params", func() {
	sampleURL := mustURL("https://example.com:8888/v1/buckets?_sort=field")

	It("parses", func() {
		Expect(Parse(sampleURL.Query(), 25)).To(Equal(&Params{
			Limit: 25,
			Sort: []SortOrder{
				{Field: "field"},
			},
		}))
	})

	It("parses _before and _since", func() {
		pms, err := Parse(url.Values{"_since": {"1515151515000"}, "_before": {"1616161616000"}}, 25)
		Expect(err).NotTo(HaveOccurred())
		Expect(pms.Condition).To(ConsistOf(
			Filter{
				Field: "last_modified", Operator: OperatorGT, Values: []schema.Value{
					{Type: gjson.Number, Raw: "1515151515000", Num: 1515151515000},
				},
			},
			Filter{
				Field: "last_modified", Operator: OperatorLT, Values: []schema.Value{
					{Type: gjson.Number, Raw: "1616161616000", Num: 1616161616000},
				},
			},
		))
	})

	It("parses filters", func() {
		Expect(Parse(nil, 25)).To(Equal(&Params{
			Limit: 25,
		}))
	})

	It("fails on bad tokens", func() {
		_, err := Parse(url.Values{"_token": {"bad"}}, 25)
		Expect(err).To(MatchError("_token has invalid content"))
	})

	It("generates NextPageURL", func() {
		pp, err := Parse(sampleURL.Query(), 20)
		Expect(err).NotTo(HaveOccurred())

		nu, err := pp.NextPageURL(sampleURL, "x", &schema.Object{
			Extra: []byte(`{"field": 33}`),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(nu.String()).To(Equal("https://example.com:8888/v1/buckets?_limit=20&_sort=field&_token=eyJub25jZSI6IngiLCJsYXN0X29iamVjdCI6eyJmaWVsZCI6MzN9fQ"))
	})
})

var _ = Describe("ParseLimit", func() {
	It("parses", func() {
		Expect(ParseLimit("", 25)).To(Equal(25))
		Expect(ParseLimit("0", 25)).To(Equal(25))
		Expect(ParseLimit("-1", 25)).To(Equal(25))
		Expect(ParseLimit("25", 25)).To(Equal(25))
		Expect(ParseLimit("99", 25)).To(Equal(25))
		Expect(ParseLimit("10", 25)).To(Equal(10))

		Expect(ParseLimit("", 0)).To(Equal(0))
		Expect(ParseLimit("-1", 0)).To(Equal(0))
		Expect(ParseLimit("25", 0)).To(Equal(25))
	})
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "pkg/params")
}

func mustURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
