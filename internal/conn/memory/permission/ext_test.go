package permission

// NumEntries is a test helper.
func (b *backend) NumEntries() (int64, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	n := len(b.users)
	for _, perms := range b.perms {
		n += len(perms)
	}
	return int64(n), nil
}
