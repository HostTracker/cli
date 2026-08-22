package htcli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Flag names the command layer owns. A generated parameter that would
// collide with one of these is renamed by the generator, so the two sets
// never overlap.
const (
	FlagJSON           = "json"
	FlagSet            = "set"
	FlagQuery          = "query"
	FlagQueryFile      = "query-file"
	FlagIdempotencyKey = "idempotency-key"
)

// AddGroups builds a command for every registered group and operation and
// hangs them off root.
func AddGroups(root *cobra.Command, opts *Options) {
	for _, group := range Groups() {
		root.AddCommand(GroupCommand(group, opts))
	}
}

// GroupCommand builds one group command with its operations under it.
func GroupCommand(group Group, opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     group.Name,
		Aliases: group.Aliases,
		Short:   group.Short,
		Long:    group.Short,
	}
	for _, op := range OperationsOf(group.Name) {
		cmd.AddCommand(OperationCommand(op, opts))
	}
	return cmd
}

// OperationCommand builds the command for one operation: its positional
// arguments from the path, its flags from the parameters and the body.
func OperationCommand(op Operation, opts *Options) *cobra.Command {
	var (
		jsonSource     string
		sets           []string
		querySource    string
		queryFile      string
		idempotencyKey string
	)

	cmd := &cobra.Command{
		Use:        op.Use(),
		Short:      op.Summary,
		Long:       longHelp(op),
		Args:       cobra.ExactArgs(len(op.Args)),
		Deprecated: op.Deprecated,
	}
	flags := cmd.Flags()
	declareParams(flags, op)

	if op.Body {
		flags.StringVar(&jsonSource, FlagJSON, "", "request body: inline JSON, @file, or - for standard input")
		flags.StringArrayVar(&sets, FlagSet, nil, "set one body member: --set name=api --set settings.interval=5 (repeatable)")
	}
	if op.QueryPath != "" {
		flags.StringVar(&querySource, FlagQuery, "", "filter with a body query against "+op.QueryPath+": inline JSON, @file, or -")
		flags.StringVar(&queryFile, FlagQueryFile, "", "filter with a body query read from a file")
	}
	if !hasParam(op, FlagIdempotencyKey) && writes(op.Method) {
		flags.StringVar(&idempotencyKey, FlagIdempotencyKey, "", "replay key for this write (default: a fresh one per call)")
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := opts.Resolve(); err != nil {
			return err
		}
		req, cursorInBody, err := buildRequest(op, args, flags, opts, requestInput{
			jsonSource:     jsonSource,
			sets:           sets,
			querySource:    querySource,
			queryFile:      queryFile,
			idempotencyKey: idempotencyKey,
		})
		if err != nil {
			return err
		}
		if opts.All {
			if !op.Paged {
				return Usagef("--all only applies to a paged collection, and %s %s is not one", op.Group, op.Name)
			}
			answer, err := opts.collect(cmd.Context(), req, cursorInBody)
			if err != nil || answer == nil {
				return err
			}
			return opts.Emit(answer)
		}
		answer, err := opts.Do(cmd.Context(), req)
		if err != nil {
			return err
		}
		return opts.Emit(answer)
	}
	return cmd
}

// requestInput is what the shared body and query flags collected.
type requestInput struct {
	jsonSource     string
	sets           []string
	querySource    string
	queryFile      string
	idempotencyKey string
}

