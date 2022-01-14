package plugin

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/riposo"
	"go.uber.org/multierr"
)

// A Plugin interface.
type Plugin interface {
	// IDs must be lowercase and only contain alphanumeric characters and hyphens.
	ID() string

	// Meta exposes plugin metadata as key-value pairs.
	// Commonly used keys are: "url" and "description".
	Meta() map[string]interface{}

	// Init callback is called on init.
	Init(*api.Routes, riposo.Helpers) error

	// Close callback is called on shutdown.
	Close() error
}

type simple struct {
	id    string
	meta  map[string]interface{}
	init  func(*api.Routes, riposo.Helpers) error
	close func() error
}

// New inits a simple plugin.
func New(id string, meta map[string]interface{}, init func(*api.Routes, riposo.Helpers) error, close func() error) Plugin {
	return &simple{
		id:    id,
		meta:  meta,
		init:  init,
		close: close,
	}
}

func (s *simple) ID() string                   { return s.id }
func (s *simple) Meta() map[string]interface{} { return s.meta }
func (s *simple) Init(rts *api.Routes, hlp riposo.Helpers) error {
	if s.init != nil {
		return s.init(rts, hlp)
	}
	return nil
}
func (s *simple) Close() error {
	if s.close != nil {
		return s.close()
	}
	return nil
}

// --------------------------------------------------------------------

// Set contains the set of loaded plugins.
type Set struct {
	pins []Plugin
	meta map[string]map[string]interface{}
}

// Close closes all plugins.
func (s *Set) Close() error {
	var err error
	for _, pin := range s.pins {
		err = multierr.Append(err, pin.Close())
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
	registry   = make(map[string]Plugin)
	registryMu sync.RWMutex
)

// Register registers a new plugin.
// It will panic if multiple plugin are registered under the same ID.
func Register(pin Plugin) {
	registryMu.Lock()
	defer registryMu.Unlock()

	id := pin.ID()
	if id == "" {
		panic("plugin ID cannot be blank")
	}

	for _, c := range id {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-' {
			panic("plugin ID '" + id + "' is invalid")
		}
	}

	if _, ok := registry[id]; ok {
		panic("plugin " + id + " is already registered")
	}
	registry[id] = pin
}

// EachMeta iterates over each ID and meta.
func EachMeta(iter func(id string, meta map[string]interface{})) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	ids := make([]string, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		pin := registry[id]
		iter(id, pin.Meta())
	}
}

// Init initializes registered plugins.
func Init(rts *api.Routes, hlp riposo.Helpers, enabled []string) (*Set, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	set := &Set{
		pins: make([]Plugin, 0, len(enabled)),
		meta: make(map[string]map[string]interface{}, len(enabled)),
	}

	for _, id := range enabled {
		// skip if already enabled
		if _, ok := set.meta[id]; ok {
			continue
		}

		// check if available
		pin, ok := registry[id]
		if !ok {
			_ = set.Close()
			return nil, fmt.Errorf("plugin %q is not available", id)
		}

		// enable plugin
		set.pins = append(set.pins, pin)
		set.meta[id] = pin.Meta()

		// init plugin
		if err := pin.Init(rts, hlp); err != nil {
			_ = set.Close()
			return nil, err
		}
	}

	return set, nil
}
