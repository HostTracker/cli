// Package htcli is the machinery the generated commands and the
// hand-written ones share: the operation descriptor, the flag surface
// built from it, the request runner on top of the Go SDK, and the exit
// codes.
package htcli

import (
	"fmt"
	"sort"
	"strings"
)

// Param is one query or header parameter of an operation.
type Param struct {
	// Name is the wire name, e.g. "monitorId".
	Name string
	// Flag is the command-line spelling, e.g. "monitor-id".
	Flag string
	// In is "query" or "header".
	In string
	// Type is "string", "integer", "number", "boolean" or "array".
	Type string
	// Items is the element type when Type is "array".
	Items string
	// Enum lists the accepted values, when the parameter has a closed set.
	Enum []string
	// Explode reports the array serialisation the document asked for:
	// true repeats the key once per element, false joins the elements
	// with commas into a single key.
	Explode bool
	// Help is the one-line flag description.
	Help string
	// Required reports a parameter the API refuses the call without.
	Required bool
}

// Arg is one path parameter, taken as a positional argument.
type Arg struct {
	// Name is the wire name, e.g. "monitorId".
	Name string
	// Help is what the argument identifies.
	Help string
}

// Operation is one API operation, as a command.
type Operation struct {
	// Group is the command group, e.g. "monitors".
	Group string
	// Name is the command name inside the group, e.g. "bulk-create".
	Name string
	// ID is the operationId, so a command can be traced to the document.
	ID string
	// Method and Path address the operation.
	Method string
	Path   string
	// Summary is the one-line help.
	Summary string
	// Description is the long help.
	Description string
	// Args are the path parameters, in path order.
	Args []Arg
	// Params are the query and header parameters.
	Params []Param
	// Body reports that the operation takes a JSON request body.
	Body bool
	// BodyRequired reports that the body is not optional.
	BodyRequired bool
	// QueryPath is the POST <path>/q body-query twin, folded into this
	// command as --query and --query-file. Empty when there is none.
	QueryPath string
	// Paged reports a collection that --all may walk to exhaustion.
	Paged bool
	// Binary reports an answer that is not JSON (a report download).
	Binary bool
	// Deprecated carries the replacement when the operation is going away.
	Deprecated string
}

// Group is a command group, one per tag of the document.
type Group struct {
	// Name is the group command, e.g. "status-pages".
	Name string
	// Tag is the tag it was built from, e.g. "StatusPages".
	Tag string
	// Short is the one-line group help.
	Short string
	// Aliases are alternative spellings of the group name.
	Aliases []string
}

var (
	groups     []Group
	operations = map[string][]Operation{}
)

// RegisterGroup declares a command group. The generated files call it
// from their init functions.
func RegisterGroup(g Group) {
	for _, existing := range groups {
		if existing.Name == g.Name {
			panic(fmt.Sprintf("htcli: group %q registered twice", g.Name))
		}
	}
	groups = append(groups, g)
}

// Register declares one operation command.
func Register(op Operation) {
	for _, existing := range operations[op.Group] {
		if existing.Name == op.Name {
			panic(fmt.Sprintf("htcli: command %q registered twice in group %q", op.Name, op.Group))
		}
	}
	operations[op.Group] = append(operations[op.Group], op)
}

// Groups lists the registered groups in alphabetical order.
func Groups() []Group {
	out := append([]Group(nil), groups...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// OperationsOf lists a group's operations in alphabetical order.
func OperationsOf(group string) []Operation {
	out := append([]Operation(nil), operations[group]...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Count reports how many operation commands are registered.
func Count() int {
	n := 0
	for _, ops := range operations {
		n += len(ops)
	}
	return n
}

// Use is the cobra usage line: the command name followed by its
// positional arguments.
func (op Operation) Use() string {
	parts := make([]string, 0, len(op.Args)+1)
	parts = append(parts, op.Name)
	for _, arg := range op.Args {
		parts = append(parts, "<"+kebab(arg.Name)+">")
	}
	return strings.Join(parts, " ")
}

// kebab spells a camelCase wire name as a command-line word.
func kebab(name string) string {
	var b strings.Builder
	for i, r := range name {
		switch {
		case r >= 'A' && r <= 'Z':
			if i > 0 && !strings.HasSuffix(b.String(), "-") {
				b.WriteByte('-')
			}
			b.WriteRune(r - 'A' + 'a')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