// buildRequest turns the parsed command line into one call, and reports
// whether a --all walk carries its cursor in the body.
func buildRequest(op Operation, args []string, flags *pflag.FlagSet, opts *Options, in requestInput) (Request, bool, error) {
	req := Request{Method: op.Method, Path: op.Path, Query: url.Values{}, Header: http.Header{}}

	useQueryTwin := op.QueryPath != "" && (flags.Changed(FlagQuery) || flags.Changed(FlagQueryFile))
	if useQueryTwin {
		req.Method = http.MethodPost
		req.Path = op.QueryPath
	}

	path, err := fillPath(req.Path, op.Args, args)
	if err != nil {
		return req, false, err
	}
	req.Path = path

	for _, param := range op.Params {
		if !flags.Changed(param.Flag) {
			continue
		}
		if useQueryTwin && param.In == "query" && param.Name != "limit" && param.Name != "cursor" {
			return req, false, Usagef("--%s filters the query string, which %s does not read; put it in the body query instead", param.Flag, op.QueryPath)
		}
		values, err := paramValues(flags, param)
		if err != nil {
			return req, false, err
		}
		switch param.In {
		case "header":
			for _, value := range values {
				req.Header.Add(param.Name, value)
			}
		default:
			for _, value := range values {
				req.Query.Add(param.Name, value)
			}
		}
	}
	if in.idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", in.idempotencyKey)
	}

	switch {
	case useQueryTwin:
		source := in.querySource
		if flags.Changed(FlagQueryFile) {
			source = "@" + in.queryFile
		}
		body, err := ReadJSONSource(source, opts.In)
		if err != nil {
			return req, false, &UsageError{Err: err}
		}
		if !json.Valid(body) {
			return req, false, Usagef("the body query is not valid JSON")
		}
		body, err = foldQueryStringIntoBody(body, req.Query)
		if err != nil {
			return req, false, err
		}
		req.Body = body
		req.Query = url.Values{}
		return req, true, nil

	case op.Body:
		body, err := BuildBody(in.jsonSource, in.sets, opts.In)
		if err != nil {
			return req, false, err
		}
		if len(body) == 0 && op.BodyRequired {
			return req, false, Usagef("this call needs a request body: pass --json '<inline>', --json @file, --json - or one or more --set key=value")
		}
		req.Body = body
		return req, false, nil
	}
	return req, false, nil
}

// foldQueryStringIntoBody moves limit and cursor into the body query, the
// only two members the query-string form and the body form share.
func foldQueryStringIntoBody(body []byte, query url.Values) ([]byte, error) {
	if len(query) == 0 {
		return body, nil
	}
	document := map[string]any{}
	if len(strings.TrimSpace(string(body))) > 0 {
		if err := json.Unmarshal(body, &document); err != nil {
			return nil, Usagef("the body query must be a JSON object")
		}
	}
	for _, name := range []string{"limit", "cursor"} {
		value := query.Get(name)
		if value == "" {
			continue
		}
		if _, taken := document[name]; taken {
			continue
		}
		if number, err := strconv.ParseFloat(value, 64); err == nil {
			document[name] = number
		} else {
			document[name] = value
		}
	}
	return json.Marshal(document)
}

// declareParams registers one flag per query and header parameter.
func declareParams(flags *pflag.FlagSet, op Operation) {
	for _, param := range op.Params {
		help := paramHelp(param)
		switch param.Type {
		case "array":
			flags.StringSlice(param.Flag, nil, help)
		case "boolean":
			flags.Bool(param.Flag, false, help)
		case "integer":
			flags.Int64(param.Flag, 0, help)
		case "number":
			flags.Float64(param.Flag, 0, help)
		default:
			flags.String(param.Flag, "", help)
		}
	}
}

