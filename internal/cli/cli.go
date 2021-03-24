package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/subcommands"
)

func exitStatus(err error) subcommands.ExitStatus {
	if err == nil {
		return subcommands.ExitSuccess
	}

	failure(err.Error())
	return subcommands.ExitFailure
}

func failure(s string) {
	fprintf(os.Stderr, "[!] exited with %s", s)
}

func fprintf(w io.Writer, s string, vv ...interface{}) {
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	fmt.Fprintf(w, s, vv...)
}
