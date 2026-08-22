// Package cmd assembles the `ht-cli` command tree: the generated command per
// API operation, and the hand-written commands around them.
package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	_ "github.com/HostTracker/cli/cmd/gen" // registers one command per operation
	"github.com/HostTracker/cli/internal/htcli"
)

const rootLong = `ht-cli is the official command-line client for the
HostTracker API.

Every operation of the API is a command, grouped by the family it belongs
to: ht-cli monitors list, ht-cli contacts create, ht-cli status-pages
get. Reads print a table on a terminal and JSON when piped; --output
picks explicitly.

Start with:

  ht-cli auth login                      store an API token
  ht-cli monitors list                   the account's monitors
  ht-cli check run https://example.com   a one-off check, no monitor needed

Tokens are minted on the HostTracker profile page.`

// New builds the root command. opts carries the global flags and is
// resolved once the command line has been parsed.
func New(opts *htcli.Options) *cobra.Command {
	root := &cobra.Command{
		Use:           "ht-cli",
		Short:         "The HostTracker command-line client",
		Long:          rootLong,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			opts.Started = true
			opts.Streams()
			return nil
		},
	}
	root.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return &htcli.UsageError{Err: err}
	})
	opts.Bind(root.PersistentFlags())

	htcli.AddGroups(root, opts)
	root.AddCommand(
		newAuthCommand(opts),
		newConfigCommand(opts),
		newCheckCommand(opts),
		newAPICommand(opts),
		newVersionCommand(opts),
		newCompletionCommand(),
		newDocsCommand(),
	)
	addJobsWait(root, opts)
	addWebhooksVerify(root, opts)

	root.SetIn(opts.In)
	root.SetOut(opts.Out)
	root.SetErr(opts.Err)
	return root
}

// Main runs the CLI and returns the process exit code.
func Main(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	opts := &htcli.Options{Out: stdout, Err: stderr, In: stdin, UserAgent: userAgent()}
	opts.Streams()

	root := New(opts)
	root.SetArgs(args)

	err := root.ExecuteContext(ctx)
	if err == nil {
		return htcli.ExitOK
	}
	// A failure raised before any command body ran came from the command
	// line itself, whatever cobra called it.
	if !opts.Started {
		err = &htcli.UsageError{Err: err}
	}
	code := htcli.ExitCode(err)
	htcli.PrintError(opts.Err, err, opts.Format())
	if code == htcli.ExitUsage {
		fmt.Fprintln(opts.Err, "Run 'ht-cli --help' for usage.")
	}
	return code
}

// findGroup returns the generated group command a hand-written command
// hangs off, so `ht-cli jobs wait` sits beside the generated `ht-cli jobs get`.
func findGroup(root *cobra.Command, name string) *cobra.Command {
	for _, cmd := range root.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}
