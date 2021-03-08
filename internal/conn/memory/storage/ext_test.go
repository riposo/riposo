package storage

// NumEntries is a test helper.
func (b *backend) NumEntries() (int64, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return int64(b.tree.Len()), nil
}
