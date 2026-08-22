package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newDocsCommand() *cobra.Command {
	var dir string
	cmd := &cobra.Command{
		Use:    "docs",
		Short:  "Write the command reference as markdown",
		Long:   "Write the command reference as markdown: an index plus one page per command group.",
		Args:   cobra.NoArgs,
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			pages, err := renderDocs(root)
			if err != nil {
				return err
			}
			for name, content := range pages {
				if err := os.WriteFile(filepath.Join(dir, name), content, 0o644); err != nil {
					return err
				}
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "wrote %d pages to %s\n", len(pages), dir)
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "docs", "directory the reference is written to")
	return cmd
}

// renderDocs builds the reference: README.md plus one page per group.
func renderDocs(root *cobra.Command) (map[string][]byte, error) {
	pages := map[string][]byte{}

	children := visible(root)
	var index bytes.Buffer
	index.WriteString("# `ht-cli` command reference\n\n")
	index.WriteString("Generated with `ht-cli docs`. Do not edit by hand.\n\n")
	index.WriteString(root.Long + "\n\n")
	index.WriteString("## Global flags\n\n")
	index.WriteString("| Flag | Type | Description |\n|---|---|---|\n")
	index.WriteString(flagTable(root.PersistentFlags()))
	index.WriteString("\n## Commands\n\n")
	index.WriteString("| Group | What it covers | Commands |\n|---|---|---|\n")

	for _, child := range children {
		page := child.Name() + ".md"
		count := len(visible(child))
		if count == 0 {
			count = 1
		}
		fmt.Fprintf(&index, "| [`ht-cli %s`](%s) | %s | %d |\n", child.Name(), page, child.Short, count)

		var body bytes.Buffer
		fmt.Fprintf(&body, "# `ht-cli %s`\n\n%s\n\n", child.Name(), child.Short)
		if child.Long != "" && child.Long != child.Short {
			fmt.Fprintf(&body, "%s\n\n", child.Long)
		}
		if grandchildren := visible(child); len(grandchildren) > 0 {
			for _, grandchild := range grandchildren {
				renderCommand(&body, grandchild)
			}
		} else {
			renderCommand(&body, child)
		}
		body.WriteString("---\n\n[Back to the index](README.md)\n")
		pages[page] = body.Bytes()
	}

	index.WriteString("\n## Exit codes\n\n")
	index.WriteString(exitCodeTable)
	pages["README.md"] = index.Bytes()
	return pages, nil
}

const exitCodeTable = `| Code | Meaning |
|---|---|
| 0 | The command did what it was asked. |
| 1 | A failure with no more specific code. |
| 2 | The command line was wrong: an unknown flag, a missing argument, a bad value. |
| 3 | The credential is missing, rejected or under-scoped. |
| 4 | The address names nothing. |
| 5 | The API refused the request: validation, a conflict, a precondition. |
| 6 | Throttled, or the quota is exhausted. |
| 7 | The API could not be reached, or faulted. |
`

// renderCommand writes one command's section.
func renderCommand(b *bytes.Buffer, cmd *cobra.Command) {
	fmt.Fprintf(b, "## `%s`\n\n", cmd.CommandPath())
	if cmd.Short != "" {
		fmt.Fprintf(b, "%s\n\n", cmd.Short)
	}
	fmt.Fprintf(b, "```\n%s\n```\n\n", cmd.UseLine())
	if cmd.Long != "" && cmd.Long != cmd.Short {
		fmt.Fprintf(b, "%s\n\n", strings.TrimSpace(cmd.Long))
	}
	if table := flagTable(cmd.NonInheritedFlags()); table != "" {
		b.WriteString("| Flag | Type | Description |\n|---|---|---|\n")
		b.WriteString(table)
		b.WriteString("\n")
	}
	for _, child := range visible(cmd) {
		renderCommand(b, child)
	}
}

// flagTable renders a flag set as markdown table rows.
func flagTable(flags *pflag.FlagSet) string {
	var rows []string
	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}
		name := "`--" + flag.Name + "`"
		if flag.Shorthand != "" {
			name = "`-" + flag.Shorthand + ", --" + flag.Name + "`"
		}
		rows = append(rows, fmt.Sprintf("| %s | %s | %s |\n", name, flag.Value.Type(), escapePipes(flag.Usage)))
	})
	sort.Strings(rows)
	return strings.Join(rows, "")
}

func escapePipes(text string) string {
	return strings.ReplaceAll(strings.Join(strings.Fields(text), " "), "|", `\|`)
}

// visible lists a command's own subcommands, skipping the ones cobra adds
// for itself.
func visible(cmd *cobra.Command) []*cobra.Command {
	var out []*cobra.Command
	for _, child := range cmd.Commands() {
		if child.Hidden || child.Name() == "help" || child.Name() == "completion" {
			continue
		}
		out = append(out, child)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}
