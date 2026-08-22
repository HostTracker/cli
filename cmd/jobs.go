package cmd

import (
	"fmt"
	"strings"
	"time"

	hosttracker "github.com/HostTracker/hosttracker-sdk-go"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/HostTracker/cli/internal/htcli"
)

// addJobsWait hangs the wait verb off the generated jobs group, so it
// sits beside ht-cli jobs get and ht-cli jobs cancel.
func addJobsWait(root *cobra.Command, opts *htcli.Options) {
	group := findGroup(root, "jobs")
	if group == nil {
		return
	}
	group.AddCommand(newJobsWaitCommand(opts))
}

func newJobsWaitCommand(opts *htcli.Options) *cobra.Command {
	var (
		timeout          time.Duration
		interval         time.Duration
		followInterrupts bool
	)
	cmd := &cobra.Command{
		Use:   "wait <id>",
		Short: "Poll a job until it reaches a terminal state",
		Long: `Poll a job until it reaches a terminal state.

The bulk operations answer 202 with a job id; this follows that job at the
pace the API asks for and prints it once it stops moving.

Reaching a terminal state is a success whatever that state is: succeeded,
partial, failed and cancelled all exit 0, and the printed state is what
says which. A job that was interrupted stops the wait as well, because it
makes no further progress until ht-cli jobs resume continues it.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Resolve(); err != nil {
				return err
			}
			client, err := opts.Client()
			if err != nil {
				return err
			}
			jobID, err := parseUUID(args[0])
			if err != nil {
				return err
			}
			job, err := client.WaitForJob(cmd.Context(), jobID, &hosttracker.WaitOptions{
				Timeout:           timeout,
				Interval:          interval,
				StopOnInterrupted: hosttracker.Ptr(!followInterrupts),
			})
			if job != nil {
				if printErr := opts.Printer().Print(job); printErr != nil {
					return printErr
				}
			}
			if err != nil {
				return err
			}
			if job != nil && job.State != nil && *job.State == hosttracker.JobViewStateInterrupted {
				fmt.Fprintf(opts.Err, "the job was interrupted; `ht-cli jobs resume %s` continues it\n", args[0])
			}
			return nil
		},
	}
	cmd.Flags().DurationVar(&timeout, "wait-timeout", 10*time.Minute, "how long to keep polling")
	cmd.Flags().DurationVar(&interval, "interval", 0, "poll interval when the answer asks for none (default 2s)")
	cmd.Flags().BoolVar(&followInterrupts, "follow-interrupts", false, "keep polling an interrupted job instead of returning it")
	return cmd
}

// parseUUID reads an identifier argument, so a typo fails before it costs
// a round trip.
func parseUUID(value string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return uuid.UUID{}, htcli.Usagef("%q is not a valid id", value)
	}
	return parsed, nil
}
