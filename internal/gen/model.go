package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// command is one generated operation command.
type command struct {
	Group        string
	Name         string
	ID           string
	Method       string
	Path         string
	Summary      string
	Description  string
	Args         []argModel
	Params       []paramModel
	Body         bool
	BodyRequired bool
	QueryPath    string
	Paged        bool
	Binary       bool
	Deprecated   string
}

type argModel struct {
	Name string
	Help string
}

type paramModel struct {
	Name     string
	Flag     string
	In       string
	Type     string
	Items    string
	Enum     []string
	Help     string
	Required bool
	Explode  bool
}

// group is one generated command group.
type group struct {
	Name     string
	Tag      string
	Short    string
	Aliases  []string
	Commands []command
}

var pathParamRe = regexp.MustCompile(`\{([^}]+)\}`)

// build turns a document into the groups the generator emits.
func build(doc *document) ([]group, error) {
	endpoints, err := doc.endpoints()
	if err != nil {
		return nil, err
	}

	// A POST <path>/q is the body-query twin of the GET at <path>: the
	// same read with its filters in a body. It is folded into the GET's
	// command as --query rather than emitted as a command of its own.
	getByPath := map[string]bool{}
	for _, e := range endpoints {
		if e.Method == "GET" {
			getByPath[e.Path] = true
		}
	}
	twinOf := map[string]string{}
	for _, e := range endpoints {
		if e.Method != "POST" || !strings.HasSuffix(e.Path, "/q") {
			continue
		}
		base := strings.TrimSuffix(e.Path, "/q")
		if getByPath[base] {
			twinOf[base] = e.Path
		}
	}

	byTag := map[string]*group{}
	var order []string
	for _, e := range endpoints {
		if e.Method == "POST" && strings.HasSuffix(e.Path, "/q") && getByPath[strings.TrimSuffix(e.Path, "/q")] {
			continue
		}
		tag := e.Op.Tags[0]
		g, ok := byTag[tag]
		if !ok {
			short, known := groupShort[tag]
			if !known {
				short = tag + " operations"
			}
			name := groupName(tag)
			g = &group{Name: name, Tag: tag, Short: short, Aliases: groupAliases[name]}
			byTag[tag] = g
			order = append(order, tag)
		}
		// Only the GET at a path folds in its body-query twin; a POST
		// or PATCH sharing the same path is a different operation.
		queryPath := ""
		if e.Method == "GET" {
			queryPath = twinOf[e.Path]
		}
		cmd, err := buildCommand(doc, e, g, queryPath)
		if err != nil {
			return nil, err
		}
		g.Commands = append(g.Commands, cmd)
	}

	sort.Strings(order)
	out := make([]group, 0, len(order))
	for _, tag := range order {
		g := byTag[tag]
		sort.Slice(g.Commands, func(i, j int) bool { return g.Commands[i].Name < g.Commands[j].Name })
		for i := 1; i < len(g.Commands); i++ {
			if g.Commands[i].Name == g.Commands[i-1].Name {
				return nil, fmt.Errorf("%s and %s both map to `ht %s %s`",
					g.Commands[i-1].ID, g.Commands[i].ID, g.Name, g.Commands[i].Name)
			}
		}
		out = append(out, *g)
	}
	return out, nil
}

func buildCommand(doc *document, e endpoint, g *group, queryPath string) (command, error) {
	body, bodyRequired := e.Op.jsonBody()
	cmd := command{
		Group:        g.Name,
		Name:         commandName(e.Op.OperationID, g.Tag),
		ID:           e.Op.OperationID,
		Method:       e.Method,
		Path:         e.Path,
		Summary:      oneLine(e.Op.Summary),
		Description:  paragraph(e.Op.Description),
		Body:         body,
		BodyRequired: bodyRequired,
		QueryPath:    queryPath,
		Binary:       e.Op.binaryAnswer(),
	}
	if e.Op.Deprecated {
		cmd.Deprecated = "this operation is deprecated"
	}
	if cmd.Summary == "" {
		cmd.Summary = firstSentence(cmd.Description)
	}

	described := map[string]string{}
	for _, ref := range e.Op.Parameters {
		param, err := doc.resolve(ref)
		if err != nil {
			return cmd, fmt.Errorf("%s: %w", e.Op.OperationID, err)
		}
		if param.In == "path" {
			described[param.Name] = oneLine(firstSentence(param.Description))
			continue
		}
		if param.In != "query" && param.In != "header" {
			continue
		}
		cmd.Params = append(cmd.Params, buildParam(param))
	}
	sort.Slice(cmd.Params, func(i, j int) bool { return cmd.Params[i].Flag < cmd.Params[j].Flag })
	for i := 1; i < len(cmd.Params); i++ {
		if cmd.Params[i].Flag == cmd.Params[i-1].Flag {
			return cmd, fmt.Errorf("%s: parameters %q and %q both spell --%s",
				e.Op.OperationID, cmd.Params[i-1].Name, cmd.Params[i].Name, cmd.Params[i].Flag)
		}
	}

	for _, name := range pathParams(e.Path) {
		cmd.Args = append(cmd.Args, argModel{Name: name, Help: argHelp(e.Path, name, described[name])})
	}

	for _, param := range cmd.Params {
		if param.In == "query" && param.Name == "cursor" {
			cmd.Paged = e.Method == "GET"
		}
	}
	return cmd, nil
}

