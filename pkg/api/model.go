package api

import (
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"
)

// Model is a CRUD resource operator.
type Model interface {
	// Get retrieves a single resource.
	Get(txn *Txn, path riposo.Path) (*schema.Resource, error)
	// Create creates a single resource.
	Create(txn *Txn, path riposo.Path, payload *schema.Resource) error
	// Update updates a resource.
	Update(txn *Txn, path riposo.Path, exst *schema.Object, payload *schema.Resource) error
	// Patch patches a resource.
	Patch(txn *Txn, path riposo.Path, exst *schema.Object, payload *schema.Resource) error
	// Delete deletes a resource.
	Delete(txn *Txn, path riposo.Path, exst *schema.Object) (*schema.Object, error)
	// DeleteAll deletes resources.
	DeleteAll(txn *Txn, path riposo.Path, objs []*schema.Object) (riposo.Epoch, []riposo.Path, error)
}

// DefaultModel is an embeddable default model type.
type DefaultModel struct{}

func (DefaultModel) Get(txn *Txn, path riposo.Path) (*schema.Resource, error) {
	// find the object
	obj, err := txn.Store.Get(path, false)
	if err != nil {
		return nil, err
	}

	// get permissions
	pms, err := txn.Perms.GetPermissions(path)
	if err != nil {
		return nil, err
	}

	return &schema.Resource{Data: obj, Permissions: pms}, nil
}

func (DefaultModel) Create(txn *Txn, path riposo.Path, payload *schema.Resource) error {
	// create new object
	err := txn.Store.Create(path, payload.Data)
	if err != nil {
		return err
	}

	// ensure permissions are not nil
	if payload.Permissions == nil {
		payload.Permissions = make(schema.PermissionSet, 1)
	}
	// include current user as writer
	if user := txn.User; user != nil && user.ID != riposo.Everyone {
		payload.Permissions.Add("write", user.ID)
	}
	// create permissions using ID
	return txn.Perms.CreatePermissions(path.WithObjectID(payload.Data.ID), payload.Permissions)
}

func (DefaultModel) Update(txn *Txn, path riposo.Path, exst *schema.Object, payload *schema.Resource) error {
	// update existing object with received data
	exst.Update(payload.Data)
	return update(txn, path, exst, payload.Permissions)
}

func (DefaultModel) Patch(txn *Txn, path riposo.Path, exst *schema.Object, payload *schema.Resource) error {
	// patch existing object with received data
	if err := exst.Patch(payload.Data); err != nil {
		return err
	}
	return update(txn, path, exst, payload.Permissions)
}

func (DefaultModel) Delete(txn *Txn, path riposo.Path, _ *schema.Object) (*schema.Object, error) {
	// delete permissions
	if err := txn.Perms.DeletePermissions([]riposo.Path{path}); err != nil {
		return nil, err
	}

	// delete object
	return txn.Store.Delete(path)
}

func (DefaultModel) DeleteAll(txn *Txn, path riposo.Path, objs []*schema.Object) (riposo.Epoch, []riposo.Path, error) {
	// collect paths
	paths := make([]riposo.Path, 0, len(objs))
	for _, obj := range objs {
		paths = append(paths, path.WithObjectID(obj.ID))
	}

	// delete permissions
	if err := txn.Perms.DeletePermissions(paths); err != nil {
		return 0, nil, err
	}

	// delete objects
	return txn.Store.DeleteAll(paths)
}

func update(txn *Txn, path riposo.Path, exst *schema.Object, ps schema.PermissionSet) error {
	// update object
	if err := txn.Store.Update(path, exst); err != nil {
		return err
	}

	// update permissions
	if ps != nil {
		// include current user as writer
		if user := txn.User; user != nil && user.ID != riposo.Everyone {
			ps.Add("write", user.ID)
		}
		if err := txn.Perms.MergePermissions(path, ps); err != nil {
			return err
		}
	}

	return nil
}
