package cli

import (
	"context"
	"flag"

	"github.com/bsm/shutdown"
	"github.com/google/subcommands"
	"github.com/riposo/riposo/internal/config"
	"github.com/riposo/riposo/internal/server"
)

// Server inits a new sub-command.
func Server() subcommands.Command { return new(serverCmd) }

type serverCmd struct{}

func (*serverCmd) Name() string             { return "server" }
func (*serverCmd) Synopsis() string         { return "Start HTTP server." }
func (*serverCmd) Usage() string            { return "server:\n  Start HTTP server.\n" }
func (*serverCmd) SetFlags(_ *flag.FlagSet) {}

func (c *serverCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return statusOf(c.run(ctx))
}

func (c *serverCmd) run(ctx context.Context) error {
	cfg, err := config.Parse()
	if err != nil {
		return usageErrorf("invalid configuration: %v", err)
	}

	term := shutdown.WithContext(ctx)
	srv, err := server.New(term, cfg)
	if err != nil {
		return err
	}
	defer srv.Close()

	return term.WaitFor(srv.ListenAndServe)
}
