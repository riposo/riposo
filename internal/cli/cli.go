package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/subcommands"
)

type usageError string

func usageErrorf(s string, vv ...interface{}) usageError {
	return usageError(fmt.Sprintf(s, vv...))
}

func (s usageError) Error() string { return string(s) }

// --------------------------------------------------------------------

func statusOf(err error) subcommands.ExitStatus {
	if err == nil {
		return subcommands.ExitSuccess
	}

	var usage usageError
	if errors.As(err, &usage) {
		stderr("[!] %s", usage)
		return subcommands.ExitUsageError
	}

	stderr("[!] exited with %v", err)
	return subcommands.ExitFailure
}

func stdout(s string, vv ...interface{}) {
	fprintf(os.Stdout, s, vv...)
}

func stderr(s string, vv ...interface{}) {
	fprintf(os.Stderr, s, vv...)
}

func fprintf(w io.Writer, s string, vv ...interface{}) {
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	fmt.Fprintf(w, s, vv...)
}
