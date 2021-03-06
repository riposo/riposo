package api

import (
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"
)

// Actions wraps a model with a callback chain.
type Actions interface {
	Get(txn *Txn, path riposo.Path) (*schema.Resource, error)
	Create(txn *Txn, path riposo.Path, payload *schema.Resource) error
	Update(txn *Txn, path riposo.Path, exst *schema.Object, payload *schema.Resource) (*schema.Resource, error)
	Patch(txn *Txn, path riposo.Path, exst *schema.Object, payload *schema.Resource) (*schema.Resource, error)
	Delete(txn *Txn, path riposo.Path, exst *schema.Object) (*schema.Object, error)
	DeleteAll(txn *Txn, path riposo.Path, exst []*schema.Object) (riposo.Epoch, error)
}

// NewActions wraps a model with callbacks.
func NewActions(mod Model, cbs []Callbacks) Actions {
	return &actions{mod: mod, cbs: cbs}
}

type actions struct {
	mod Model
	cbs []Callbacks
}

func (a *actions) Get(txn *Txn, path riposo.Path) (*schema.Resource, error) {
	return a.mod.Get(txn, path)
}

func (a *actions) Create(txn *Txn, path riposo.Path, payload *schema.Resource) error {
	// prepare callbacks
	callbacks := a.prepareCallbacks(func(cb Callbacks) interface{} {
		return cb.OnCreate(txn, path)
	})
	defer callbacks.Release()

	// run before callbacks
	for _, c := range callbacks.S {
		if err := c.(CreateCallback).BeforeCreate(payload); err != nil {
			return err
		}
	}

	// create actions
	err := a.mod.Create(txn, path, payload)
	if err != nil {
		return err
	}

	// run after callbacks in reverse order
	for i := len(callbacks.S) - 1; i >= 0; i-- {
		if err := callbacks.S[i].(CreateCallback).AfterCreate(payload); err != nil {
			return err
		}
	}

	return nil
}

func (a *actions) Update(txn *Txn, path riposo.Path, exst *schema.Object, payload *schema.Resource) (*schema.Resource, error) {
	// prepare callbacks
	callbacks := a.prepareCallbacks(func(cb Callbacks) interface{} {
		return cb.OnUpdate(txn, path)
	})
	defer callbacks.Release()

	// run before callbacks
	for _, c := range callbacks.S {
		if err := c.(UpdateCallback).BeforeUpdate(exst, payload); err != nil {
			return nil, err
		}
	}

	// update actions & permissions
	if err := a.mod.Update(txn, path, exst, payload); err != nil {
		return nil, err
	}

	// fetch updated permissions
	ps, err := txn.Perms.GetPermissions(path)
	if err != nil {
		return nil, err
	}

	// prepare response
	res := &schema.Resource{Data: exst, Permissions: ps}

	// run after callbacks in reverse order
	for i := len(callbacks.S) - 1; i >= 0; i-- {
		if err := callbacks.S[i].(UpdateCallback).AfterUpdate(res); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (a *actions) Patch(txn *Txn, path riposo.Path, exst *schema.Object, payload *schema.Resource) (*schema.Resource, error) {
	// prepare callbacks
	callbacks := a.prepareCallbacks(func(cb Callbacks) interface{} {
		return cb.OnPatch(txn, path)
	})
	defer callbacks.Release()

	// run before callbacks
	for _, c := range callbacks.S {
		if err := c.(PatchCallback).BeforePatch(exst, payload); err != nil {
			return nil, err
		}
	}

	// patch actions & permissions
	if err := a.mod.Patch(txn, path, exst, payload); err != nil {
		return nil, err
	}

	// fetch updated permissions
	ps, err := txn.Perms.GetPermissions(path)
	if err != nil {
		return nil, err
	}

	// prepare response
	res := &schema.Resource{Data: exst, Permissions: ps}

	// run after callbacks in reverse order
	for i := len(callbacks.S) - 1; i >= 0; i-- {
		if err := callbacks.S[i].(PatchCallback).AfterPatch(res); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (a *actions) Delete(txn *Txn, path riposo.Path, exst *schema.Object) (*schema.Object, error) {
	// prepare callbacks
	callbacks := a.prepareCallbacks(func(cb Callbacks) interface{} {
		return cb.OnDelete(txn, path)
	})
	defer callbacks.Release()

	// run before callbacks
	for _, c := range callbacks.S {
		if err := c.(DeleteCallback).BeforeDelete(exst); err != nil {
			return nil, err
		}
	}

	// delete actions
	deleted, err := a.mod.Delete(txn, path, exst)
	if err != nil {
		return nil, err
	}

	// run after callbacks in reverse order
	for i := len(callbacks.S) - 1; i >= 0; i-- {
		if err := callbacks.S[i].(DeleteCallback).AfterDelete(deleted); err != nil {
			return nil, err
		}
	}

	return deleted, nil
}

func (a *actions) DeleteAll(txn *Txn, path riposo.Path, objs []*schema.Object) (riposo.Epoch, error) {
	// prepare callbacks
	callbacks := a.prepareCallbacks(func(cb Callbacks) interface{} {
		return cb.OnDeleteAll(txn, path)
	})
	defer callbacks.Release()

	// run before callbacks
	for _, c := range callbacks.S {
		if err := c.(DeleteAllCallback).BeforeDeleteAll(objs); err != nil {
			return 0, err
		}
	}

	// delete items
	modTime, deleted, err := a.mod.DeleteAll(txn, path, objs)
	if err != nil {
		return 0, err
	}

	// run after callbacks in reverse order
	for i := len(callbacks.S) - 1; i >= 0; i-- {
		if err := callbacks.S[i].(DeleteAllCallback).AfterDeleteAll(modTime, deleted); err != nil {
			return 0, err
		}
	}

	return modTime, nil
}

func (a *actions) prepareCallbacks(fn func(Callbacks) interface{}) *callbacksSlice {
	slice := poolCallbacksSlice()
	for _, cb := range a.cbs {
		if cm := fn(cb); cm != nil {
			slice.S = append(slice.S, cm)
		}
	}
	return slice
}
