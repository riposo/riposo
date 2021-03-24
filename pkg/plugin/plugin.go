package plugin

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/riposo/riposo/pkg/api"
	"go.uber.org/multierr"
)

// Factory initializes a new plugin at runtime.
type Factory func(*api.Routes) (Plugin, error)

// A Plugin must export human-readable definitions.
type Plugin interface {
	// Meta exposes plugin metadata as key-value pairs.
	// Commonly used keys are: "url" and "description".
	Meta() map[string]interface{}

	// Close is called on server shutdown.
	io.Closer
}

// Set contains the set of loaded plugins.
type Set struct {
	closers []io.Closer
	meta    map[string]map[string]interface{}
}

func newSet(rts *api.Routes, factories map[string]Factory) (*Set, error) {
	set := &Set{
		closers: make([]io.Closer, 0, len(factories)),
		meta:    make(map[string]map[string]interface{}, len(factories)),
	}

	for name, factory := range factories {
		pin, err := factory(rts)
		if err != nil {
			_ = set.Close()
			return nil, err
		}

		set.closers = append(set.closers, pin)
		set.meta[name] = pin.Meta()
	}
	return set, nil
}

// Close closes all plugins.
func (s *Set) Close() error {
	var err error
	for _, c := range s.closers {
		err = multierr.Append(err, c.Close())
	}
	return err
}

// MarshalJSON encodes the set as JSON.
func (s *Set) MarshalJSON() ([]byte, error) {
	if s.meta == nil {
		return []byte{'{', '}'}, nil
	}
	return json.Marshal(s.meta)
}

// --------------------------------------------------------------------

var (
	registry   = make(map[string]Factory)
	registryMu sync.RWMutex
)

// Register registers a new factory by name.
// It will panic if multiple factories are registered under the same name.
// Names must be lowercase and only contain alphanumeric characters and underscores.
func Register(name string, factory Factory) {
	registryMu.Lock()
	defer registryMu.Unlock()

	_, ok := registry[name]
	if ok {
		panic("factory " + name + " is already registered")
	}
	registry[name] = factory
}

// Init initializes registered plugins.
func Init(rts *api.Routes) (*Set, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	return newSet(rts, registry)
}
