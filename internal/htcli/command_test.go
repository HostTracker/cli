package htcli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// seen is what the fake API recorded about the request the CLI sent.
type seen struct {
	Method string
	Path   string
	Query  string
	Body   string
	Header http.Header
}

// serve stands up a fake API that records one request and answers with
// the given body.
func serve(t *testing.T, record *seen, answer string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		*record = seen{Method: r.Method, Path: r.URL.Path, Query: r.URL.RawQuery, Body: string(raw), Header: r.Header.Clone()}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, answer)
	}))
	t.Cleanup(server.Close)
	return server
}

// run executes one operation command with the given command line.
func run(t *testing.T, op Operation, server *httptest.Server, stdin string, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	opts := &Options{
		Out: out, Err: io.Discard, In: strings.NewReader(stdin),
		BaseURL: server.URL, Token: "test", OutputFmt: "json",
	}
	t.Setenv("HT_CONFIG_DIR", t.TempDir())
	cmd := OperationCommand(op, opts)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(context.Background())
	return out.String(), err
}

var listOp = Operation{
	Group: "monitors", Name: "list", ID: "listMonitor", Method: "GET", Path: "/monitor",
	Summary: "List monitors.", QueryPath: "/monitor/q", Paged: true,
	Params: []Param{
		{Name: "limit", Flag: "limit", In: "query", Type: "integer", Help: "Rows."},
		{Name: "cursor", Flag: "cursor", In: "query", Type: "string", Help: "Cursor."},
		{Name: "state", Flag: "state", In: "query", Type: "array", Items: "string", Enum: []string{"up", "down"}, Explode: true, Help: "State."},
		{Name: "fields", Flag: "fields", In: "query", Type: "array", Items: "string", Help: "Fields."},
		{Name: "like", Flag: "like", In: "query", Type: "boolean", Help: "Substring match."},
	},
}

var getOp = Operation{
	Group: "results", Name: "get-monitor", ID: "getMonitorResult", Method: "GET",
	Path:    "/monitor/{monitorId}/result/{id}",
	Summary: "Get one result.",
	Args:    []Arg{{Name: "monitorId", Help: "The monitor."}, {Name: "id", Help: "The result."}},
}

var createOp = Operation{
	Group: "monitors", Name: "create", ID: "createMonitor", Method: "POST", Path: "/monitor",
	Summary: "Create a monitor.", Body: true, BodyRequired: true,
}

func TestQueryParametersReachTheWire(t *testing.T) {
	var got seen
	server := serve(t, &got, `{"data":[],"hasMore":false,"nextCursor":null}`)
	if _, err := run(t, listOp, server, "", "--limit", "5", "--state", "up", "--state", "down", "--fields", "id,name", "--like"); err != nil {
		t.Fatal(err)
	}
	// An exploded array repeats the key; a non-exploded one is joined
	// with commas, which is what the document asked for each of these.
	want := "fields=id%2Cname&like=true&limit=5&state=up&state=down"
	if got.Query != want {
		t.Errorf("query = %q, want %q", got.Query, want)
	}
	if got.Method != http.MethodGet || got.Path != "/monitor" {
		t.Errorf("sent %s %s, want GET /monitor", got.Method, got.Path)
	}
}

func TestEnumIsRefusedBeforeTheCall(t *testing.T) {
	var got seen
	server := serve(t, &got, `{}`)
	_, err := run(t, listOp, server, "", "--state", "sideways")
	if err == nil {
		t.Fatal("a value outside the enum was accepted")
	}
	if ExitCode(err) != ExitUsage {
		t.Errorf("exit code %d, want %d", ExitCode(err), ExitUsage)
	}
	if got.Method != "" {
		t.Error("the call was made anyway")
	}
}

func TestPathArgumentsAreSubstituted(t *testing.T) {
	var got seen
	server := serve(t, &got, `{"id":"r-1"}`)
	if _, err := run(t, getOp, server, "", "m-1", "r-1"); err != nil {
		t.Fatal(err)
	}
	if got.Path != "/monitor/m-1/result/r-1" {
		t.Errorf("path = %q, want /monitor/m-1/result/r-1", got.Path)
	}
}

func TestMissingArgumentIsAUsageFailure(t *testing.T) {
	var got seen
	server := serve(t, &got, `{}`)
	_, err := run(t, getOp, server, "", "m-1")
	if err == nil {
		t.Fatal("a missing argument was accepted")
	}
	if got.Method != "" {
		t.Error("the call was made anyway")
	}
}

