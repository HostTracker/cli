// Package output renders a decoded API answer as JSON, YAML or a table.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Format is one of the shapes an answer is printed in.
type Format string

// The supported formats.
const (
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
	FormatTable Format = "table"
)

// Formats lists the accepted values of --output, in help order.
var Formats = []string{string(FormatJSON), string(FormatYAML), string(FormatTable)}

// ParseFormat validates a --output value.
func ParseFormat(s string) (Format, error) {
	switch Format(strings.ToLower(strings.TrimSpace(s))) {
	case FormatJSON:
		return FormatJSON, nil
	case FormatYAML:
		return FormatYAML, nil
	case FormatTable:
		return FormatTable, nil
	default:
		return "", fmt.Errorf("unknown output format %q (known: %s)", s, strings.Join(Formats, ", "))
	}
}

// IsTerminal reports whether w is an interactive terminal.
func IsTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// Default is the format used when --output was not given: a table on an
// interactive terminal, JSON when the answer is being piped somewhere.
func Default(w io.Writer) Format {
	if IsTerminal(w) {
		return FormatTable
	}
	return FormatJSON
}

// Printer writes answers in one format.
type Printer struct {
	Out    io.Writer
	Format Format
}

// Print renders doc. Anything that is not already generic JSON is passed
// through encoding/json first, so a typed SDK view and a raw answer are
// rendered by exactly the same code.
func (p Printer) Print(doc any) error {
	value, err := Generic(doc)
	if err != nil {
		return err
	}
	switch p.Format {
	case FormatYAML:
		enc := yaml.NewEncoder(p.Out)
		enc.SetIndent(2)
		if err := enc.Encode(forYAML(value)); err != nil {
			return err
		}
		return enc.Close()
	case FormatTable:
		return Table(p.Out, value)
	default:
		enc := json.NewEncoder(p.Out)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		return enc.Encode(value)
	}
}

// Generic turns any value into the map/slice/scalar shape the renderers
// work on. A json.RawMessage is decoded rather than re-encoded.
func Generic(doc any) (any, error) {
	switch v := doc.(type) {
	case nil:
		return nil, nil
	case json.RawMessage:
		return decode(v)
	case []byte:
		return decode(v)
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	return decode(raw)
}

// forYAML converts the json.Number values the decoder keeps into the Go
// numbers the YAML encoder writes unquoted. Decoding keeps them as
// json.Number so a 64-bit id or a Unix second survives the round trip
// without passing through a float.
func forYAML(value any) any {
	switch v := value.(type) {
	case json.Number:
		if n, err := v.Int64(); err == nil {
			return n
		}
		if f, err := v.Float64(); err == nil {
			return f
		}
		return v.String()
	case map[string]any:
		out := make(map[string]any, len(v))
		for name, member := range v {
			out[name] = forYAML(member)
		}
		return out
	case []any:
		out := make([]any, 0, len(v))
		for _, item := range v {
			out = append(out, forYAML(item))
		}
		return out
	default:
		return value
	}
}

func decode(raw []byte) (any, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return nil, nil
	}
	dec := json.NewDecoder(strings.NewReader(trimmed))
	dec.UseNumber()
	var value any
	if err := dec.Decode(&value); err != nil {
		return nil, fmt.Errorf("the answer is not JSON: %w", err)
	}
	return value, nil
}
