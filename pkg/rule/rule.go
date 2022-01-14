package rule

import (
	"context"
	"fmt"
	"sync"

	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/riposo"
	"go.uber.org/multierr"
	"gopkg.in/yaml.v3"
)

// A Rule interface.
type Rule interface {
	api.Component
}

// Factory initializes the Rule.
type Factory func(*yaml.Node) (Rule, error)

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

	_, ok := registry[name]
	if ok {
		panic("factory " + name + " is already registered")
	}
	registry[name] = factory
}

// ----------------------------------------------------------------------------

type config struct {
	Type string `yaml:"type"`
}

// Set contains the set of loaded rules.
type Set struct {
	rules []Rule
}

// Init initializes registered rules.
func Init(ctx context.Context, rts *api.Routes, hlp riposo.Helpers, nodes []*yaml.Node) (*Set, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	set := &Set{
		rules: make([]Rule, 0, len(nodes)),
	}

	for _, node := range nodes {
		var c config
		if err := node.Decode(&c); err != nil {
			_ = set.Close()
			return nil, err
		}

		// ensure type is set
		if c.Type == "" {
			_ = set.Close()
			return nil, fmt.Errorf("rule configuration %v is missing type", node)
		}

		// check if available
		fct, ok := registry[c.Type]
		if !ok {
			_ = set.Close()
			return nil, fmt.Errorf("rule %q is not available", c.Type)
		}

		// enable plugin
		rule, err := fct(node)
		if err != nil {
			_ = set.Close()
			return nil, fmt.Errorf("rule %q init failed with: %w", c.Type, err)
		}

		// store rule
		set.rules = append(set.rules, rule)
	}

	return set, nil
}

// Close closes all plugins.
func (s *Set) Close() error {
	var err error
	for _, rule := range s.rules {
		err = multierr.Append(err, rule.Close())
	}
	return err
}
