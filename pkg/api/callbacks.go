package api

import (
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"
)

// Callbacks are a set of callbacks around a model.
type Callbacks interface {
	// Match returns true if callbacks are applicable for path.
	Match(path riposo.Path) bool
	// OnCreate is triggered on create.
	OnCreate(txn *Txn, path riposo.Path) CreateCallback
	// OnUpdate is triggered on update.
	OnUpdate(txn *Txn, path riposo.Path) UpdateCallback
	// OnPatch is triggered on patch.
	OnPatch(txn *Txn, path riposo.Path) PatchCallback
	// OnDelete is triggered on delete.
	OnDelete(txn *Txn, path riposo.Path) DeleteCallback
	// OnDeleteAll is triggered on delete-all.
	OnDeleteAll(txn *Txn, path riposo.Path) DeleteAllCallback
}

// CreateCallback instances run around create actions.
type CreateCallback interface {
	BeforeCreate(payload *schema.Resource) error
	AfterCreate(created *schema.Resource) error
}

// UpdateCallback instances run around update actions.
type UpdateCallback interface {
	BeforeUpdate(existing *schema.Object, payload *schema.Resource) error
	AfterUpdate(updated *schema.Resource) error
}

// PatchCallback instances run around patch actions.
type PatchCallback interface {
	BeforePatch(existing *schema.Object, payload *schema.Resource) error
	AfterPatch(patched *schema.Resource) error
}

// DeleteCallback instances run around delete actions.
type DeleteCallback interface {
	BeforeDelete(existing *schema.Object) error
	AfterDelete(deleted *schema.Object) error
}

// DeleteAllCallback instances run around delete-all actions.
type DeleteAllCallback interface {
	BeforeDeleteAll(objIDs []string) error
	AfterDeleteAll(modTime riposo.Epoch, deleted []riposo.Path) error
}

// NoopCallbacks is an embeddable noop callback type.
type NoopCallbacks struct{}

func (NoopCallbacks) Match(riposo.Path) bool                              { return false }
func (NoopCallbacks) OnCreate(_ *Txn, _ riposo.Path) CreateCallback       { return nil }
func (NoopCallbacks) OnUpdate(_ *Txn, _ riposo.Path) UpdateCallback       { return nil }
func (NoopCallbacks) OnPatch(_ *Txn, _ riposo.Path) PatchCallback         { return nil }
func (NoopCallbacks) OnDelete(_ *Txn, _ riposo.Path) DeleteCallback       { return nil }
func (NoopCallbacks) OnDeleteAll(_ *Txn, _ riposo.Path) DeleteAllCallback { return nil }
