package storage

import (
	"github.com/riposo/riposo/pkg/riposo"
)

// ReloadHelpers is a test helper.
func (cn *conn) ReloadHelpers(hlp riposo.Helpers) {
	cn.hlp = hlp
}

// NumEntries is a test helper.
func (tx *transaction) NumEntries() (int64, error) {
	var cnt int64
	err := tx.
		QueryRowContext(tx.ctx, `SELECT COUNT(1) FROM storage_objects WHERE NOT deleted`).
		Scan(&cnt)
	return cnt, err
}
