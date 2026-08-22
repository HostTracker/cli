package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestParseFormat(t *testing.T) {
	for _, name := range []string{"json", "YAML", " table "} {
		if _, err := ParseFormat(name); err != nil {
			t.Errorf("ParseFormat(%q): %v", name, err)
		}
	}
	if _, err := ParseFormat("xml"); err == nil {
		t.Error("an unknown format was accepted")
	}
}

func TestPrintJSON(t *testing.T) {
	var b bytes.Buffer
	doc := json.RawMessage(`{"b":2,"a":1}`)
	if err := (Printer{Out: &b, Format: FormatJSON}).Print(doc); err != nil {
		t.Fatal(err)
	}
	// Members come back sorted, so a diff between two runs is stable.
	if got := b.String(); !strings.HasPrefix(got, "{\n  \"a\": 1,\n  \"b\": 2\n}") {
		t.Errorf("got %q", got)
	}
}

func TestPrintYAML(t *testing.T) {
	var b bytes.Buffer
	if err := (Printer{Out: &b, Format: FormatYAML}).Print(json.RawMessage(`{"name":"api","interval":5}`)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "name: api") || !strings.Contains(b.String(), "interval: 5") {
		t.Errorf("got %q", b.String())
	}
}

func TestPrintTableOfACollection(t *testing.T) {
	var b bytes.Buffer
	answer := json.RawMessage(`{"data":[
	  {"id":"m-1","name":"one","state":"up","created":1735689600,"settings":{"a":1}},
	  {"id":"m-2","name":"two","state":"down","created":1735689600,"settings":{"a":1}}
	],"hasMore":true,"count":{"matched":9}}`)
	if err := (Printer{Out: &b, Format: FormatTable}).Print(answer); err != nil {
		t.Fatal(err)
	}
	text := b.String()
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if !strings.HasPrefix(lines[0], "ID") {
		t.Errorf("the preferred columns should lead, got %q", lines[0])
	}
	if !strings.Contains(lines[0], "NAME") || !strings.Contains(lines[0], "STATE") {
		t.Errorf("header = %q", lines[0])
	}
	// A nested member is folded, never spilled into a column.
	if strings.Contains(text, `"a":1`) {
		t.Errorf("a nested object leaked into the table:\n%s", text)
	}
	// Unix seconds are rendered as the instants they are.
	if !strings.Contains(text, "2025-01-01T00:00:00Z") {
		t.Errorf("created was not rendered as an instant:\n%s", text)
	}
	if !strings.Contains(text, "9 matched") || !strings.Contains(text, "more rows available") {
		t.Errorf("the envelope footer is missing:\n%s", text)
	}
}

func TestPrintTableOfOneObject(t *testing.T) {
	var b bytes.Buffer
	if err := (Printer{Out: &b, Format: FormatTable}).Print(json.RawMessage(`{"id":"a-1","email":"x@example.com"}`)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "ID") || !strings.Contains(b.String(), "x@example.com") {
		t.Errorf("got %q", b.String())
	}
}

func TestPrintTableOfNoRows(t *testing.T) {
	var b bytes.Buffer
	if err := (Printer{Out: &b, Format: FormatTable}).Print(json.RawMessage(`{"data":[],"hasMore":false}`)); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(b.String()) != "(no rows)" {
		t.Errorf("got %q", b.String())
	}
}

func TestHeaderSpelling(t *testing.T) {
	cases := map[string]string{"id": "ID", "openStat": "OPEN STAT", "nextCursor": "NEXT CURSOR"}
	for input, want := range cases {
		if got := header(input); got != want {
			t.Errorf("header(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestInstantsOnlyWhereTheyBelong(t *testing.T) {
	// A counter under an ...At name must not be mangled into a date.
	if got := cell("attemptsAt", json.Number("3")); got != "3" {
		t.Errorf("cell = %q, want the number untouched", got)
	}
	if got := cell("occurredAt", json.Number("1735689600")); got != "2025-01-01T00:00:00Z" {
		t.Errorf("cell = %q", got)
	}
}

func TestDefaultFormatWhenPiped(t *testing.T) {
	if got := Default(&bytes.Buffer{}); got != FormatJSON {
		t.Errorf("Default = %q, want json when the answer is not going to a terminal", got)
	}
}
