package htcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	hosttracker "github.com/HostTracker/hosttracker-sdk-go"

	"github.com/HostTracker/cli/internal/output"
)

// Exit codes. A command that succeeded exits 0; everything else picks the
// code that says what kind of failure it was, so a script can branch
// without parsing the message.
const (
	// ExitOK is a command that did what it was asked.
	ExitOK = 0
	// ExitError is a failure with no more specific code.
	ExitError = 1
	// ExitUsage is a malformed command line: an unknown flag, a missing
	// argument, a bad value.
	ExitUsage = 2
	// ExitAuth is a missing, rejected or under-scoped credential.
	ExitAuth = 3
	// ExitNotFound is an address that names nothing.
	ExitNotFound = 4
	// ExitInvalid is a request the API refused on its merits: a
	// validation problem, a conflict, a precondition.
	ExitInvalid = 5
	// ExitRateLimit is a throttle or an exhausted quota.
	ExitRateLimit = 6
	// ExitNetwork is a failure to reach the API, or a fault on its side.
	ExitNetwork = 7
)

// UsageError marks a failure caused by the command line itself, so it
// exits 2 and the usage text is worth showing.
type UsageError struct {
	Err error
}

// Usagef builds a UsageError.
func Usagef(format string, args ...any) error {
	return &UsageError{Err: fmt.Errorf(format, args...)}
}

// Error implements error.
func (e *UsageError) Error() string { return e.Err.Error() }

// Unwrap exposes the cause.
func (e *UsageError) Unwrap() error { return e.Err }

// ExitCode maps an error onto the exit-code table.
func ExitCode(err error) int {
	if err == nil {
		return ExitOK
	}
	var usage *UsageError
	if errors.As(err, &usage) {
		return ExitUsage
	}
	apiErr, ok := hosttracker.AsError(err)
	if !ok {
		return ExitError
	}
	switch apiErr.Code {
	case hosttracker.CodeNetworkError:
		return ExitNetwork
	case hosttracker.CodeNotFound:
		return ExitNotFound
	case hosttracker.CodeInvalidToken, hosttracker.CodeMissingScope:
		return ExitAuth
	case hosttracker.CodeRateLimited, hosttracker.CodeQuotaExceeded:
		return ExitRateLimit
	}
	switch {
	case apiErr.Status == http.StatusUnauthorized || apiErr.Status == http.StatusForbidden:
		return ExitAuth
	case apiErr.Status == http.StatusNotFound:
		return ExitNotFound
	case apiErr.Status == http.StatusTooManyRequests:
		return ExitRateLimit
	case apiErr.Status >= 500:
		return ExitNetwork
	case apiErr.Status >= 400:
		return ExitInvalid
	default:
		return ExitError
	}
}

// PrintError writes a failure to w. An API failure is reported as the
// RFC 9457 problem it is: the machine code, the human detail, the
// offending members, and the request id to quote in a support request.
// Under --output json the whole problem document is printed instead, so
// a script can read it.
func PrintError(w io.Writer, err error, format output.Format) {
	if err == nil {
		return
	}
	apiErr, ok := hosttracker.AsError(err)
	if !ok {
		fmt.Fprintf(w, "ht-cli: %s\n", err)
		return
	}

	if format == output.FormatJSON || format == output.FormatYAML {
		if printProblem(w, apiErr, format) {
			return
		}
	}

	head := apiErr.Code
	if apiErr.Status > 0 {
		head = fmt.Sprintf("%s (HTTP %d)", apiErr.Code, apiErr.Status)
	}
	fmt.Fprintf(w, "ht-cli: %s\n", head)
	if detail := strings.TrimSpace(apiErr.Detail); detail != "" {
		fmt.Fprintf(w, "    %s\n", detail)
	} else if title := strings.TrimSpace(apiErr.Title); title != "" {
		fmt.Fprintf(w, "    %s\n", title)
	} else if apiErr.Err != nil {
		fmt.Fprintf(w, "    %s\n", apiErr.Err)
	}
	for _, item := range apiErr.Errors {
		fmt.Fprintf(w, "    %s\n", problemItem(item))
	}
	if apiErr.RetryAfter > 0 {
		fmt.Fprintf(w, "    retry after %s\n", apiErr.RetryAfter)
	}
	if apiErr.RequestID != "" {
		fmt.Fprintf(w, "    request %s\n", apiErr.RequestID)
	}
	if apiErr.Type != "" {
		fmt.Fprintf(w, "    %s\n", apiErr.Type)
	}
}

// printProblem writes the problem document itself, and reports whether it
// managed to.
func printProblem(w io.Writer, apiErr *hosttracker.Error, format output.Format) bool {
	var doc any
	if len(apiErr.Body) > 0 && json.Unmarshal(apiErr.Body, &doc) == nil {
		if _, isObject := doc.(map[string]any); isObject {
			return output.Printer{Out: w, Format: format}.Print(doc) == nil
		}
	}
	// A failure that carried no document still deserves the same shape.
	synthetic := map[string]any{
		"code":   apiErr.Code,
		"status": apiErr.Status,
		"title":  apiErr.Title,
	}
	if apiErr.Detail != "" {
		synthetic["detail"] = apiErr.Detail
	}
	if apiErr.RequestID != "" {
		synthetic["requestId"] = apiErr.RequestID
	}
	if apiErr.Err != nil {
		synthetic["detail"] = apiErr.Err.Error()
	}
	return output.Printer{Out: w, Format: format}.Print(synthetic) == nil
}

// problemItem renders one errors[] entry: where the problem is, followed
// by whatever the code published to fix it.
func problemItem(item map[string]any) string {
	var where string
	if p, ok := item["pointer"].(string); ok && p != "" {
		where = p
	} else if p, ok := item["parameter"].(string); ok && p != "" {
		where = "?" + p
	}
	rest := make([]string, 0, len(item))
	for _, key := range []string{"reason", "value", "allowed", "min", "max", "didYouMean", "existingId", "limit", "remaining", "retryAfter"} {
		if value, ok := item[key]; ok {
			rest = append(rest, fmt.Sprintf("%s=%v", key, compact(value)))
		}
	}
	switch {
	case where != "" && len(rest) > 0:
		return where + ": " + strings.Join(rest, " ")
	case where != "":
		return where
	case len(rest) > 0:
		return strings.Join(rest, " ")
	default:
		return compact(item)
	}
}

func compact(value any) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(raw)
	}
}