// paramValues reads a parameter's flag back as the strings that go on the
// wire, honouring the array serialisation the document asked for.
func paramValues(flags *pflag.FlagSet, param Param) ([]string, error) {
	switch param.Type {
	case "array":
		items, err := flags.GetStringSlice(param.Flag)
		if err != nil {
			return nil, err
		}
		if err := checkEnum(param, items); err != nil {
			return nil, err
		}
		if param.Explode {
			return items, nil
		}
		if len(items) == 0 {
			return nil, nil
		}
		return []string{strings.Join(items, ",")}, nil
	case "boolean":
		value, err := flags.GetBool(param.Flag)
		if err != nil {
			return nil, err
		}
		return []string{strconv.FormatBool(value)}, nil
	case "integer":
		value, err := flags.GetInt64(param.Flag)
		if err != nil {
			return nil, err
		}
		return []string{strconv.FormatInt(value, 10)}, nil
	case "number":
		value, err := flags.GetFloat64(param.Flag)
		if err != nil {
			return nil, err
		}
		return []string{strconv.FormatFloat(value, 'f', -1, 64)}, nil
	default:
		value, err := flags.GetString(param.Flag)
		if err != nil {
			return nil, err
		}
		if err := checkEnum(param, []string{value}); err != nil {
			return nil, err
		}
		return []string{value}, nil
	}
}

// checkEnum refuses a value outside a closed set before it costs a round
// trip.
func checkEnum(param Param, values []string) error {
	if len(param.Enum) == 0 {
		return nil
	}
	for _, value := range values {
		if !contains(param.Enum, value) {
			return Usagef("--%s: %q is not one of %s", param.Flag, value, strings.Join(param.Enum, ", "))
		}
	}
	return nil
}

// fillPath substitutes the positional arguments into the path template.
func fillPath(template string, params []Arg, args []string) (string, error) {
	if len(args) != len(params) {
		return "", Usagef("this call takes %d argument(s): %s", len(params), argNames(params))
	}
	path := template
	for i, param := range params {
		value := strings.TrimSpace(args[i])
		if value == "" {
			return "", Usagef("<%s> must not be empty", kebab(param.Name))
		}
		path = strings.ReplaceAll(path, "{"+param.Name+"}", url.PathEscape(value))
	}
	if strings.ContainsAny(path, "{}") {
		return "", fmt.Errorf("the path template %q was not fully substituted", template)
	}
	return path, nil
}

func argNames(params []Arg) string {
	names := make([]string, 0, len(params))
	for _, param := range params {
		names = append(names, "<"+kebab(param.Name)+">")
	}
	return strings.Join(names, " ")
}

// longHelp is the command's long description: what it does, followed by
// what it maps to and what its arguments mean.
func longHelp(op Operation) string {
	var b strings.Builder
	if op.Description != "" {
		b.WriteString(op.Description)
	} else {
		b.WriteString(op.Summary)
	}
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "%s %s (%s)\n", op.Method, op.Path, op.ID)
	if len(op.Args) > 0 {
		b.WriteString("\nArguments:\n")
		for _, arg := range op.Args {
			fmt.Fprintf(&b, "  <%s>\t%s\n", kebab(arg.Name), arg.Help)
		}
	}
	if op.QueryPath != "" {
		fmt.Fprintf(&b, "\nFilters too long for a query string go in a body query: --query-file filter.json (POST %s).\n", op.QueryPath)
	}
	if op.Paged {
		b.WriteString("\nOne page is returned by default. --all walks every page.\n")
	}
	return b.String()
}

func paramHelp(param Param) string {
	help := strings.TrimSpace(param.Help)
	if help == "" {
		help = param.Name
	}
	if len(param.Enum) > 0 {
		help += " (" + listEnum(param.Enum) + ")"
	}
	if param.Type == "array" {
		help += " (repeatable)"
	}
	if param.Required {
		help += " (required)"
	}
	return help
}

// maxEnumInHelp caps how much of a closed value set a help line spells
// out. The whole set is still validated, and the reference documentation
// carries it in full.
const maxEnumInHelp = 8

func listEnum(values []string) string {
	if len(values) <= maxEnumInHelp {
		return strings.Join(values, " | ")
	}
	return strings.Join(values[:maxEnumInHelp], " | ") + fmt.Sprintf(" | ... %d more", len(values)-maxEnumInHelp)
}

func hasParam(op Operation, flag string) bool {
	for _, param := range op.Params {
		if param.Flag == flag {
			return true
		}
	}
	return false
}

func writes(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
