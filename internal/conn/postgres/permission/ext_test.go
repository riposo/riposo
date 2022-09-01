package permission

// NumEntries is a test helper.
func (tx *transaction) NumEntries() (int64, error) {
	var cnt int64

	//nolint:execinquery
	err := tx.
		QueryRowContext(tx.ctx, `
			WITH totals AS (SELECT COUNT(1) AS cnt FROM permission_principals UNION ALL SELECT COUNT(1) AS cnt FROM permission_paths)
			SELECT SUM(cnt) FROM totals
		`).
		Scan(&cnt)
	return cnt, err
}
