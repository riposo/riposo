package identity

import (
	"fmt"
	"sync"

	"github.com/bsm/nanoid"
	"github.com/google/uuid"
)

func init() {
	Register("nanoid", NanoID)
	Register("uuid", UUID)
}

var (
	registry   = make(map[string]Factory)
	registryMu sync.RWMutex
)

// Factory is an interface to various ID factories.
type Factory func() string

// Register registers a new factory by name.
// It will panic if multiple factories are registered under the same name.
func Register(name string, fct Factory) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if _, ok := registry[name]; ok {
		panic("factory " + name + " is already registered")
	}
	registry[name] = fct
}

// Get returns a Factory by name.
func Get(name string) (Factory, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	if fct, ok := registry[name]; ok {
		return fct, nil
	}
	return nil, fmt.Errorf("unknown ID factory %q", name)
}

// NanoID implements Factory function.
func NanoID() string {
	return nanoid.Base58.MustGenerate(20)
}

// UUID implements Factory function.
func UUID() string {
	return uuid.New().String()
}

// IsValid validates an ID.
func IsValid(id string) bool {
	sz := len(id)
	if sz < 1 || sz > 255 {
		return false
	}

	for _, c := range id {
		switch c {
		case '!', '#', '$', '%', '&', '(', ')', '+', '-', '.', '@', '[', ']', '^', '_', '{', '|', '}', '~':
			// OK
		default:
			if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') {
				return false
			}
		}
	}
	return true
}
