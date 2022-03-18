package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/conn/storage"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/slowhash"
)

func init() {
	Register("basic", func(_ context.Context, _ riposo.Helpers) (Method, error) {
		return Basic(), nil
	})
}

type basic struct{}

// Basic inits a HTTP basic auth Method.
func Basic() Method { return basic{} }

func (basic) Authenticate(r *http.Request) (*api.User, error) {
	// parse user credentials
	user, pass, ok := r.BasicAuth()
	if !ok {
		return nil, Errorf("no basic auth credentials")
	}

	// retrieve account from store
	txn := api.GetTxn(r)
	obj, err := txn.Store.Get(riposo.Path("/accounts/" + user))
	if errors.Is(err, storage.ErrNotFound) {
		return nil, Errorf("unknown user account")
	} else if err != nil {
		return nil, err
	}

	// decode account data
	var extra struct {
		Password string `json:"password"`
	}
	if err := obj.DecodeExtra(&extra); err != nil {
		return nil, err
	}

	// verify password
	if ok, err := slowhash.Verify(extra.Password, pass); err != nil {
		return nil, err
	} else if !ok {
		return nil, Errorf("invalid password")
	}

	return &api.User{ID: "account:" + user}, nil
}

func (basic) Close() error {
	return nil
}
