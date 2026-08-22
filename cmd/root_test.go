package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/HostTracker/cli/internal/htcli"
)

// exec runs the whole CLI against a fake API and reports what it printed
// and which exit code it would have left behind.
func exec(t *testing.T, server *httptest.Server, stdin string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	t.Setenv("HT_CONFIG_DIR", t.TempDir())
	t.Setenv(htcli.EnvToken, "test")
	t.Setenv(htcli.EnvBaseURL, server.URL)
	t.Setenv(htcli.EnvOutput, "json")

	out, errs := &bytes.Buffer{}, &bytes.Buffer{}
	code = Main(context.Background(), args, strings.NewReader(stdin), out, errs)
	return out.String(), errs.String(), code
}

func fakeAPI(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-Id", "req-1")
		switch {
		case r.URL.Path == "/account":
			_, _ = io.WriteString(w, `{"id":"a-1","email":"cli@example.com"}`)
		case r.URL.Path == "/monitor" && r.Method == http.MethodGet:
			if r.URL.Query().Get("cursor") == "c1" {
				_, _ = io.WriteString(w, `{"data":[{"id":"m-3"}],"nextCursor":null,"hasMore":false}`)
				return
			}
			_, _ = io.WriteString(w, `{"data":[{"id":"m-1"},{"id":"m-2"}],"nextCursor":"c1","hasMore":true}`)
		case r.URL.Path == "/monitor/m-1":
			_, _ = io.WriteString(w, `{"id":"m-1","name":"one"}`)
		default:
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, `{"type":"https://api2.host-tracker.com/problems/not_found","title":"No such thing.","status":404,"code":"not_found"}`)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func TestEveryGroupIsRegistered(t *testing.T) {
	opts := &htcli.Options{Out: io.Discard, Err: io.Discard, In: strings.NewReader("")}
	root := New(opts)
	want := []string{
		"account", "alerts", "api", "auth", "check", "config", "contacts",
		"incidents", "instant-checks", "jobs", "maintenance", "monitor-types",
		"monitoring-locations", "monitors", "reports", "results", "status-pages",
		"version", "webhooks",
	}
	have := map[string]bool{}
	for _, cmd := range root.Commands() {
		have[cmd.Name()] = true
	}
	for _, name := range want {
		if !have[name] {
			t.Errorf("the root command has no %q", name)
		}
	}
	if htcli.Count() < 100 {
		t.Errorf("only %d operation commands are registered", htcli.Count())
	}
}

// TestConvenienceVerbsJoinTheirGroups proves the hand-written verbs sit
// next to the generated ones rather than in a group of their own.
func TestConvenienceVerbsJoinTheirGroups(t *testing.T) {
	opts := &htcli.Options{Out: io.Discard, Err: io.Discard, In: strings.NewReader("")}
	root := New(opts)
	cases := map[string]string{"jobs": "wait", "webhooks": "verify"}
	for group, verb := range cases {
		parent := findGroup(root, group)
		if parent == nil {
			t.Fatalf("no %s group", group)
		}
		found := false
		for _, child := range parent.Commands() {
			if child.Name() == verb {
				found = true
			}
		}
		if !found {
			t.Errorf("ht %s %s is missing", group, verb)
		}
	}
}

func TestReadPrintsJSONWhenPiped(t *testing.T) {
	out, _, code := exec(t, fakeAPI(t), "", "account", "get")
	if code != htcli.ExitOK {
		t.Fatalf("exit %d, output %q", code, out)
	}
	var account map[string]string
	if err := json.Unmarshal([]byte(out), &account); err != nil {
		t.Fatalf("output is not JSON: %v (%q)", err, out)
	}
	if account["email"] != "cli@example.com" {
		t.Errorf("account = %v", account)
	}
}

func TestNotFoundExitsFour(t *testing.T) {
	_, errs, code := exec(t, fakeAPI(t), "", "monitors", "get", "m-404")
	if code != htcli.ExitNotFound {
		t.Fatalf("exit %d, stderr %q", code, errs)
	}
	if !strings.Contains(errs, "not_found") {
		t.Errorf("stderr = %q", errs)
	}
}

func TestUnknownCommandExitsTwo(t *testing.T) {
	_, errs, code := exec(t, fakeAPI(t), "", "nonsense")
	if code != htcli.ExitUsage {
		t.Fatalf("exit %d, stderr %q", code, errs)
	}
	if !strings.Contains(errs, "ht --help") {
		t.Errorf("a usage failure should point at the help, got %q", errs)
	}
}

func TestUnknownFlagExitsTwo(t *testing.T) {
	_, _, code := exec(t, fakeAPI(t), "", "monitors", "list", "--nonsense")
	if code != htcli.ExitUsage {
		t.Errorf("exit %d, want %d", code, htcli.ExitUsage)
	}
}

func TestAllWalksEveryPage(t *testing.T) {
	out, _, code := exec(t, fakeAPI(t), "", "monitors", "list", "--all")
	if code != htcli.ExitOK {
		t.Fatalf("exit %d", code)
	}
	var envelope struct {
		Data []map[string]string `json:"data"`
	}
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Data) != 3 {
		t.Errorf("got %d rows, want 3", len(envelope.Data))
	}
}

