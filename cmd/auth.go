package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	hosttracker "github.com/HostTracker/hosttracker-sdk-go"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/HostTracker/cli/internal/config"
	"github.com/HostTracker/cli/internal/htcli"
	"github.com/HostTracker/cli/internal/output"
)

func newAuthCommand(opts *htcli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Store, inspect and remove the API token",
		Long: `Store, inspect and remove the API token.

The token is kept in the profile file (ht-cli config path), which is written
0600. A token given with --token, or found in ` + htcli.EnvToken + `, wins over the
stored one for that command, so a CI job needs no file at all.`,
	}
	cmd.AddCommand(newAuthLoginCommand(opts), newAuthLogoutCommand(opts), newAuthStatusCommand(opts))
	return cmd
}

func newAuthLoginCommand(opts *htcli.Options) *cobra.Command {
	var (
		tokenStdin bool
		verify     bool
	)
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Store an API token in a profile",
		Long: `Store an API token in a profile.

The token is read from --token, from ` + htcli.EnvToken + `, from standard input with
--token-stdin, or from a prompt that does not echo. It is verified against
GET /account before it is written, unless --verify=false.

  ht-cli auth login                              prompt for a token
  ht-cli auth login --token-stdin < token.txt    read it from a file
  ht-cli auth login --profile staging            store it under another profile`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Resolve(); err != nil {
				return err
			}
			token, err := readToken(opts, tokenStdin)
			if err != nil {
				return err
			}
			if token == "" {
				return htcli.Usagef("no token given")
			}

			profile := opts.ResolvedProfile()
			profile.Token = token
			if opts.BaseURL != "" {
				profile.BaseURL = opts.BaseURL
			}

			if verify {
				opts.Token = token
				if err := verifyToken(cmd, opts); err != nil {
					return err
				}
			}

			file := opts.File()
			file.Put(opts.ProfileName(), profile)
			if err := file.Save(); err != nil {
				return err
			}
			fmt.Fprintf(opts.Err, "token stored in profile %q (%s)\n", opts.ProfileName(), file.FilePath())
			return nil
		},
	}
	cmd.Flags().BoolVar(&tokenStdin, "token-stdin", false, "read the token from standard input")
	cmd.Flags().BoolVar(&verify, "verify", true, "check the token against the API before storing it")
	return cmd
}

func newAuthLogoutCommand(opts *htcli.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove a profile's stored token",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Resolve(); err != nil {
				return err
			}
			file := opts.File()
			profile := opts.ResolvedProfile()
			if profile.Token == "" {
				fmt.Fprintf(opts.Err, "profile %q holds no token\n", opts.ProfileName())
				return nil
			}
			profile.Token = ""
			file.Put(opts.ProfileName(), profile)
			if err := file.Save(); err != nil {
				return err
			}
			fmt.Fprintf(opts.Err, "token removed from profile %q\n", opts.ProfileName())
			return nil
		},
	}
}

func newAuthStatusCommand(opts *htcli.Options) *cobra.Command {
	var offline bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Report which credential is in force and whether it works",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Resolve(); err != nil {
				return err
			}
			token := opts.ResolvedToken()
			status := map[string]any{
				"profile":       opts.ProfileName(),
				"configFile":    opts.File().FilePath(),
				"baseUrl":       opts.ResolvedBaseURL(),
				"tokenPresent":  token != "",
				"tokenRedacted": config.Redact(token),
				"tokenSource":   tokenSource(opts),
			}
			if !offline && token != "" {
				account, err := readAccount(cmd, opts)
				if err != nil {
					return err
				}
				status["account"] = account
			}
			return opts.Printer().Print(status)
		},
	}
	cmd.Flags().BoolVar(&offline, "offline", false, "report the stored settings without calling the API")
	return cmd
}

// tokenSource names where the credential in force came from, which is the
// first thing to check when the wrong account answers.
func tokenSource(opts *htcli.Options) string {
	switch {
	case strings.TrimSpace(opts.Token) != "":
		return "--token"
	case strings.TrimSpace(os.Getenv(htcli.EnvToken)) != "":
		return htcli.EnvToken
	case opts.ResolvedProfile().Token != "":
		return "profile " + opts.ProfileName()
	default:
		return "none"
	}
}

// readToken collects the token from the flag, the environment, standard
// input or a prompt.
func readToken(opts *htcli.Options, fromStdin bool) (string, error) {
	if token := strings.TrimSpace(opts.Token); token != "" {
		return token, nil
	}
	if token := strings.TrimSpace(os.Getenv(htcli.EnvToken)); token != "" {
		return token, nil
	}
	if fromStdin {
		reader := bufio.NewReader(opts.In)
		line, err := reader.ReadString('\n')
		if err != nil && line == "" {
			return "", err
		}
		return strings.TrimSpace(line), nil
	}
	if file, ok := opts.In.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		fmt.Fprint(opts.Err, "API token: ")
		raw, err := term.ReadPassword(int(file.Fd()))
		fmt.Fprintln(opts.Err)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(raw)), nil
	}
	return "", htcli.Usagef("no token given: pass --token, set %s, or use --token-stdin", htcli.EnvToken)
}

// verifyToken proves the credential before it is written, so a typo fails
// at login rather than at the next command.
func verifyToken(cmd *cobra.Command, opts *htcli.Options) error {
	if _, err := readAccount(cmd, opts); err != nil {
		if hosttracker.IsCode(err, hosttracker.CodeInvalidToken) {
			return err
		}
		if e, ok := hosttracker.AsError(err); ok && e.Status == 0 {
			// Unreachable is not a bad token; say so and store it.
			fmt.Fprintf(opts.Err, "could not reach %s to verify the token: %s\n", opts.ResolvedBaseURL(), e.Err)
			return nil
		}
		return err
	}
	return nil
}

// readAccount reads GET /account through the SDK.
func readAccount(cmd *cobra.Command, opts *htcli.Options) (any, error) {
	answer, err := opts.Do(cmd.Context(), htcli.Request{Method: "GET", Path: "/account"})
	if err != nil {
		return nil, err
	}
	return output.Generic(answer.Body)
}
