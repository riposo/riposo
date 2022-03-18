package riposo

import (
	"github.com/google/subcommands"
	"github.com/riposo/riposo/internal/cli"

	_ "github.com/riposo/riposo/internal/conn/memory"   // include memory storage support by default
	_ "github.com/riposo/riposo/internal/conn/postgres" // include postgres storage support by default
)

func init() {
	subcommands.Register(cli.Server(), "server")
	subcommands.Register(cli.Plugins(), "plugins")
}
