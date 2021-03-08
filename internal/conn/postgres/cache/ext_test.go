package cache

// NumEntries is a test helper.
func (tx *transaction) NumEntries() (int64, error) {
	var cnt int64
	err := tx.
		QueryRowContext(tx.ctx, `SELECT COUNT(1) FROM cache_keys`).
		Scan(&cnt)
	return cnt, err
}
