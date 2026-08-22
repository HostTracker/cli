package htcli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	hosttracker "github.com/HostTracker/hosttracker-sdk-go"

	"github.com/HostTracker/cli/internal/output"
)

func TestExitCode(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"success", nil, ExitOK},
		{"plain failure", errors.New("something"), ExitError},
		{"usage", Usagef("bad flag"), ExitUsage},
		{"invalid token", &hosttracker.Error{Status: 401, Code: hosttracker.CodeInvalidToken}, ExitAuth},
		{"missing scope", &hosttracker.Error{Status: 403, Code: hosttracker.CodeMissingScope}, ExitAuth},
		{"forbidden with an unknown code", &hosttracker.Error{Status: 403, Code: "not_allowed"}, ExitAuth},
		{"not found", &hosttracker.Error{Status: 404, Code: hosttracker.CodeNotFound}, ExitNotFound},
		{"validation", &hosttracker.Error{Status: 422, Code: "validation_failed"}, ExitInvalid},
		{"conflict", &hosttracker.Error{Status: 409, Code: "conflict"}, ExitInvalid},
		// Both are 429, and they want opposite handling; only the code
		// separates them, which is why the table branches on it.
		{"rate limited", &hosttracker.Error{Status: 429, Code: hosttracker.CodeRateLimited}, ExitRateLimit},
		{"quota exceeded", &hosttracker.Error{Status: 429, Code: hosttracker.CodeQuotaExceeded}, ExitRateLimit},
		{"server fault", &hosttracker.Error{Status: 502, Code: hosttracker.CodeHTTPError}, ExitNetwork},
		{"unreachable", &hosttracker.Error{Code: hosttracker.CodeNetworkError}, ExitNetwork},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ExitCode(c.err); got != c.want {
				t.Errorf("ExitCode = %d, want %d", got, c.want)
			}
		})
	}
}

func TestPrintErrorAsText(t *testing.T) {
	err := &hosttracker.Error{
		Status:    422,
		Code:      "validation_failed",
		Type:      "https://api2.host-tracker.com/problems/validation_failed",
		Title:     "The request is not valid.",
		Detail:    "url must be absolute.",
		Errors:    []map[string]any{{"pointer": "/url", "reason": "not_absolute"}},
		RequestID: "req-7",
	}
	var b bytes.Buffer
	PrintError(&b, err, output.FormatTable)
	text := b.String()
	for _, want := range []string{"validation_failed", "HTTP 422", "url must be absolute.", "/url: reason=not_absolute", "request req-7"} {
		if !strings.Contains(text, want) {
			t.Errorf("the report does not mention %q:\n%s", want, text)
		}
	}
}

func TestPrintErrorAsProblemDocument(t *testing.T) {
	err := &hosttracker.Error{
		Status: 404,
		Code:   "not_found",
		Body:   []byte(`{"type":"https://api2.host-tracker.com/problems/not_found","code":"not_found","status":404}`),
	}
	var b bytes.Buffer
	PrintError(&b, err, output.FormatJSON)
	if !strings.Contains(b.String(), `"code": "not_found"`) {
		t.Errorf("--output json should print the problem document, got:\n%s", b.String())
	}
}

func TestPrintErrorSynthesisesADocument(t *testing.T) {
	err := &hosttracker.Error{Code: hosttracker.CodeNetworkError, Err: errors.New("dial tcp: refused")}
	var b bytes.Buffer
	PrintError(&b, err, output.FormatJSON)
	if !strings.Contains(b.String(), "network_error") || !strings.Contains(b.String(), "refused") {
		t.Errorf("a failure with no document still needs a report, got:\n%s", b.String())
	}
}