func TestBodyFromJSONAndSet(t *testing.T) {
	var got seen
	server := serve(t, &got, `{"id":"m-new"}`)
	if _, err := run(t, createOp, server, "", "--json", `{"type":"http"}`, "--set", "name=api", "--set", "settings.interval=5", "--set", "tags.0=prod"); err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(got.Body), &body); err != nil {
		t.Fatalf("the body is not JSON: %v (%s)", err, got.Body)
	}
	if body["type"] != "http" || body["name"] != "api" {
		t.Errorf("body = %s", got.Body)
	}
	settings, _ := body["settings"].(map[string]any)
	if settings == nil || settings["interval"] != float64(5) {
		t.Errorf("--set did not write a nested number: %s", got.Body)
	}
	tags, _ := body["tags"].([]any)
	if len(tags) != 1 || tags[0] != "prod" {
		t.Errorf("--set did not write an array element: %s", got.Body)
	}
	if got.Header.Get("Idempotency-Key") == "" {
		t.Error("the SDK did not put an idempotency key on the write")
	}
}

func TestBodyFromStdin(t *testing.T) {
	var got seen
	server := serve(t, &got, `{}`)
	if _, err := run(t, createOp, server, `{"name":"piped"}`, "--json", "-"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.Body, "piped") {
		t.Errorf("body = %q", got.Body)
	}
}

func TestARequiredBodyIsNotOptional(t *testing.T) {
	var got seen
	server := serve(t, &got, `{}`)
	_, err := run(t, createOp, server, "")
	if err == nil || ExitCode(err) != ExitUsage {
		t.Fatalf("err = %v, exit %d; want a usage failure", err, ExitCode(err))
	}
}

func TestQueryTwinFold(t *testing.T) {
	var got seen
	server := serve(t, &got, `{"data":[],"hasMore":false}`)
	if _, err := run(t, listOp, server, "", "--query", `{"state":["down"]}`, "--limit", "2"); err != nil {
		t.Fatal(err)
	}
	if got.Method != http.MethodPost || got.Path != "/monitor/q" {
		t.Errorf("sent %s %s, want POST /monitor/q", got.Method, got.Path)
	}
	if got.Query != "" {
		t.Errorf("the query string should be empty, got %q", got.Query)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(got.Body), &body); err != nil {
		t.Fatal(err)
	}
	if body["limit"] != float64(2) {
		t.Errorf("--limit did not fold into the body: %s", got.Body)
	}
	// A /q twin is a read, so it carries no idempotency key.
	if got.Header.Get("Idempotency-Key") != "" {
		t.Error("a body query was sent with an idempotency key")
	}
}

func TestQueryTwinRefusesQueryStringFilters(t *testing.T) {
	var got seen
	server := serve(t, &got, `{}`)
	_, err := run(t, listOp, server, "", "--query", `{}`, "--state", "up")
	if err == nil || ExitCode(err) != ExitUsage {
		t.Fatalf("err = %v; want a usage failure naming --state", err)
	}
}

func TestAllWalksEveryPage(t *testing.T) {
	pages := map[string]string{
		"":   `{"data":[{"id":"a"},{"id":"b"}],"nextCursor":"c1","hasMore":true,"count":{"matched":3}}`,
		"c1": `{"data":[{"id":"c"}],"nextCursor":null,"hasMore":false}`,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, pages[r.URL.Query().Get("cursor")])
	}))
	t.Cleanup(server.Close)

	out := &bytes.Buffer{}
	opts := &Options{Out: out, Err: io.Discard, In: strings.NewReader(""), BaseURL: server.URL, Token: "t", OutputFmt: "json", All: true}
	t.Setenv("HT_CONFIG_DIR", t.TempDir())
	cmd := OperationCommand(listOp, opts)
	cmd.SetArgs(nil)
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}

	var envelope struct {
		Data       []map[string]string `json:"data"`
		HasMore    bool                `json:"hasMore"`
		NextCursor *string             `json:"nextCursor"`
		Count      map[string]int      `json:"count"`
	}
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Data) != 3 {
		t.Errorf("got %d rows, want 3", len(envelope.Data))
	}
	if envelope.HasMore || envelope.NextCursor != nil {
		t.Error("the merged envelope still claims another page")
	}
	if envelope.Count["matched"] != 3 {
		t.Error("the envelope's count block was dropped")
	}
}

func TestAllRefusesAnUnpagedOperation(t *testing.T) {
	var got seen
	server := serve(t, &got, `{}`)
	out := &bytes.Buffer{}
	opts := &Options{Out: out, Err: io.Discard, In: strings.NewReader(""), BaseURL: server.URL, Token: "t", OutputFmt: "json", All: true}
	t.Setenv("HT_CONFIG_DIR", t.TempDir())
	cmd := OperationCommand(getOp, opts)
	cmd.SetArgs([]string{"m-1", "r-1"})
	err := cmd.ExecuteContext(context.Background())
	if err == nil || ExitCode(err) != ExitUsage {
		t.Fatalf("err = %v; want a usage failure", err)
	}
}

func TestUseLine(t *testing.T) {
	if got := getOp.Use(); got != "get-monitor <monitor-id> <id>" {
		t.Errorf("Use() = %q", got)
	}
	if got := listOp.Use(); got != "list" {
		t.Errorf("Use() = %q", got)
	}
}
