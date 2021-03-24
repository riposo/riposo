package cli

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
	"github.com/riposo/riposo/pkg/plugin"
)

// Plugins inits a new sub-command.
func Plugins() subcommands.Command { return new(pluginsCmd) }

type pluginsCmd struct{}

func (*pluginsCmd) Name() string             { return "plugins" }
func (*pluginsCmd) Synopsis() string         { return "List available plugins." }
func (*pluginsCmd) Usage() string            { return "plugins:\n  List available plugins.\n" }
func (*pluginsCmd) SetFlags(_ *flag.FlagSet) {}

func (c *pluginsCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	plugin.EachMeta(func(id string, meta map[string]interface{}) {
		fprintf(os.Stdout, "* %s", id)
		for key, val := range meta {
			fprintf(os.Stdout, "  %s: %v", key, val)
		}
	})
	return subcommands.ExitSuccess
}
