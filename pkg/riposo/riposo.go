// Package riposo contains the basic necessities.
package riposo

import "github.com/kelseyhightower/envconfig"

const (
	// Version references the code version.
	Version = "0.1.0"

	// APIVersion references the implemented API version.
	APIVersion = "1.22"
)

// Static principals.
const (
	Everyone      = "system.Everyone"
	Authenticated = "system.Authenticated"
)

// ParseEnv uses github.com/kelseyhightower/envconfig to process
// RIPOSO_* env variables and parse the values into target config.
func ParseEnv(config interface{}) error {
	return envconfig.Process("riposo", config)
}
