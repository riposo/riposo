package config

import (
	"os"
	"strings"
	"time"

	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/plugin"
	"github.com/riposo/riposo/pkg/riposo"
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
		URL string `default:"memory:"`
	}
	Cache struct {
		URL string `default:"memory:"`
	}

	Batch struct {
		MaxRequests int `default:"25" split_words:"true"`
	}
	Auth struct {
		Methods []string `default:"basic"`
		Hash    string   `default:"argon2id"`
	}
	CORS struct {
		Origins []string      `default:"*"`
		MaxAge  time.Duration `default:"1h" split_words:"true"`
	}
	Pagination struct {
		TokenValidity time.Duration `default:"10m"`
		MaxLimit      int           `default:"10000" split_words:"true"`
	}

	Server struct {
		Addr            string        `default:":8888"`
		ReadTimeout     time.Duration `default:"60s" split_words:"true"`
		WriteTimeout    time.Duration `default:"60s" split_words:"true"`
		ShutdownTimeout time.Duration `default:"5s" split_words:"true"`
	}

	Temp struct {
		Dir string
	}

	// End of Service
	EOS struct {
		Time    time.Time
		Message string
		URL     string
	}

	Plugins      []string
	Capabilities *plugin.Set `ignored:"true"`
}

// Parse parses the config from environment.
func Parse() (*Config, error) {
	var c Config
	c.Project.Version = riposo.Version
	c.Capabilities = new(plugin.Set)

	if err := riposo.ParseEnv(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

// APIConfig returns an API config.
func (c *Config) APIConfig() *api.Config {
	return &api.Config{
		Guard: c.parseGuard(),
		Pagination: (struct {
			TokenValidity time.Duration
			MaxLimit      int
		})(c.Pagination),
	}
}

// InitHelpers inits rts helpers.
func (c *Config) InitHelpers() (*riposo.Helpers, error) {
	return riposo.NewHelpers(riposo.HelpersOptions{
		Identity: c.ID.Factory,
		SlowHash: c.Auth.Hash,
	})
}

func (c *Config) parseGuard() api.Guard {
	static := make(api.Guard)
	for _, raw := range os.Environ() {
		if !strings.HasPrefix(raw, "RIPOSO_") {
			continue
		}

		pair := strings.SplitN(raw, "=", 2)
		if len(pair) != 2 || pair[1] == "" || !strings.HasSuffix(pair[0], "_PRINCIPALS") {
			continue
		}

		key := strings.TrimSuffix(strings.TrimPrefix(pair[0], "RIPOSO_"), "_PRINCIPALS")
		key = strings.ReplaceAll(key, "_", ":")
		key = strings.ToLower(key)
		static[key] = strings.Split(pair[1], ",")
	}
	return static
}
