package htcli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildBodyReadings(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "body.json")
	if err := os.WriteFile(file, []byte(`{"from":"file"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name   string
		source string
		stdin  string
		want   string
	}{
		{"inline", `{"from":"inline"}`, "", "inline"},
		{"file", "@" + file, "", "file"},
		{"stdin", "-", `{"from":"stdin"}`, "stdin"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			body, err := BuildBody(c.source, nil, strings.NewReader(c.stdin))
			if err != nil {
				t.Fatal(err)
			}
			var document map[string]string
			if err := json.Unmarshal(body, &document); err != nil {
				t.Fatal(err)
			}
			if document["from"] != c.want {
				t.Errorf("read %q, want %q", document["from"], c.want)
			}
		})
	}
}

func TestBuildBodySetValues(t *testing.T) {
	body, err := BuildBody("", []string{
		"name=api",
		"enabled=true",
		"interval=5",
		"ratio=0.5",
		"parent=null",
		"quoted=\"5\"",
		"nested.deep.value=x",
		"list.1=second",
		"raw={\"a\":1}",
	}, strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(body, &document); err != nil {
		t.Fatal(err)
	}
	if document["name"] != "api" {
		t.Errorf("name = %v", document["name"])
	}
	if document["enabled"] != true {
		t.Errorf("enabled = %v, want the boolean", document["enabled"])
	}
	if document["interval"] != float64(5) {
		t.Errorf("interval = %v, want the number", document["interval"])
	}
	if document["ratio"] != 0.5 {
		t.Errorf("ratio = %v", document["ratio"])
	}
	if _, present := document["parent"]; !present || document["parent"] != nil {
		t.Errorf("parent = %v, want an explicit null", document["parent"])
	}
	if document["quoted"] != "5" {
		t.Errorf("quoted = %v, want the string", document["quoted"])
	}
	nested := document["nested"].(map[string]any)["deep"].(map[string]any)
	if nested["value"] != "x" {
		t.Errorf("nested = %v", document["nested"])
	}
	list := document["list"].([]any)
	if len(list) != 2 || list[0] != nil || list[1] != "second" {
		t.Errorf("list = %v, want a grown array", list)
	}
	if document["raw"].(map[string]any)["a"] != float64(1) {
		t.Errorf("raw = %v, want the parsed object", document["raw"])
	}
}

func TestBuildBodyRefusals(t *testing.T) {
	if _, err := BuildBody("{not json", nil, strings.NewReader("")); err == nil {
		t.Error("invalid JSON was accepted")
	}
	if _, err := BuildBody("", []string{"noequals"}, strings.NewReader("")); err == nil {
		t.Error("a --set without = was accepted")
	}
	if _, err := BuildBody(`{"name":"x"}`, []string{"name.deep=1"}, strings.NewReader("")); err == nil {
		t.Error("writing through a string was accepted")
	}
	if body, err := BuildBody("", nil, strings.NewReader("")); err != nil || body != nil {
		t.Errorf("an absent body should stay absent, got %q, %v", body, err)
	}
}
