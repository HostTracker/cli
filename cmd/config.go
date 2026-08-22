package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/HostTracker/cli/internal/config"
	"github.com/HostTracker/cli/internal/htcli"
)

func newConfigCommand(opts *htcli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Read and write the stored settings",
		Long: `Read and write the stored settings.

Each profile holds a token, a base-url and a default output format. The
file is written 0600 under the OS configuration directory; ` + config.EnvConfigDir + `
moves it somewhere else, which is what a throwaway or CI profile wants.`,
	}
	cmd.AddCommand(
		newConfigGetCommand(opts),
		newConfigSetCommand(opts),
		newConfigListCommand(opts),
		newConfigPathCommand(opts),
	)
	return cmd
}

func newConfigGetCommand(opts *htcli.Options) *cobra.Command {
	return &cobra.Command{
		Use:       "get <key>",
		Short:     "Print one setting of the current profile",
		Args:      cobra.ExactArgs(1),
		ValidArgs: config.Settable,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Resolve(); err != nil {
				return err
			}
			value, err := opts.ResolvedProfile().Get(args[0])
			if err != nil {
				return &htcli.UsageError{Err: err}
			}
			if args[0] == "token" {
				value = config.Redact(value)
			}
			fmt.Fprintln(opts.Out, value)
			return nil
		},
	}
}

func newConfigSetCommand(opts *htcli.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Write one setting of the current profile",
		Long: `Write one setting of the current profile.

  ht-cli config set base-url https://api2.host-tracker.com
  ht-cli config set output json
  ht-cli config set token "$HT_TOKEN"

Known keys: ` + strings.Join(config.Settable, ", ") + `.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Resolve(); err != nil {
				return err
			}
			profile := opts.ResolvedProfile()
			if err := profile.Set(args[0], args[1]); err != nil {
				return &htcli.UsageError{Err: err}
			}
			file := opts.File()
			file.Put(opts.ProfileName(), profile)
			if err := file.Save(); err != nil {
				return err
			}
			fmt.Fprintf(opts.Err, "%s set in profile %q\n", args[0], opts.ProfileName())
			return nil
		},
	}
}

func newConfigListCommand(opts *htcli.Options) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List the profiles and their settings",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Resolve(); err != nil {
				return err
			}
			file := opts.File()
			rows := make([]any, 0, len(file.Profiles))
			for _, name := range file.Names() {
				profile := file.Profiles[name]
				rows = append(rows, map[string]any{
					"name":    name,
					"current": name == opts.ProfileName(),
					"baseUrl": profile.BaseURL,
					"output":  profile.Output,
					"token":   config.Redact(profile.Token),
				})
			}
			return opts.Printer().Print(map[string]any{"data": rows})
		},
	}
}

func newConfigPathCommand(opts *htcli.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print where the settings are stored",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.Path()
			if err != nil {
				return err
			}
			fmt.Fprintln(opts.Out, path)
			return nil
		},
	}
}
