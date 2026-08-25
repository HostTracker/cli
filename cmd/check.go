package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	hosttracker "github.com/HostTracker/hosttracker-sdk-go"
	"github.com/spf13/cobra"

	"github.com/HostTracker/cli/internal/htcli"
)

func newCheckCommand(opts *htcli.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Run a one-off check without creating a monitor",
		Long: `Run a one-off check without creating a monitor.

The full instant-check surface, including the catalogue of types and the
history of past runs, is under ht-cli instant-checks.`,
	}
	cmd.AddCommand(newCheckRunCommand(opts))
	return cmd
}

func newCheckRunCommand(opts *htcli.Options) *cobra.Command {
	var (
		checkType  string
		pools      []string
		countries  []string
		cities     []string
		bodySource string
		strictTls  bool
		wait       bool
		waitFor    time.Duration
		key        string
	)

	cmd := &cobra.Command{
		Use:   "run <url>",
		Short: "Start an instant check, and follow it to its result",
		Long: `Start an instant check, and follow it to its result.

  ht-cli check run https://www.host-tracker.com --wait
  ht-cli check run example.com --type ping --country de --country us --wait
  ht-cli check run example.com:443 --type port --pool premium
  ht-cli check run https://example.com --strict-tls --wait

Without --wait the command prints the 202 receipt, whose id is what
ht-cli instant-checks get <db-id> <id> reads later. With --wait it polls
at the pace the API asks for and prints the finished result.

--json supplies the whole request body for a shape the flags do not
reach, and the flags then fill in what it left out.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Resolve(); err != nil {
				return err
			}
			client, err := opts.Client()
			if err != nil {
				return err
			}

			var request hosttracker.IcCreateRequest
			if bodySource != "" {
				raw, err := htcli.ReadJSONSource(bodySource, opts.In)
				if err != nil {
					return &htcli.UsageError{Err: err}
				}
				if err := json.Unmarshal(raw, &request); err != nil {
					return htcli.Usagef("--json is not a valid instant-check request: %v", err)
				}
			}
			request.Url = args[0]
			if checkType != "" {
				kind := hosttracker.IcCreateRequestType(checkType)
				if !kind.Valid() {
					return htcli.Usagef("--type: %q is not a known check type; ht-cli instant-checks list-type prints the catalogue", checkType)
				}
				request.Type = &kind
			}
			if len(pools) > 0 {
				request.Pools = &pools
			}
			if clause := locationClause(countries, cities); clause != nil {
				request.Locations = &[]hosttracker.IcLocationClause{*clause}
			}
			if strictTls {
				request.StrictTls = &strictTls
			}

			if !wait {
				params := &hosttracker.CreateInstantCheckParams{}
				if key != "" {
					params.IdempotencyKey = &key
				}
				resp, err := client.CreateInstantCheckWithResponse(cmd.Context(), params, request)
				if err != nil {
					return err
				}
				if err := opts.Printer().Print(resp.JSON202); err != nil {
					return err
				}
				fmt.Fprintf(opts.Err, "started; pass --wait to follow it, or read it with `ht-cli instant-checks get %d %s`\n",
					resp.JSON202.DbId, resp.JSON202.Id)
				return nil
			}

			result, err := client.RunCheck(cmd.Context(), request, &hosttracker.RunCheckOptions{
				Timeout:        waitFor,
				IdempotencyKey: key,
			})
			if result != nil {
				if printErr := opts.Printer().Print(result); printErr != nil {
					return printErr
				}
			}
			return err
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&checkType, "type", "", "which kind of check to run (default http)")
	flags.StringArrayVar(&pools, "pool", nil, "monitoring-location pool to run from (repeatable)")
	flags.StringArrayVar(&countries, "country", nil, "run only from locations in this country (repeatable)")
	flags.StringArrayVar(&cities, "city", nil, "run only from locations in this city (repeatable)")
	flags.BoolVar(&strictTls, "strict-tls", false, "validate the TLS handshake strictly (untrusted root, incomplete chain, name mismatch and self-signed certificates fail); http checks only")
	flags.StringVar(&bodySource, "json", "", "request body: inline JSON, @file, or - for standard input")
	flags.BoolVar(&wait, "wait", false, "follow the check until it finishes and print the result")
	flags.DurationVar(&waitFor, "wait-timeout", 2*time.Minute, "how long --wait keeps polling")
	flags.StringVar(&key, "idempotency-key", "", "replay key for this run (default: a fresh one per call)")
	return cmd
}

// locationClause builds the single where-to-run clause the flags express.
// Richer shapes, several clauses or an exclusion, go through --json.
func locationClause(countries, cities []string) *hosttracker.IcLocationClause {
	if len(countries) == 0 && len(cities) == 0 {
		return nil
	}
	clause := &hosttracker.IcLocationClause{}
	if len(countries) > 0 {
		clause.Countries = &countries
	}
	if len(cities) > 0 {
		clause.Cities = &cities
	}
	return clause
}
