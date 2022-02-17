package cache

// NumEntries is a test helper.
func (t *transaction) NumEntries() (int64, error) {
	return int64(len(t.b.keys)), nil
}
