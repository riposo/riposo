package riposo

import (
	"github.com/google/subcommands"
	"github.com/riposo/riposo/internal/cli"

	_ "github.com/riposo/riposo/internal/conn/memory"
	_ "github.com/riposo/riposo/internal/conn/postgres"
)

func init() {
	subcommands.Register(cli.Server(), "server")
	subcommands.Register(cli.Plugins(), "plugins")
}
