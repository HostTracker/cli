package htcli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// ReadJSONSource reads one of the three spellings a JSON argument takes:
//
//	'{"name":"api"}'   inline
//	@body.json         a file
//	-                  standard input
func ReadJSONSource(source string, stdin io.Reader) ([]byte, error) {
	trimmed := strings.TrimSpace(source)
	switch {
	case trimmed == "":
		return nil, nil
	case trimmed == "-":
		raw, err := io.ReadAll(stdin)
		if err != nil {
			return nil, fmt.Errorf("reading standard input: %w", err)
		}
		return raw, nil
	case strings.HasPrefix(trimmed, "@"):
		path := trimmed[1:]
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return raw, nil
	default:
		return []byte(source), nil
	}
}

// BuildBody assembles a request body from --json and --set.
//
// --json supplies the document; each --set then writes one value into it
// at a dotted path, so a small change needs no file:
//
//	--set name=api --set settings.interval=5 --set tags.0=prod
//
// A value is read as JSON when it parses as one (a number, true, false,
// null, an object, an array) and as a plain string otherwise. Wrapping it
// in quotes forces the string reading: --set name='"5"'.
func BuildBody(jsonSource string, sets []string, stdin io.Reader) ([]byte, error) {
	raw, err := ReadJSONSource(jsonSource, stdin)
	if err != nil {
		return nil, &UsageError{Err: err}
	}
	if len(sets) == 0 {
		if len(raw) == 0 {
			return nil, nil
		}
		if !json.Valid(raw) {
			return nil, Usagef("--json is not valid JSON")
		}
		return raw, nil
	}

	var document any
	if len(strings.TrimSpace(string(raw))) > 0 {
		if err := json.Unmarshal(raw, &document); err != nil {
			return nil, Usagef("--json is not valid JSON: %v", err)
		}
	}
	for _, assignment := range sets {
		key, value, found := strings.Cut(assignment, "=")
		if !found {
			return nil, Usagef("--set wants key=value, got %q", assignment)
		}
		document, err = setPath(document, strings.Split(key, "."), parseSetValue(value))
		if err != nil {
			return nil, &UsageError{Err: err}
		}
	}
	return json.Marshal(document)
}

// parseSetValue reads a --set value as JSON when it is one, and as a
// plain string otherwise.
func parseSetValue(raw string) any {
	trimmed := strings.TrimSpace(raw)
	switch trimmed {
	case "":
		return ""
	case "null":
		return nil
	case "true":
		return true
	case "false":
		return false
	}
	if _, err := strconv.ParseFloat(trimmed, 64); err == nil {
		var number any
		if json.Unmarshal([]byte(trimmed), &number) == nil {
			return number
		}
	}
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, `"`) {
		var value any
		if json.Unmarshal([]byte(trimmed), &value) == nil {
			return value
		}
	}
	return raw
}

// setPath writes value into document at path, growing objects and arrays
// on the way. A numeric segment addresses an array index; anything else
// addresses an object member.
func setPath(document any, path []string, value any) (any, error) {
	if len(path) == 0 {
		return value, nil
	}
	segment := path[0]
	if index, err := strconv.Atoi(segment); err == nil && index >= 0 {
		list, ok := document.([]any)
		if !ok && document != nil {
			return nil, fmt.Errorf("--set: %q addresses an array but the document holds %s there", segment, kindOf(document))
		}
		for len(list) <= index {
			list = append(list, nil)
		}
		child, err := setPath(list[index], path[1:], value)
		if err != nil {
			return nil, err
		}
		list[index] = child
		return list, nil
	}

	object, ok := document.(map[string]any)
	if !ok {
		if document != nil {
			return nil, fmt.Errorf("--set: %q addresses an object member but the document holds %s there", segment, kindOf(document))
		}
		object = map[string]any{}
	}
	child, err := setPath(object[segment], path[1:], value)
	if err != nil {
		return nil, err
	}
	object[segment] = child
	return object, nil
}

func kindOf(value any) string {
	switch value.(type) {
	case map[string]any:
		return "an object"
	case []any:
		return "an array"
	case string:
		return "a string"
	case bool:
		return "a boolean"
	case float64, json.Number:
		return "a number"
	default:
		return "a value"
	}
}
