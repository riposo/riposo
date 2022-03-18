package params

import (
	"errors"
	"net/url"
	"strconv"

	"github.com/riposo/riposo/pkg/schema"
)

var errInvalidToken = errors.New("_token has invalid content")

type field struct {
	Name string
	schema.Value
}

// Params are parsed query params.
type Params struct {
	Condition Condition
	Sort      []SortOrder
	Limit     int
	Token     *Pagination
}

// Parse parses query params.
func Parse(query url.Values, maxLimit int) (*Params, error) {
	pms := &Params{Limit: maxLimit}
	for key := range query {
		switch key {
		case "_limit":
			pms.Limit = ParseLimit(query.Get(key), maxLimit)
		case "_sort":
			pms.Sort = ParseSort(query.Get(key))
		case "_token":
			if token, err := ParseToken(query.Get(key)); errors.Is(err, ErrNoToken) {
				// skip
			} else if err != nil {
				return nil, errInvalidToken
			} else {
				pms.Token = token
			}
		case "_before":
			if filter := ParseFilter("lt_last_modified", query.Get(key)); filter.isValid() {
				pms.Condition = append(pms.Condition, filter)
			}
		case "_since":
			if filter := ParseFilter("gt_last_modified", query.Get(key)); filter.isValid() {
				pms.Condition = append(pms.Condition, filter)
			}
		case "_fields":
			// TODO: respect field limitation, eventually
		default:
			if filter := ParseFilter(key, query.Get(key)); filter.isValid() {
				pms.Condition = append(pms.Condition, filter)
			}
		}
	}
	return pms, nil
}

// NextPageURL generates a next-page paginated URL.
func (p *Params) NextPageURL(u *url.URL, nonce string, lastObj *schema.Object) (*url.URL, error) {
	token, err := newPagination(nonce, lastObj, p.Sort).Encode()
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("_limit", strconv.Itoa(p.Limit))
	q.Set("_token", token)

	r := new(url.URL)
	*r = *u
	r.RawQuery = q.Encode()
	return r, nil
}

// ParseLimit parses the limit.
func ParseLimit(s string, max int) int {
	n, _ := strconv.Atoi(s)
	if n < 0 {
		n = 0
	}

	if max > 0 && (n < 1 || n > max) {
		n = max
	}
	return n
}
