package config

import "os"

// Env is the source of environment data
type Env interface {
	Get(string) string
}

// MapEnv is a mock environment.
type MapEnv map[string]string

// Get implements the Env interface.
func (e MapEnv) Get(key string) string {
	return e[key]
}

// OSEnv is the default Env.
type OSEnv struct{}

// Get implements the Env interface.
func (e OSEnv) Get(key string) string {
	return os.Getenv(key)
}
