package api

import (
	"context"
	"net/http"

	"github.com/riposo/riposo/pkg/conn"
	"github.com/riposo/riposo/pkg/conn/cache"
	"github.com/riposo/riposo/pkg/conn/permission"
	"github.com/riposo/riposo/pkg/conn/storage"
	"github.com/riposo/riposo/pkg/riposo"
	"go.uber.org/multierr"
)

type txnKey struct{}

// WithTxn adds the transaction as a value to the context.
func WithTxn(ctx context.Context, txn *Txn) context.Context {
	return context.WithValue(ctx, txnKey{}, txn)
}

// GetTxn extracts the current transaction from the request.
func GetTxn(req *http.Request) *Txn {
	if v := req.Context().Value(txnKey{}); v != nil {
		return v.(*Txn)
	}
	return nil
}

// Txn wraps an API transaction.
type Txn struct {
	context.Context
	Store   storage.Transaction
	Perms   permission.Transaction
	Cache   cache.Transaction
	Helpers riposo.Helpers
	User    *User
	Data    map[string]interface{}
}

// NewTxn inits a new transaction.
func NewTxn(ctx context.Context, cns *conn.Set, hlp riposo.Helpers) (*Txn, error) {
	store, err := cns.Store().Begin(ctx)
	if err != nil {
		return nil, err
	}

	perms, err := cns.Perms().Begin(ctx)
	if err != nil {
		_ = store.Rollback()
		return nil, err
	}

	cache, err := cns.Cache().Begin(ctx)
	if err != nil {
		_ = store.Rollback()
		_ = perms.Rollback()
		return nil, err
	}

	return &Txn{
		Context: ctx,
		Store:   store,
		Perms:   perms,
		Cache:   cache,
		Helpers: hlp,
		Data:    make(map[string]interface{}),
		User:    &User{ID: riposo.Everyone},
	}, nil
}

// Commit is used internally to commit all transactions.
func (t *Txn) Commit() error {
	return multierr.Combine(
		t.Store.Commit(),
		t.Perms.Commit(),
		t.Cache.Commit(),
	)
}

// Rollback is used internally to rollback all transactions.
func (t *Txn) Rollback() error {
	return multierr.Combine(
		t.Store.Rollback(),
		t.Perms.Rollback(),
		t.Cache.Rollback(),
	)
}
