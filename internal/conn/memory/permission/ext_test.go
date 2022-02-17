package permission

// NumEntries is a test helper.
func (t *transaction) NumEntries() (int64, error) {
	n := len(t.b.users)
	for _, perms := range t.b.perms {
		n += len(perms)
	}
	return int64(n), nil
}
