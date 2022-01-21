package api

import (
	"github.com/riposo/riposo/pkg/conn/permission"
)

// Authz represents an authorization guard and can verify access of user
// principals to target entities.
type Authz map[string][]string

// Verify returns true if any of the user principals can access any of the
// target entities.
func (v Authz) Verify(txn permission.Transaction, principals []string, target []permission.ACE) (bool, error) {
	if len(principals) == 0 {
		return false, nil
	}

	// check static principals first
	for _, ent := range target {
		if allowed, ok := v[ent.Perm]; ok {
			if containsOneOf(principals, allowed) {
				return true, nil
			}
		}
	}

	// retrieve all stored principals that can access target
	allowed, err := txn.GetAllACEPrincipals(target)
	if err != nil {
		return false, err
	}
	return containsOneOf(principals, allowed), nil
}

func containsOneOf(vv, ww []string) bool {
	if len(ww) == 0 {
		return false
	}

	for _, v := range vv {
		for _, w := range ww {
			if v == w {
				return true
			}
		}
	}
	return false
}
