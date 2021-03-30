package group

import (
	"sort"

	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/conn/storage"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/schema"
)

type modGroup struct{ api.Model }

// Model inits a new group model.
func Model() api.Model {
	return modGroup{Model: api.StdModel()}
}

// Create overrides.
func (m modGroup) Create(txn *api.Txn, path riposo.Path, payload *schema.Resource) error {
	// normalize payload
	extra, err := normGroup(payload.Data, true)
	if err != nil {
		return err
	}

	// perform action
	if err := m.Model.Create(txn, path, payload); err != nil {
		return err
	}

	// add principal to members
	principal := path.WithObjectID(payload.Data.ID).String()
	if err := addPrincipal(txn, principal, extra.Members); err != nil {
		return err
	}

	return nil
}

// Update overrides.
func (m modGroup) Update(txn *api.Txn, path riposo.Path, hs storage.UpdateHandle, payload *schema.Resource) error {
	// normalize payload
	extra, err := normGroup(payload.Data, true)
	if err != nil {
		return err
	}

	// purge principal
	principal := path.String()
	if err := purgePrincipals(txn, principal); err != nil {
		return err
	}

	// perform action
	if err := m.Model.Update(txn, path, hs, payload); err != nil {
		return err
	}

	// add principal to members
	if err := addPrincipal(txn, principal, extra.Members); err != nil {
		return err
	}

	return nil
}

// Patch overrides.
func (m modGroup) Patch(txn *api.Txn, path riposo.Path, hs storage.UpdateHandle, payload *schema.Resource) error {
	// normalize payload
	_, err := normGroup(payload.Data, false)
	if err != nil {
		return err
	}

	// purge principal
	principal := path.String()
	if err := purgePrincipals(txn, principal); err != nil {
		return err
	}

	// perform action
	if err := m.Model.Patch(txn, path, hs, payload); err != nil {
		return err
	}

	// parse merged result
	extra, err := parseGroup(hs.Object())
	if err != nil {
		return err
	}

	// add principal to members
	if err := addPrincipal(txn, principal, extra.Members); err != nil {
		return err
	}

	return nil
}

// Delete overrides.
func (m modGroup) Delete(txn *api.Txn, path riposo.Path) (*schema.Object, error) {
	principal := path.String()

	// purge principal
	if err := purgePrincipals(txn, principal); err != nil {
		return nil, err
	}

	// perform action
	return m.Model.Delete(txn, path)
}

// DeleteAll deletes resources in bulk.
func (m modGroup) DeleteAll(txn *api.Txn, path riposo.Path, objIDs ...string) (riposo.Epoch, error) {
	if len(objIDs) != 0 {
		// purge principals
		principals := make([]string, 0, len(objIDs))
		for _, objID := range objIDs {
			principals = append(principals, path.WithObjectID(objID).String())
		}
		if err := purgePrincipals(txn, principals...); err != nil {
			return 0, err
		}
	}

	// perform action
	return m.Model.DeleteAll(txn, path, objIDs...)
}

func addPrincipal(txn *api.Txn, principal string, userIDs []string) error {
	return txn.Perms.AddUserPrincipal(principal, userIDs)
}

func purgePrincipals(txn *api.Txn, principals ...string) error {
	return txn.Perms.PurgeUserPrincipals(principals...)
}

// --------------------------------------------------------------------

// Extra is the payload object.
type extra struct {
	Members []string `json:"members"`
}

func (p *extra) norm() {
	sort.Strings(p.Members)

	m := p.Members[:0]
	for _, v := range p.Members {
		if v != "" {
			if n := len(m); n == 0 || m[n-1] != v {
				m = append(m, v)
			}
		}
	}
	p.Members = m
}

func parseGroup(obj *schema.Object) (*extra, error) {
	var p *extra
	if err := obj.DecodeExtra(&p); err != nil {
		return nil, schema.BadRequest(err)
	}
	return p, nil
}

func normGroup(obj *schema.Object, provision bool) (*extra, error) {
	// parse
	p, err := parseGroup(obj)
	if err != nil {
		return nil, err
	}

	// validate
	if provision && p.Members == nil {
		p.Members = []string{}
	}

	// norm
	p.norm()
	if err := obj.EncodeExtra(p); err != nil {
		return nil, err
	}

	return p, nil
}
