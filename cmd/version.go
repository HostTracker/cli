package cmd

import (
	"fmt"
	"runtime"
	"runtime/debug"

	hosttracker "github.com/HostTracker/hosttracker-sdk-go"
	"github.com/spf13/cobra"

	"github.com/HostTracker/cli/internal/htcli"
)

// Build information. The release workflow stamps these with -ldflags; a
// `go install` build leaves them empty and they are read back off the
// embedded build info instead.
var (
	version = ""
	commit  = ""
	date    = ""
)

// Version is the CLI's version, as reported by `ht-cli version`.
func Version() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

// userAgent is the product token the CLI adds to the SDK's own.
func userAgent() string { return "hosttracker-cli/" + Version() }

func newVersionCommand(opts *htcli.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of ht-cli, its SDK and its toolchain",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Resolve(); err != nil {
				return err
			}
			info := map[string]any{
				"version":    Version(),
				"sdk":        hosttracker.Version,
				"apiVersion": hosttracker.APIVersion,
				"go":         runtime.Version(),
				"platform":   runtime.GOOS + "/" + runtime.GOARCH,
				"commands":   htcli.Count(),
			}
			if commit != "" {
				info["commit"] = commit
			}
			if date != "" {
				info["built"] = date
			}
			if opts.Format() == "table" {
				fmt.Fprintf(opts.Out, "ht-cli %s (sdk %s, %s, %s/%s)\n",
					Version(), hosttracker.Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
				return nil
			}
			return opts.Printer().Print(info)
		},
	}
}
