package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/riposo/riposo/pkg/api"
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

	Plugin struct {
		Dir  []string `default:"./plugins"`
		URLs []string
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

	Capabilities map[string]map[string]interface{} `ignored:"true"`
}

// Parse parses the config from environment.
func Parse() (*Config, error) {
	var c Config
	c.Project.Version = riposo.Version
	c.Capabilities = make(map[string]map[string]interface{})

	if err := riposo.ParseEnv(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

// FetchPlugins downloads plugins.
func (c *Config) FetchPlugins() error {
	if len(c.Plugin.URLs) == 0 {
		return nil
	}

	tempDir := c.Temp.Dir
	if tempDir == "" {
		tempDir = os.TempDir()
	}
	pluginDir := filepath.Join(tempDir, "riposo-plugins")
	for _, url := range c.Plugin.URLs {
		riposo.Logger.Println("fetching plugin", url)
		if err := safeDownload(url, pluginDir); err != nil {
			return err
		}
	}

	c.Plugin.Dir = append(c.Plugin.Dir, pluginDir)
	return nil
}

// LoadPlugins includes plugins.
func (c *Config) LoadPlugins(rts *api.Routes) (Plugins, error) {
	filesSeen := make(map[string]struct{})
	var plugins Plugins

	for _, dir := range c.Plugin.Dir {
		files, err := filepath.Glob(filepath.Join(dir, "*.so"))
		if err != nil {
			_ = plugins.Close()
			return nil, err
		}
		for _, fname := range files {
			if _, ok := filesSeen[fname]; ok {
				continue
			}
			filesSeen[fname] = struct{}{}

			pft, err := LoadPlugin(fname)
			if err != nil {
				_ = plugins.Close()
				return nil, err
			}

			pin, err := pft(rts)
			if err != nil {
				_ = plugins.Close()
				return nil, err
			}

			plugins = append(plugins, pin)
			if _, ok := c.Capabilities[pin.ID()]; ok {
				_ = plugins.Close()
				return nil, fmt.Errorf("plugin: %q is already registered (%s)", pin.ID(), fname)
			}
			c.Capabilities[pin.ID()] = pin.Meta()

			riposo.Logger.Printf("loaded plugin %q (%s)", pin.ID(), fname)
		}
	}
	return plugins, nil
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
