package cache

// NumEntries is a test helper.
func (b *backend) NumEntries() (int64, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return int64(len(b.keys)), nil
}
