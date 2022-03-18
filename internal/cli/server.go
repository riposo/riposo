package cli

import (
	"context"
	"flag"
	"os/signal"
	"syscall"

	"github.com/google/subcommands"
	"github.com/riposo/riposo/internal/config"
	"github.com/riposo/riposo/internal/server"
)

// Server inits a new sub-command.
func Server() subcommands.Command { return new(serverCmd) }

type serverCmd struct {
	configFile string
}

func (*serverCmd) Name() string     { return "server" }
func (*serverCmd) Synopsis() string { return "Start HTTP server." }
func (*serverCmd) Usage() string    { return "server:\n  Start HTTP server.\n" }
func (c *serverCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.configFile, "config", "", "Optional YAML config file")
}

func (c *serverCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	cfg, err := config.Parse(c.configFile, nil)
	if err != nil {
		failure("invalid configuration: " + err.Error())
		return subcommands.ExitUsageError
	}

	return exitStatus(c.run(ctx, cfg))
}

func (c *serverCmd) run(ctx context.Context, cfg *config.Config) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv, err := server.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer srv.Close()

	return srv.ListenAndServe()
}
