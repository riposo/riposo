package storage

// NumEntries is a test helper.
func (t *transaction) NumEntries() (int64, error) {
	return int64(t.b.tree.Len()), nil
}
