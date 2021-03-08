package riposo

import (
	"fmt"
	"time"
)

const httpFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

// Epoch is the number of milliseconds since 1970.
type Epoch int64

// CurrentEpoch returns the current epoch.
func CurrentEpoch() Epoch {
	return EpochFromTime(time.Now())
}

// EpochFromTime creates an epoch from time.
func EpochFromTime(t time.Time) Epoch {
	t = t.Round(time.Millisecond)
	return Epoch(t.Unix()*1000 + int64(t.Nanosecond()/1e6))
}

// IsZero returns true if  zero.
func (e Epoch) IsZero() bool {
	return e < 1
}

// Time converts epoch to a time.
func (e Epoch) Time() time.Time {
	n := int64(e)
	return time.Unix(n/1000, (n%1000)*1e6)
}

// ETag returns a http.Header compatible ETag value.
func (e Epoch) ETag() string {
	return fmt.Sprintf(`"%d"`, e)
}

// HTTPFormat returns a http.Header compatible time string.
func (e Epoch) HTTPFormat() string {
	return e.Time().UTC().Format(httpFormat)
}
