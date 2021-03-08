package config

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	plugingo "plugin"
	"strings"

	"github.com/riposo/riposo/pkg/api"
	"github.com/riposo/riposo/pkg/plugin"
	"go.uber.org/multierr"
)

// Plugins is a slice of plugins.
type Plugins []plugin.Plugin

// Close implements the io.Closer interface.
func (pp Plugins) Close() (err error) {
	for _, pin := range pp {
		multierr.Append(err, pin.Close())
	}
	return
}

// LoadPlugin loads a single plugin from a file path.
func LoadPlugin(name string) (plugin.Factory, error) {
	src, err := plugingo.Open(name)
	if err != nil {
		return nil, err
	}

	sym, err := src.Lookup("Plugin")
	if err != nil {
		return nil, err
	}

	pfc, ok := sym.(func(*api.Routes) (plugin.Plugin, error))
	if !ok {
		return nil, fmt.Errorf("plugin: symbol Plugin not a valid type (%s)", name)
	}

	return pfc, nil
}

func safeDownload(url, dir string) error {
	// generate full file name
	name := path.Base(url)
	if i := strings.IndexByte(name, '?'); i > -1 {
		name = name[:i]
	}
	name = filepath.Join(dir, name)

	// check if compressed
	compressed := false
	if strings.HasSuffix(name, ".gz") {
		compressed = true
		name = strings.TrimSuffix(name, ".gz")
	}

	// skip if file exists
	if fileExists(name) {
		return nil
	}

	// create dir
	if err := os.MkdirAll(filepath.Dir(name), 0777); err != nil {
		return err
	}

	// open + download file
	file, err := os.Create(name + ".tmpl")
	if err != nil {
		return err
	}
	defer file.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var src io.Reader = resp.Body
	if compressed {
		crd, err := gzip.NewReader(resp.Body)
		if err != nil {
			return err
		}
		defer crd.Close()

		src = crd
	}

	if _, err := io.Copy(file, src); err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	// atomically rename
	return os.Rename(name+".tmpl", name)
}

func fileExists(name string) bool {
	info, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
