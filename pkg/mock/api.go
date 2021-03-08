package mock

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/riposo"
)

// Txn inits a mock API transaction.
func Txn() *api.Txn {
	hlp := Helpers()
	txn, err := api.NewTxn(context.Background(), Conns(hlp), hlp)
	if err != nil {
		panic(err)
	}
	return txn
}

// Request creates a HTTP mock request with a transaction context.
func Request(txn *api.Txn, method, path string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, path, body)
	ctx := api.WithTxn(req.Context(), txn)
	return req.WithContext(ctx)
}

// User inits a new mock user with a given user ID.
func User(id string, principals ...string) *api.User {
	user := &api.User{ID: id}
	if user.ID == "" {
		user.ID = riposo.Everyone
	}
	switch user.ID {
	case riposo.Everyone:
		user.Principals = append(user.Principals, riposo.Everyone)
	case riposo.Authenticated:
		user.Principals = append(user.Principals, riposo.Authenticated, riposo.Everyone)
	default:
		user.Principals = append(user.Principals, user.ID, riposo.Authenticated, riposo.Everyone)
	}
	user.Principals = append(user.Principals, principals...)
	return user
}
