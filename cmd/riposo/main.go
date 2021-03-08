package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
	"github.com/riposo/riposo/internal/cli"

	_ "github.com/riposo/riposo/internal/conn/memory"
	_ "github.com/riposo/riposo/internal/conn/postgres"
)

func init() {
	subcommands.Register(subcommands.HelpCommand(), "general help")
	subcommands.Register(subcommands.FlagsCommand(), "general help")
	subcommands.Register(subcommands.CommandsCommand(), "general help")
	subcommands.Register(cli.Server(), "server")
	subcommands.Register(cli.PluginCheck(), "plugin")
}

func main() {
	flag.Parse()

	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
