package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

// maxColumns caps how wide a generated table gets. A row carrying more
// members than this is still complete in --output json.
const maxColumns = 8

// preferred orders the columns a reader looks for first. Members outside
// it keep the order the answer declared them in.
var preferred = []string{
	"id", "name", "type", "kind", "state", "status", "url", "address",
	"email", "enabled", "active", "confirmed", "interval", "code",
	"title", "value", "count", "total",
}

// instants are the members carrying Unix seconds. The wire spells every
// instant that way, so a table renders them as UTC to stay readable.
var instants = map[string]bool{
	"created": true, "updated": true, "since": true, "from": true, "to": true,
}

// Table renders a decoded answer as an aligned table: a collection as one
// row per item, an object as a name/value list, a scalar as itself.
func Table(w io.Writer, value any) error {
	switch v := value.(type) {
	case nil:
		return nil
	case map[string]any:
		if rows, ok := v["data"].([]any); ok {
			if err := rowTable(w, rows); err != nil {
				return err
			}
			return envelopeFooter(w, v, len(rows))
		}
		return objectTable(w, v)
	case []any:
		return rowTable(w, v)
	default:
		_, err := fmt.Fprintln(w, scalar(value))
		return err
	}
}

// rowTable prints one row per item, with the columns the items share.
func rowTable(w io.Writer, rows []any) error {
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "(no rows)")
		return err
	}
	objects := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		obj, ok := row.(map[string]any)
		if !ok {
			// A collection of scalars prints as a single column.
			for _, item := range rows {
				if _, err := fmt.Fprintln(w, scalar(item)); err != nil {
					return err
				}
			}
			return nil
		}
		objects = append(objects, obj)
	}

	columns := columnsOf(objects)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, strings.Join(headers(columns), "\t"))
	for _, obj := range objects {
		cells := make([]string, 0, len(columns))
		for _, column := range columns {
			cells = append(cells, cell(column, obj[column]))
		}
		fmt.Fprintln(tw, strings.Join(cells, "\t"))
	}
	return tw.Flush()
}

// objectTable prints a single object as a name/value list.
func objectTable(w io.Writer, obj map[string]any) error {
	keys := orderKeys(keysOf([]map[string]any{obj}))
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, key := range keys {
		fmt.Fprintf(tw, "%s\t%s\n", header(key), cell(key, obj[key]))
	}
	return tw.Flush()
}

// envelopeFooter reports what the collection envelope says beyond its
// rows: how many were matched, and whether a page was left behind.
func envelopeFooter(w io.Writer, envelope map[string]any, shown int) error {
	var notes []string
	if count, ok := envelope["count"].(map[string]any); ok {
		if matched, ok := count["matched"]; ok {
			notes = append(notes, fmt.Sprintf("%s matched", scalar(matched)))
		} else if total, ok := count["total"]; ok {
			notes = append(notes, fmt.Sprintf("%s total", scalar(total)))
		}
	}
	if more, _ := envelope["hasMore"].(bool); more {
		notes = append(notes, "more rows available (pass --all, or --cursor with nextCursor)")
	}
	if len(notes) == 0 {
		return nil
	}
	_, err := fmt.Fprintf(w, "\n%d shown, %s\n", shown, strings.Join(notes, "; "))
	return err
}

// columnsOf picks the columns of a row table: the scalar members the
// items declare, in preferred order, capped at maxColumns.
func columnsOf(objects []map[string]any) []string {
	scalars := make([]string, 0, 8)
	for _, key := range keysOf(objects) {
		for _, obj := range objects {
			value, ok := obj[key]
			if !ok {
				continue
			}
			if isScalar(value) {
				scalars = append(scalars, key)
			}
			break
		}
	}
	ordered := orderKeys(scalars)
	if len(ordered) > maxColumns {
		ordered = ordered[:maxColumns]
	}
	return ordered
}

// keysOf collects every member name in first-seen order.
func keysOf(objects []map[string]any) []string {
	seen := map[string]bool{}
	var out []string
	for _, obj := range objects {
		names := make([]string, 0, len(obj))
		for name := range obj {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			if !seen[name] {
				seen[name] = true
				out = append(out, name)
			}
		}
	}
	return out
}

// orderKeys sorts by the preferred list, leaving the rest as they came.
func orderKeys(keys []string) []string {
	rank := map[string]int{}
	for i, key := range preferred {
		rank[key] = i
	}
	out := append([]string(nil), keys...)
	sort.SliceStable(out, func(i, j int) bool {
		ri, oki := rank[out[i]]
		rj, okj := rank[out[j]]
		switch {
		case oki && okj:
			return ri < rj
		case oki:
			return true
		case okj:
			return false
		default:
			return false
		}
	})
	return out
}

func headers(columns []string) []string {
	out := make([]string, 0, len(columns))
	for _, column := range columns {
		out = append(out, header(column))
	}
	return out
}

// header spells a member name as a column title: UPDATED AT, not updatedAt.
func header(name string) string {
	var b strings.Builder
	for i, r := range name {
		if r >= 'A' && r <= 'Z' && i > 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	return strings.ToUpper(b.String())
}

func isScalar(value any) bool {
	switch value.(type) {
	case map[string]any, []any:
		return false
	default:
		return true
	}
}

// cell renders one value, folding a nested member into a placeholder so
// the table keeps its shape.
func cell(name string, value any) string {
	switch v := value.(type) {
	case nil:
		return "-"
	case map[string]any:
		return fmt.Sprintf("{%d}", len(v))
	case []any:
		return fmt.Sprintf("[%d]", len(v))
	}
	if isInstant(name) {
		if s, ok := asInstant(value); ok {
			return s
		}
	}
	return scalar(value)
}

// isInstant reports whether a member name carries Unix seconds: the
// spelled-out ones, plus the `...At` convention.
func isInstant(name string) bool {
	return instants[name] || strings.HasSuffix(name, "At")
}

// asInstant renders Unix seconds as UTC, but only for a value inside the
// range an instant plausibly occupies (2001 to 2096), so a plain counter
// under an `...At` name is never mangled.
func asInstant(value any) (string, bool) {
	var seconds int64
	switch v := value.(type) {
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			return "", false
		}
		seconds = n
	case float64:
		seconds = int64(v)
	case int64:
		seconds = v
	default:
		return "", false
	}
	if seconds < 1_000_000_000 || seconds > 4_000_000_000 {
		return "", false
	}
	return time.Unix(seconds, 0).UTC().Format(time.RFC3339), true
}

func scalar(value any) string {
	switch v := value.(type) {
	case nil:
		return "-"
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case json.Number:
		return v.String()
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(raw)
	}
}