func buildParam(param specParam) paramModel {
	kind := "string"
	if param.Schema != nil && param.Schema.Type != "" {
		kind = param.Schema.Type
	}
	out := paramModel{
		Name:     param.Name,
		Flag:     flagName(param.Name),
		In:       param.In,
		Type:     kind,
		Enum:     enumOf(param.Schema),
		Help:     clip(plain(firstSentence(param.Description)), 150),
		Required: param.Required,
		// A form-style array explodes by default; the document says so
		// explicitly where it wants the comma-joined spelling instead.
		Explode: param.Explode == nil || *param.Explode,
	}
	if kind == "array" {
		out.Items = itemsType(param.Schema)
	}
	return out
}

// plain strips the markdown a help line cannot carry. The backticks
// matter most: pflag reads the first backquoted word of a usage string as
// the flag's value placeholder, so a quoted example would rename the
// flag's type in the help output.
func plain(text string) string {
	return oneLine(strings.NewReplacer("`", "", "**", "", "*", "").Replace(text))
}

// genericPathHelp is what the document says about most path parameters.
// It tells a reader nothing a command line does not already show, so the
// generator writes its own from the path instead.
const genericPathHelp = "See this operation's description for how this parameter narrows the result."

// segmentNames spells a path segment as the noun a reader would use.
var segmentNames = map[string]string{
	"statuspage": "status page",
	"q":          "",
}

// argNames spells a path parameter whose name alone reads as jargon.
var argNames = map[string]string{
	"dbId": "the federation the check was registered in",
}

// pathParams lists the path template's parameters in path order.
func pathParams(path string) []string {
	var out []string
	for _, match := range pathParamRe.FindAllStringSubmatch(path, -1) {
		out = append(out, match[1])
	}
	return out
}

// argHelp says what a positional argument identifies. The document's own
// text is used when it says something; otherwise the name and the path
// do, since <id> under /monitor/{id} is the monitor's.
func argHelp(path, name, described string) string {
	if described != "" && described != genericPathHelp {
		return described
	}
	if spelled, known := argNames[name]; known {
		return spelled
	}
	if prefix, found := strings.CutSuffix(name, "Id"); found && prefix != "" {
		return "the " + strings.Join(words(prefix), " ") + " id"
	}
	if name == "id" {
		if noun := owningSegment(path, name); noun != "" {
			return "the " + noun + " id"
		}
	}
	return "the " + strings.Join(words(name), " ")
}

// owningSegment is the last literal segment of the path before the named
// parameter: the resource the id belongs to.
func owningSegment(path, name string) string {
	noun := ""
	for _, segment := range strings.Split(strings.Trim(path, "/"), "/") {
		if segment == "{"+name+"}" {
			return noun
		}
		if strings.HasPrefix(segment, "{") {
			continue
		}
		if spelled, known := segmentNames[segment]; known {
			noun = spelled
			continue
		}
		noun = singular(segment)
	}
	return noun
}

// oneLine collapses whitespace so a description fits a help line.
func oneLine(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

// paragraph keeps the first paragraph of a description, which is the part
// that says what the operation does; the rest is reference detail the
// published documentation carries.
func paragraph(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	if i := strings.Index(text, "\n\n"); i > 0 {
		text = text[:i]
	}
	return oneLine(text)
}

// minHelpLength is how short a lead sentence may be before the next one
// is taken as well. "Optional." on its own says nothing.
const minHelpLength = 24

// firstSentence is the opening of a description, grown past a lead too
// short to mean anything.
func firstSentence(text string) string {
	text = oneLine(text)
	for i := 0; i < len(text)-1; i++ {
		if text[i] == '.' && text[i+1] == ' ' && i+1 >= minHelpLength {
			return text[:i+1]
		}
	}
	return text
}

// clip shortens a help line at a word boundary.
func clip(text string, limit int) string {
	if len(text) <= limit {
		return text
	}
	cut := strings.LastIndexByte(text[:limit], ' ')
	if cut <= 0 {
		cut = limit
	}
	return strings.TrimRight(text[:cut], " ,;:") + "..."
}
