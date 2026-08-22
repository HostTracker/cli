//go:build live

// A smoke test against a real API. It is opt-in, needs a real token, and
// is never part of `go test ./...`:
//
//	HT_TOKEN=... go test -tags live ./cmd -run TestLive -v
//	HT_TOKEN_FILE=~/.ht-cli-token HT_BASE_URL=https://api2.host-tracker.com \
//	  go test -tags live ./cmd -run TestLive -v
//
// It only reads.
package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/HostTracker/cli/internal/htcli"
)

// liveToken is the credential, from HT_TOKEN or the file HT_TOKEN_FILE
// names. An absent one skips the test rather than failing it.
func liveToken(t *testing.T) string {
	t.Helper()
	if token := strings.TrimSpace(os.Getenv(htcli.EnvToken)); token != "" {
		return token
	}
	if path := strings.TrimSpace(os.Getenv("HT_TOKEN_FILE")); path != "" {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("HT_TOKEN_FILE: %v", err)
		}
		return strings.TrimSpace(string(raw))
	}
	t.Skip("set HT_TOKEN or HT_TOKEN_FILE to run the live smoke")
	return ""
}

func live(t *testing.T, args ...string) string {
	t.Helper()
	t.Setenv("HT_CONFIG_DIR", t.TempDir())
	t.Setenv(htcli.EnvToken, liveToken(t))
	t.Setenv(htcli.EnvOutput, "json")

	out, errs := &bytes.Buffer{}, &bytes.Buffer{}
	if code := Main(context.Background(), args, strings.NewReader(""), out, errs); code != htcli.ExitOK {
		t.Fatalf("ht-cli %s exited %d: %s", strings.Join(args, " "), code, errs)
	}
	return out.String()
}

func TestLiveAccount(t *testing.T) {
	var account map[string]any
	if err := json.Unmarshal([]byte(live(t, "account", "get")), &account); err != nil {
		t.Fatal(err)
	}
	if account["id"] == nil {
		t.Errorf("the account carries no id: %v", account)
	}
	t.Logf("account %v", account["id"])
}

func TestLiveMonitors(t *testing.T) {
	var page struct {
		Data    []map[string]any `json:"data"`
		HasMore bool             `json:"hasMore"`
	}
	if err := json.Unmarshal([]byte(live(t, "monitors", "list", "--limit", "1")), &page); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d row(s), more pages: %v", len(page.Data), page.HasMore)
}

func TestLiveMonitorTypes(t *testing.T) {
	// The reference tier answers without a credential too, so this also
	// proves the anonymous path.
	var page struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal([]byte(live(t, "monitor-types", "list")), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Data) == 0 {
		t.Error("the type catalogue came back empty")
	}
}
