package cli

import (
	"context"
	"flag"
	"fmt"

	"github.com/google/subcommands"
	"github.com/riposo/riposo/internal/config"
	"github.com/riposo/riposo/pkg/api"
)

// PluginCheck checks is a plugin is compatible.
func PluginCheck() subcommands.Command { return new(pluginCheck) }

type pluginCheck struct{}

func (*pluginCheck) Name() string             { return "plugin:check" }
func (*pluginCheck) Synopsis() string         { return "Check plugin compatibility." }
func (*pluginCheck) Usage() string            { return "plugin:check PATH\n" }
func (*pluginCheck) SetFlags(_ *flag.FlagSet) {}

func (c *pluginCheck) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return statusOf(c.run(ctx, f))
}

func (c *pluginCheck) run(ctx context.Context, f *flag.FlagSet) error {
	name := f.Arg(0)
	if name == "" {
		return usageError("file path argument required")
	}

	pft, err := config.LoadPlugin(name)
	if err != nil {
		return err
	}

	pin, err := pft(api.NewRoutes(nil))
	if err != nil {
		return err
	}
	defer pin.Close()

	meta := pin.Meta()
	max := 5
	for key := range meta {
		if len(key) > max {
			max = len(key)
		}
	}

	pad := fmt.Sprintf("%%-%ds", max)
	stdout(pad+" : %s", "status:", "OK")
	stdout(pad+" : %s", "plugin:", pin.ID())
	for key, value := range meta {
		stdout(pad+" : %s", key, value)
	}
	return nil
}
