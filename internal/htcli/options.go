package htcli

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	hosttracker "github.com/HostTracker/hosttracker-sdk-go"
	"github.com/spf13/pflag"

	"github.com/HostTracker/cli/internal/config"
	"github.com/HostTracker/cli/internal/output"
)

// Environment variables the global options fall back to.
const (
	EnvToken   = "HT_TOKEN"
	EnvBaseURL = "HT_BASE_URL"
	EnvProfile = "HT_PROFILE"
	EnvOutput  = "HT_OUTPUT"
)

// Options is the global flag set, plus everything a command needs to run:
// the resolved profile, the streams, and the SDK client built from them.
type Options struct {
	Profile   string
	Token     string
	BaseURL   string
	OutputFmt string
	Timeout   time.Duration
	NoRetry   bool
	Verbose   bool
	All       bool

	// Streams. Tests replace them; main leaves them at the process ones.
	Out io.Writer
	Err io.Writer
	In  io.Reader

	// UserAgent is appended to the SDK's own product token.
	UserAgent string

	// Started reports that a command body was reached. A failure raised
	// before that came from the command line, not from the API.
	Started bool

	file        *config.File
	profileName string
	profile     config.Profile
	format      output.Format
	client      *hosttracker.Client
}

// Bind declares the global flags on fs.
func (o *Options) Bind(fs *pflag.FlagSet) {
	fs.StringVar(&o.Profile, "profile", "", "configuration profile to use (default: the current one)")
	fs.StringVar(&o.Token, "token", "", "API token, overriding the profile and "+EnvToken)
	fs.StringVar(&o.BaseURL, "base-url", "", "API root (default: "+hosttracker.DefaultBaseURL+")")
	fs.StringVarP(&o.OutputFmt, "output", "o", "", "output format: "+strings.Join(output.Formats, ", ")+" (default: table on a terminal, json when piped)")
	fs.DurationVar(&o.Timeout, "timeout", 0, "per-attempt request deadline (default 30s)")
	fs.BoolVar(&o.NoRetry, "no-retry", false, "do not retry a throttled or unavailable answer")
	fs.BoolVar(&o.Verbose, "verbose", false, "report every request, its status, request id and rate-limit budget on stderr")
	fs.BoolVar(&o.All, "all", false, "walk every page of a collection instead of the first one")
}

// Streams fills in the process streams for anything left unset.
func (o *Options) Streams() {
	if o.Out == nil {
		o.Out = os.Stdout
	}
	if o.Err == nil {
		o.Err = os.Stderr
	}
	if o.In == nil {
		o.In = os.Stdin
	}
}

// Resolve loads the configuration file and settles every setting, in the
// order flag, environment, profile, default. It runs once per command.
func (o *Options) Resolve() error {
	o.Streams()
	if o.file != nil {
		return nil
	}
	file, err := config.Load()
	if err != nil {
		return err
	}
	o.file = file

	name := firstNonEmpty(o.Profile, os.Getenv(EnvProfile))
	o.profileName, o.profile = file.Resolve(name)

	format, err := output.ParseFormat(firstNonEmpty(o.OutputFmt, os.Getenv(EnvOutput), o.profile.Output, string(output.Default(o.Out))))
	if err != nil {
		return &UsageError{Err: err}
	}
	o.format = format
	return nil
}

// File is the loaded configuration document.
func (o *Options) File() *config.File { return o.file }

// ProfileName is the profile the command is running under.
func (o *Options) ProfileName() string { return o.profileName }

// ResolvedProfile is the profile's stored settings.
func (o *Options) ResolvedProfile() config.Profile { return o.profile }

// Format is the settled output format.
func (o *Options) Format() output.Format { return o.format }

// Printer writes answers to the command's output stream.
func (o *Options) Printer() output.Printer {
	return output.Printer{Out: o.Out, Format: o.format}
}

// ResolvedToken is the credential in force, from --token, the
// environment or the profile.
func (o *Options) ResolvedToken() string {
	return firstNonEmpty(o.Token, os.Getenv(EnvToken), o.profile.Token)
}

// ResolvedBaseURL is the API root in force.
func (o *Options) ResolvedBaseURL() string {
	return firstNonEmpty(o.BaseURL, os.Getenv(EnvBaseURL), o.profile.BaseURL, hosttracker.DefaultBaseURL)
}

// Client builds the SDK client the command runs on, once per command.
// Everything the CLI does on the wire goes through it, so auth, retries,
// idempotency keys and error mapping are the SDK's, not the CLI's.
func (o *Options) Client() (*hosttracker.Client, error) {
	if o.client != nil {
		return o.client, nil
	}
	if err := o.Resolve(); err != nil {
		return nil, err
	}

	opts := []hosttracker.Option{
		hosttracker.WithBaseURL(o.ResolvedBaseURL()),
		hosttracker.WithUserAgent(o.UserAgent),
	}
	if o.Timeout > 0 {
		opts = append(opts, hosttracker.WithTimeout(o.Timeout))
	}
	if o.NoRetry {
		opts = append(opts, hosttracker.WithMaxRetries(0))
	}
	if o.Verbose {
		opts = append(opts, hosttracker.WithHTTPClient(&verboseDoer{next: &http.Client{}, log: o.Err}))
	}

	client, err := hosttracker.New(o.ResolvedToken(), opts...)
	if err != nil {
		return nil, err
	}
	o.client = client
	return client, nil
}

// verboseDoer traces each attempt on stderr. It sits under the SDK's own
// transport, so what it prints is the request as it went on the wire,
// retries and idempotency key included.
type verboseDoer struct {
	next hosttracker.HttpRequestDoer
	log  io.Writer
}

func (d *verboseDoer) Do(req *http.Request) (*http.Response, error) {
	started := time.Now()
	fmt.Fprintf(d.log, "> %s %s\n", req.Method, req.URL)
	if key := req.Header.Get("Idempotency-Key"); key != "" {
		fmt.Fprintf(d.log, ">   Idempotency-Key: %s\n", key)
	}
	resp, err := d.next.Do(req)
	elapsed := time.Since(started).Round(time.Millisecond)
	if err != nil {
		fmt.Fprintf(d.log, "< %s after %s\n", err, elapsed)
		return resp, err
	}
	fmt.Fprintf(d.log, "< %d in %s\n", resp.StatusCode, elapsed)
	meta := hosttracker.MetaOf(resp)
	if meta.RequestID != "" {
		fmt.Fprintf(d.log, "<   request %s\n", meta.RequestID)
	}
	if meta.RateLimit.Metered() {
		line := "<   rate limit " + meta.RateLimit.Policy
		if meta.RateLimit.Remaining != nil {
			line += fmt.Sprintf(", %d left", *meta.RateLimit.Remaining)
		}
		if meta.RateLimit.Reset != nil {
			line += fmt.Sprintf(", resets in %ds", *meta.RateLimit.Reset)
		}
		fmt.Fprintln(d.log, line)
	}
	if meta.IdempotencyReplayed {
		fmt.Fprintln(d.log, "<   idempotency: replayed a stored answer")
	}
	if meta.RetryAfter > 0 {
		fmt.Fprintf(d.log, "<   retry after %s\n", meta.RetryAfter)
	}
	return resp, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if v := strings.TrimSpace(value); v != "" {
			return v
		}
	}
	return ""
}
