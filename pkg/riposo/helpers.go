package riposo

// Helers are common routines that are used throughout the app. They are exposed
// to plugins for convenience and consistency.
type Helpers interface {
	// SlowHash applies a password-hash to a plain string
	// returning a cryptographically secure, hashed string.
	SlowHash(plain string) (string, error)
	// NextID returns a new globally unique ID.
	NextID() string
	// ParseConfig parses configuration into a target struct using RIPOSO_* env
	// variables and an optional YAML file provided on boot.
	ParseConfig(v interface{}) error
}
