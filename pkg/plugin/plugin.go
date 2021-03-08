package plugin

import "github.com/riposo/riposo/pkg/api"

// Factory initializes a new plugin at runtime.
type Factory func(*api.Routes) (Plugin, error)

// A Plugin must export human-readable definitions.
type Plugin interface {
	// ID returns the globally unique plugin identifier.
	// IDs must be lowercase and only contain alphanumeric characters and hyphens.
	ID() string

	// Meta exposes plugin metadata as key-value pairs.
	// Commonly used keys are: "url" and "description".
	Meta() map[string]interface{}

	// Close is called on server shutdown.
	Close() error
}

type plugin struct {
	id    string
	meta  map[string]interface{}
	close func() error
}

// New inits a new plugin.
func New(id string, meta map[string]interface{}, close func() error) Plugin {
	return &plugin{
		id:    id,
		meta:  meta,
		close: close,
	}
}

func (p *plugin) ID() string {
	return p.id
}

func (p *plugin) Meta() map[string]interface{} {
	return p.meta
}

func (p *plugin) Close() error {
	if p.close != nil {
		return p.close()
	}
	return nil
}
