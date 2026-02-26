package cmd

import (
	"fmt"
	"io"
)

var globalHelp = `c3x — Architecture-aware CLI for C3 projects

Commands:
  list                   Topology view with relationships
  check                  Doc integrity (broken links, orphans, duplicates)
  init                   Scaffold .c3/ skeleton
  add <type> <slug>      Create entity with auto-numbering + wiring

Options:
  -h, --help             Show help (with command for details)
  -v, --version          Print version
  --json                 Machine-readable output
  --c3-dir <path>        Override .c3/ detection

Run 'c3x <command> --help' for details and examples.`

var commandHelp = map[string]string{
	"list": `Usage: c3x list [options]

Topology view of all C3 architecture docs with relationships

Options:
  --flat                 Simple file list (no topology)
  --json                 Output as JSON

Examples:
  c3x list
  c3x list --flat
  c3x list --json`,

	"check": `Usage: c3x check [options]

Doc integrity check: broken links, orphans, duplicates, missing parents

Options:
  --json                 Output as JSON

Examples:
  c3x check
  c3x check --json`,

	"init": `Usage: c3x init

Scaffold a new .c3/ directory skeleton

Creates:
  .c3/config.yaml
  .c3/README.md              (context template)
  .c3/refs/                  (empty)
  .c3/adr/adr-00000000-c3-adoption.md

Examples:
  c3x init`,

	"add": `Usage: c3x add <type> <slug> [options]

Create a new C3 entity with auto-numbering and relationship wiring

Types: container, component, ref, adr

Options:
  --container <id>       Parent container (required for component)
  --feature              Feature component (10+) instead of foundation (01-09)

Examples:
  c3x add container payments
  c3x add component auth-provider --container c3-3
  c3x add component checkout --container c3-3 --feature
  c3x add ref rate-limiting
  c3x add adr oauth-support`,
}

// ShowHelp prints command help or global help.
func ShowHelp(command string, w io.Writer) {
	if help, ok := commandHelp[command]; ok {
		fmt.Fprintln(w, help)
	} else {
		fmt.Fprintln(w, globalHelp)
	}
}
