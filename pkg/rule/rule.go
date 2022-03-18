package rule

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/riposo"
	"go.uber.org/multierr"
	"gopkg.in/yaml.v3"
)

// A Set is a collection of rules grouped by type.
type Set interface {
	api.Component
}

// Factory initializes the rule Set.
type Factory func([]*yaml.Node) (Set, error)

// ----------------------------------------------------------------------------

var (
	registry   = make(map[string]Factory)
	registryMu sync.RWMutex
)

// Register registers a new factory by name.
// It will panic if multiple factories are registered under the same name.
func Register(name string, factory Factory) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if _, ok := registry[name]; ok {
		panic("factory " + name + " is already registered")
	}
	registry[name] = factory
}

// ----------------------------------------------------------------------------

type config struct {
	Type string `yaml:"type"`
}

type closer []Set

// Init initializes registered rules.
func Init(ctx context.Context, rts *api.Routes, hlp riposo.Helpers, nodes []*yaml.Node) (io.Closer, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	typed := make(map[string][]*yaml.Node)
	for _, node := range nodes {
		var c config
		if err := node.Decode(&c); err != nil {
			return nil, err
		}

		// ensure type is set
		if c.Type == "" {
			return nil, fmt.Errorf("rule configuration %v is missing type", node)
		}

		// check if available
		_, ok := registry[c.Type]
		if !ok {
			return nil, fmt.Errorf("rule %q is not available", c.Type)
		}

		// store config
		typed[c.Type] = append(typed[c.Type], node)
	}

	cc := make(closer, 0, len(typed))
	for typ, nodes := range typed {
		// get factory
		fct := registry[typ]

		// enable plugin
		set, err := fct(nodes)
		if err != nil {
			_ = cc.Close()
			return nil, fmt.Errorf("rule %q init failed with: %w", typ, err)
		}

		// store set
		cc = append(cc, set)
	}

	return cc, nil
}

// Close closes all plugins.
func (cc closer) Close() (err error) {
	for _, set := range cc {
		err = multierr.Append(err, set.Close())
	}
	return err
}