func TestOutputFormatFlag(t *testing.T) {
	out, _, code := exec(t, fakeAPI(t), "", "account", "get", "--output", "yaml")
	if code != htcli.ExitOK {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(out, "email: cli@example.com") {
		t.Errorf("out = %q", out)
	}
	_, _, code = exec(t, fakeAPI(t), "", "account", "get", "--output", "xml")
	if code != htcli.ExitUsage {
		t.Errorf("an unknown format exited %d, want %d", code, htcli.ExitUsage)
	}
}

func TestRawAPICommand(t *testing.T) {
	out, _, code := exec(t, fakeAPI(t), "", "api", "GET", "/monitor/m-1")
	if code != htcli.ExitOK {
		t.Fatalf("exit %d, out %q", code, out)
	}
	if !strings.Contains(out, `"name": "one"`) {
		t.Errorf("out = %q", out)
	}
	if _, _, code := exec(t, fakeAPI(t), "", "api", "SING", "/monitor"); code != htcli.ExitUsage {
		t.Errorf("an invented method exited %d", code)
	}
}

func TestVersion(t *testing.T) {
	out, _, code := exec(t, fakeAPI(t), "", "version")
	if code != htcli.ExitOK {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(out, `"sdk"`) {
		t.Errorf("out = %q", out)
	}
}

func TestConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HT_CONFIG_DIR", dir)
	out, errs := &bytes.Buffer{}, &bytes.Buffer{}
	if code := Main(context.Background(), []string{"config", "set", "output", "yaml"}, strings.NewReader(""), out, errs); code != 0 {
		t.Fatalf("exit %d, stderr %q", code, errs)
	}
	out.Reset()
	if code := Main(context.Background(), []string{"config", "get", "output"}, strings.NewReader(""), out, errs); code != 0 {
		t.Fatalf("exit %d, stderr %q", code, errs)
	}
	if strings.TrimSpace(out.String()) != "yaml" {
		t.Errorf("got %q", out.String())
	}
}

func TestWebhookVerify(t *testing.T) {
	// A delivery signed with the published scheme, verified offline.
	body := `{"id":"d_1","event":"monitor.down","occurredAt":1735689600,"apiVersion":"v2","data":{}}`
	secret := "whsec_" + strings.Repeat("a", 32)
	signature := signHT(t, secret, body)

	server := fakeAPI(t)
	out, errs, code := exec(t, server, body, "webhooks", "verify", "--secret", secret, "--header", signature)
	if code != htcli.ExitOK {
		t.Fatalf("exit %d, stderr %q", code, errs)
	}
	if !strings.Contains(out, "monitor.down") {
		t.Errorf("out = %q", out)
	}

	_, _, code = exec(t, server, body, "webhooks", "verify", "--secret", "whsec_"+strings.Repeat("b", 32), "--header", signature)
	if code != htcli.ExitInvalid {
		t.Errorf("a wrong secret exited %d, want %d", code, htcli.ExitInvalid)
	}
}
