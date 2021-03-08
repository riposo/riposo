package testdata

import "time"

// MockNow returns a std mock time.
func MockNow() time.Time {
	return time.Unix(1567815678, 987987987)
}
