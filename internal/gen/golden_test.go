package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the golden files")

// TestGolden runs the generator over a three-operation document and
// compares the emitted source with the recorded one. It is what catches a
// naming or flag-mapping change that was not meant.
func TestGolden(t *testing.T) {
	doc, err := load(filepath.Join("testdata", "spec.json"))
	if err != nil {
		t.Fatal(err)
	}
	groups, err := build(doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Fatalf("got %d groups, want 1", len(groups))
	}
	if got := len(groups[0].Commands); got != 3 {
		t.Fatalf("got %d commands, want 3 (the /q twin folds into list)", got)
	}

	source, err := emit("gen", groups[0])
	if err != nil {
		t.Fatal(err)
	}
	golden := filepath.Join("testdata", "golden", "monitors.go.golden")
	if *update {
		if err := os.WriteFile(golden, source, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	// A Windows checkout may convert the golden file to CRLF; compare on LF.
	want = bytes.ReplaceAll(want, []byte("\r\n"), []byte("\n"))
	if string(source) != string(want) {
		t.Errorf("the generated source moved.\n--- got ---\n%s\n--- want ---\n%s", source, want)
	}
}

// TestFoldsQueryTwin proves the /q twin becomes a flag on the GET it
// belongs to, and not a command of its own.
func TestFoldsQueryTwin(t *testing.T) {
	doc, err := load(filepath.Join("testdata", "spec.json"))
	if err != nil {
		t.Fatal(err)
	}
	groups, err := build(doc)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]command{}
	for _, cmd := range groups[0].Commands {
		byName[cmd.Name] = cmd
	}
	if _, found := byName["query"]; found {
		t.Error("the /q twin was emitted as its own command")
	}
	if got := byName["list"].QueryPath; got != "/monitor/q" {
		t.Errorf("list.QueryPath = %q, want /monitor/q", got)
	}
	if !byName["list"].Paged {
		t.Error("list is not marked paged, so --all would refuse it")
	}
	// A POST sharing the path must not inherit the twin.
	if got := byName["create"].QueryPath; got != "" {
		t.Errorf("create.QueryPath = %q, want empty", got)
	}
	if !byName["create"].BodyRequired {
		t.Error("create should require a body")
	}
	if got := len(byName["get"].Args); got != 1 {
		t.Errorf("get takes %d positional arguments, want 1", got)
	}
}
