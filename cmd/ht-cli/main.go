// Command ht-cli is the official HostTracker command-line client.
//
// Install it with Homebrew (brew install HostTracker/tap/ht-cli), with
// `go install github.com/HostTracker/cli/cmd/ht-cli@latest`, with the
// install script, or from a release archive. `ht-cli --help` lists the
// command groups; the README covers authentication, output formats and
// exit codes.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/HostTracker/cli/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(cmd.Main(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
