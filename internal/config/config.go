package config

import (
	"time"

	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/identity"
	"github.com/riposo/riposo/pkg/plugin"
	"github.com/riposo/riposo/pkg/riposo"
	"github.com/riposo/riposo/pkg/slowhash"
)

// Config configures the server.
type Config struct {
	Project struct {
		Name    string `default:"riposo"`
		Version string
		Docs    string `default:"https://github.com/riposo/riposo/"`
	}

	ID struct {
		Factory string `default:"nanoid"`
	}

	Storage struct {
		URL string `default:"memory:"`
	}
	Permission struct {
		URL      string `default:"memory:"`
		Defaults map[string][]string
	}
	Cache struct {
		URL string `default:"memory:"`
	}

	Batch struct {
		MaxRequests int `default:"25" yaml:"max_requests"`
	}
	Auth struct {
		Methods []string `default:"basic"`
		Hash    string   `default:"argon2id"`
	}
	CORS struct {
		Origins []string      `default:"*"`
		MaxAge  time.Duration `default:"1h" yaml:"max_age"`
	}
	Pagination struct {
		TokenValidity time.Duration `default:"10m" yaml:"token_validity"`
		MaxLimit      int           `default:"10000" yaml:"max_limit"`
	}

	Backoff struct {
		Duration   time.Duration
		Percentage int
	}
	RetryAfter time.Duration `default:"30s" yaml:"retry_after"`

	Server struct {
		Address         string        `default:":8888"`
		ReadTimeout     time.Duration `default:"60s" yaml:"read_timeout"`
		WriteTimeout    time.Duration `default:"60s" yaml:"write_timeout"`
		ShutdownTimeout time.Duration `default:"5s" yaml:"shutdown_timeout"`
	}

	Temp struct {
		Dir string
	}

	Plugins []string

	// End of Service
	EOS struct {
		Time    time.Time
		Message string
		URL     string
	}

	// Ignored & private fields
	Capabilities *plugin.Set `yaml:"-"`
	parseFunc
}

// Parse parses the config from environment.
func Parse(configFile string, env Env) (*Config, error) {
	parseFunc := newParseFunc(configFile, env)

	var c Config
	c.Project.Version = riposo.Version
	c.Capabilities = new(plugin.Set)

	if err := parseFunc(&c); err != nil {
		return nil, err
	}
	c.parseFunc = parseFunc
	return &c, nil
}

// APIConfig returns an API config.
func (c *Config) APIConfig() *api.Config {
	return &api.Config{
		Guard: c.Permission.Defaults,
		Pagination: (struct {
			TokenValidity time.Duration
			MaxLimit      int
		})(c.Pagination),
	}
}

// InitHelpers inits helpers.
func (c *Config) InitHelpers() (riposo.Helpers, error) {
	if c.parseFunc == nil {
		panic("config parser is not initialised")
	}

	nextID, err := identity.Get(c.ID.Factory)
	if err != nil {
		return nil, err
	}

	slowHash, err := slowhash.Get(c.Auth.Hash)
	if err != nil {
		return nil, err
	}

	return &helpers{
		parseConfig: c.parseFunc,
		nextID:      nextID,
		slowHash:    slowHash,
	}, nil
}
