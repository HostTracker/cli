package cmd

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/HostTracker/cli/internal/htcli"
)

func newAPICommand(opts *htcli.Options) *cobra.Command {
	var (
		jsonSource  string
		sets        []string
		queryValues []string
		headerLines []string
		key         string
	)
	cmd := &cobra.Command{
		Use:   "api <method> <path>",
		Short: "Send one request to any API address",
		Long: `Send one request to any API address.

The escape hatch for an address the generated commands do not model, and
the shortest way to reproduce a call from the reference documentation. The
request still rides the SDK, so the token, the idempotency key, the retry
ladder and the error mapping are the same as every other command's.

  ht api GET /account
  ht api GET /monitor --query limit=5 --query state=down
  ht api POST /monitor --set name=api --set type=http --set url=https://example.com
  ht api PATCH /monitor/<id> --json '{"enabled":false}'`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Resolve(); err != nil {
				return err
			}
			method := strings.ToUpper(args[0])
			switch method {
			case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			default:
				return htcli.Usagef("%q is not an HTTP method", args[0])
			}

			request := htcli.Request{Method: method, Path: args[1], Query: url.Values{}, Header: http.Header{}}
			for _, pair := range queryValues {
				name, value, ok := strings.Cut(pair, "=")
				if !ok {
					return htcli.Usagef("--query wants name=value, got %q", pair)
				}
				request.Query.Add(name, value)
			}
			for _, line := range headerLines {
				name, value, ok := strings.Cut(line, ":")
				if !ok {
					return htcli.Usagef("--header wants 'Name: value', got %q", line)
				}
				request.Header.Add(strings.TrimSpace(name), strings.TrimSpace(value))
			}
			if key != "" {
				request.Header.Set("Idempotency-Key", key)
			}

			body, err := htcli.BuildBody(jsonSource, sets, opts.In)
			if err != nil {
				return err
			}
			request.Body = body

			answer, err := opts.Do(cmd.Context(), request)
			if err != nil {
				return err
			}
			return opts.Emit(answer)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&jsonSource, "json", "", "request body: inline JSON, @file, or - for standard input")
	flags.StringArrayVar(&sets, "set", nil, "set one body member: --set name=api (repeatable)")
	flags.StringArrayVar(&queryValues, "query", nil, "one query-string entry, as name=value (repeatable)")
	flags.StringArrayVar(&headerLines, "header", nil, "one request header, as 'Name: value' (repeatable)")
	flags.StringVar(&key, "idempotency-key", "", "replay key for this write (default: a fresh one per call)")
	return cmd
}
