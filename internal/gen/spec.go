package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// document is the part of an OpenAPI document this generator reads.
type document struct {
	Info struct {
		Title   string `json:"title"`
		Version string `json:"version"`
	} `json:"info"`
	Paths      map[string]map[string]json.RawMessage `json:"paths"`
	Components struct {
		Parameters map[string]specParam `json:"parameters"`
	} `json:"components"`
}

type specOperation struct {
	OperationID string          `json:"operationId"`
	Tags        []string        `json:"tags"`
	Summary     string          `json:"summary"`
	Description string          `json:"description"`
	Deprecated  bool            `json:"deprecated"`
	Parameters  []specParamRef  `json:"parameters"`
	RequestBody *specBody       `json:"requestBody"`
	Responses   map[string]spec `json:"responses"`
}

type spec struct {
	Content map[string]json.RawMessage `json:"content"`
}

type specBody struct {
	Required bool                       `json:"required"`
	Content  map[string]json.RawMessage `json:"content"`
}

type specParamRef struct {
	Ref string `json:"$ref"`
	specParam
}

type specParam struct {
	Name        string      `json:"name"`
	In          string      `json:"in"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Explode     *bool       `json:"explode"`
	Schema      *specSchema `json:"schema"`
}

type specSchema struct {
	Type  string      `json:"type"`
	Enum  []any       `json:"enum"`
	Items *specSchema `json:"items"`
}

// methods are the HTTP methods a path item may declare, in the order the
// generator walks them.
var methods = []string{"get", "post", "put", "patch", "delete", "head", "options"}

// load reads and parses an OpenAPI document.
func load(path string) (*document, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc document
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if len(doc.Paths) == 0 {
		return nil, fmt.Errorf("%s: the document declares no paths", path)
	}
	return &doc, nil
}

// endpoint is one operation of the document, with its path and method.
type endpoint struct {
	Path   string
	Method string
	Op     specOperation
}

// endpoints lists every operation, sorted by path then method, so the
// generated files do not move when a map iterates differently.
func (d *document) endpoints() ([]endpoint, error) {
	var out []endpoint
	paths := make([]string, 0, len(d.Paths))
	for path := range d.Paths {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		item := d.Paths[path]
		for _, method := range methods {
			raw, ok := item[method]
			if !ok {
				continue
			}
			var op specOperation
			if err := json.Unmarshal(raw, &op); err != nil {
				return nil, fmt.Errorf("%s %s: %w", strings.ToUpper(method), path, err)
			}
			if op.OperationID == "" {
				return nil, fmt.Errorf("%s %s: the operation declares no operationId", strings.ToUpper(method), path)
			}
			if len(op.Tags) == 0 {
				return nil, fmt.Errorf("%s: the operation declares no tag", op.OperationID)
			}
			out = append(out, endpoint{Path: path, Method: strings.ToUpper(method), Op: op})
		}
	}
	return out, nil
}

// resolve dereferences a parameter that the document declared in
// components.
func (d *document) resolve(ref specParamRef) (specParam, error) {
	if ref.Ref == "" {
		return ref.specParam, nil
	}
	name := ref.Ref[strings.LastIndexByte(ref.Ref, '/')+1:]
	param, ok := d.Components.Parameters[name]
	if !ok {
		return specParam{}, fmt.Errorf("the document references an undeclared parameter %q", ref.Ref)
	}
	return param, nil
}

// jsonBody reports whether the operation takes a JSON request body, and
// whether that body is required.
func (op specOperation) jsonBody() (bool, bool) {
	if op.RequestBody == nil {
		return false, false
	}
	for media := range op.RequestBody.Content {
		if strings.Contains(media, "json") {
			return true, op.RequestBody.Required
		}
	}
	return false, false
}

// binaryAnswer reports a success answer that is not JSON: a rendered
// report, an image.
func (op specOperation) binaryAnswer() bool {
	for status, response := range op.Responses {
		if !strings.HasPrefix(status, "2") {
			continue
		}
		for media := range response.Content {
			if !strings.Contains(media, "json") {
				return true
			}
		}
	}
	return false
}

// enumOf reads a closed value set off a schema, string members only.
func enumOf(schema *specSchema) []string {
	if schema == nil {
		return nil
	}
	source := schema
	if schema.Type == "array" && schema.Items != nil {
		source = schema.Items
	}
	out := make([]string, 0, len(source.Enum))
	for _, value := range source.Enum {
		text, ok := value.(string)
		if !ok {
			return nil
		}
		out = append(out, text)
	}
	return out
}

// itemsType is the element type of an array schema.
func itemsType(schema *specSchema) string {
	if schema == nil || schema.Items == nil {
		return "string"
	}
	if schema.Items.Type == "" {
		return "string"
	}
	return schema.Items.Type
}
