package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	hosttracker "github.com/HostTracker/hosttracker-sdk-go"
	"github.com/spf13/cobra"

	"github.com/HostTracker/cli/internal/htcli"
)

// addWebhooksVerify hangs the verify verb off the generated webhooks
// group. It is the one command that talks to no API at all.
func addWebhooksVerify(root *cobra.Command, opts *htcli.Options) {
	group := findGroup(root, "webhooks")
	if group == nil {
		return
	}
	group.AddCommand(newWebhooksVerifyCommand(opts))
}

func newWebhooksVerifyCommand(opts *htcli.Options) *cobra.Command {
	var (
		secrets     []string
		headerLines []string
		headersFile string
		bodyFile    string
		tolerance   time.Duration
	)
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Check a webhook signature against the delivered body",
		Long: `Check a webhook signature against the delivered body.

The body is read from standard input (or --body-file), the headers from
--header or --headers-file, which takes a captured header block:

  HT-Signature: t=1735689600,v1=6f1c...
  HT-Delivery-Id: 018f...

The signature is computed over the RAW body, so it must be the exact bytes
the endpoint received: a re-serialised copy verifies against nothing.

  ht-cli webhooks verify --secret whsec_... --headers-file headers.txt < body.json

An accepted signature exits 0 and prints the event; a rejected one exits 5.
Both signature schemes the API sends are accepted, and several --secret
values may be passed at once, which is what a secret rotation needs.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Resolve(); err != nil {
				return err
			}
			if len(secrets) == 0 {
				return htcli.Usagef("--secret is required")
			}
			body, err := readBody(bodyFile, opts.In)
			if err != nil {
				return &htcli.UsageError{Err: err}
			}
			headers, err := readHeaders(headerLines, headersFile)
			if err != nil {
				return &htcli.UsageError{Err: err}
			}

			var options *hosttracker.VerifyOptions
			if tolerance > 0 {
				options = &hosttracker.VerifyOptions{Tolerance: tolerance}
			}
			if err := hosttracker.VerifyWebhookSignature(headers, body, secrets, options); err != nil {
				return &hosttracker.Error{
					Status: http.StatusBadRequest,
					Code:   "invalid_signature",
					Title:  "The webhook signature does not verify.",
					Detail: err.Error(),
				}
			}

			event, err := hosttracker.ParseWebhookEvent(body)
			if err != nil {
				// The signature is what this command answers for; a body
				// that is authentic but not a known event is handed on
				// rather than swallowed.
				fmt.Fprintln(opts.Err, "signature verified; the body is not a HostTracker webhook event")
				if json.Valid(body) {
					return opts.Printer().Print(json.RawMessage(body))
				}
				_, writeErr := opts.Out.Write(body)
				return writeErr
			}
			fmt.Fprintln(opts.Err, "signature verified")
			return opts.Printer().Print(event)
		},
	}
	flags := cmd.Flags()
	flags.StringArrayVar(&secrets, "secret", nil, "endpoint signing secret, whsec_ prefix included (repeatable during a rotation)")
	flags.StringArrayVar(&headerLines, "header", nil, "one delivery header, as 'Name: value' (repeatable)")
	flags.StringVar(&headersFile, "headers-file", "", "file holding the delivery's header block")
	flags.StringVar(&bodyFile, "body-file", "", "file holding the raw delivered body (default: standard input)")
	flags.DurationVar(&tolerance, "tolerance", 0, "how old a signature may be (default 5m)")
	return cmd
}

func readBody(path string, stdin io.Reader) ([]byte, error) {
	if path != "" {
		return os.ReadFile(path)
	}
	return io.ReadAll(stdin)
}

// readHeaders collects the delivery headers from repeated --header flags
// and from a captured header block.
func readHeaders(lines []string, path string) (http.Header, error) {
	headers := http.Header{}
	add := func(line string) error {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			return nil
		}
		name, value, ok := strings.Cut(line, ":")
		if !ok {
			return fmt.Errorf("a header must be spelled 'Name: value', got %q", line)
		}
		headers.Add(strings.TrimSpace(name), strings.TrimSpace(value))
		return nil
	}
	for _, line := range lines {
		if err := add(line); err != nil {
			return nil, err
		}
	}
	if path != "" {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if err := add(scanner.Text()); err != nil {
				return nil, err
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}
	if len(headers) == 0 {
		return nil, fmt.Errorf("no headers given: pass --header or --headers-file")
	}
	return headers, nil
}
