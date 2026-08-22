package cmd

import (
	"github.com/spf13/cobra"
)

func newCompletionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion <bash|zsh|fish|powershell>",
		Short: "Print a shell completion script",
		Long: `Print a shell completion script.

  bash        ht-cli completion bash > /etc/bash_completion.d/ht-cli
  zsh         ht-cli completion zsh > "${fpath[1]}/_ht-cli"
  fish        ht-cli completion fish > ~/.config/fish/completions/ht-cli.fish
  powershell  ht-cli completion powershell | Out-String | Invoke-Expression

Installed through Homebrew, the bash and zsh scripts are already in place.`,
		Args:                  cobra.ExactArgs(1),
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletionV2(cmd.OutOrStdout(), true)
			case "zsh":
				return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			default:
				return cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			}
		},
	}
	return cmd
}
