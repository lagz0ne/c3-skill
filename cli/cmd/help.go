package cmd

import (
	"fmt"
	"io"
)

var globalHelp = `c3x - architecture-aware CLI for C3 projects

Usage: c3x <command> [args] [options]

Commands:
  init                       Scaffold .c3/ skeleton
  list                       Topology view with relationships
  check                      Validate docs, schema, code refs, consistency
  add <type> <slug>          Create entity (auto-numbering + wiring)
  set <id> <field> <value>   Update frontmatter field
  set <id> --section <name>  Update section content (text or JSON table)
  wire <src> cite <tgt>      Link component to ref (3-sided)
  unwire <src> cite <tgt>    Remove cite link (3-sided)
  schema <type>              Show known sections for entity type

Entity Types: context, container, component, ref, adr

Global Options:
  --json                     Machine-readable output
  --c3-dir <path>            Override .c3/ auto-detection
  -h, --help                 Show help
  -v, --version              Print version`

var commandHelp = map[string]string{
	"init": `Usage: c3x init

Scaffold .c3/ skeleton (config, README, refs/, adr/).`,

	"list": `Usage: c3x list [--flat] [--json]

Topology view of all entities with relationships.
  --flat    Simple file list
  --json    JSON output`,

	"check": `Usage: c3x check [--json]

Three-layer validation:
  Layer 1: Broken links, orphans, duplicates, missing parents
  Layer 2: Required sections empty/missing per schema
  Layer 3: Code refs exist on disk, entity IDs in graph, cite consistency`,

	"add": `Usage: c3x add <type> <slug> [options]

Types: container, component, ref, adr

Options:
  --container <id>       Parent container (component only)
  --feature              Feature numbering (10+) vs foundation (01-09)
  --goal <text>          Pre-fill goal in frontmatter + body
  --summary <text>       Pre-fill summary
  --boundary <text>      Pre-fill boundary (container only)

Examples:
  c3x add container payments --goal "Process payments" --boundary service
  c3x add component auth --container c3-1 --goal "JWT authentication"
  c3x add ref rate-limiting --goal "Consistent rate limiting"
  c3x add adr use-grpc --goal "Migrate to gRPC"`,

	"set": `Usage: c3x set <id> <field> <value>
       c3x set <id> --section <name> <value> [--append]

Frontmatter fields: goal, summary, status, boundary, category, title, date
Array via JSON:     scope, affects, refs

Section mode accepts text or JSON (array for replace, object for --append):
  c3x set c3-101 goal "Handle JWT auth"
  c3x set c3-101 --section "Code References" '[{"File":"src/auth.ts","Purpose":"Auth"}]'
  c3x set c3-101 --section "Dependencies" --append '{"Direction":"IN","What":"creds","From/To":"c3-102"}'`,

	"wire": `Usage: c3x wire <source> cite <target>

Creates cite relationship (3 sides updated atomically):
  1. source frontmatter refs[] += target
  2. source "Related Refs" table += row
  3. target "Cited By" table += row

Example: c3x wire c3-101 cite ref-jwt`,

	"unwire": `Usage: c3x unwire <source> cite <target>

Removes cite relationship from all 3 sides.
Idempotent — no error if not wired.

Example: c3x unwire c3-101 cite ref-jwt`,

	"schema": `Usage: c3x schema <type> [--json]

Show known sections for an entity type.
Types: context, container, component, ref, adr

JSON output includes column types (filepath, entity_id, enum, ref_id).

Example: c3x schema component --json`,
}

// ShowHelp prints command help or global help.
func ShowHelp(command string, w io.Writer) {
	if help, ok := commandHelp[command]; ok {
		fmt.Fprintln(w, help)
	} else {
		fmt.Fprintln(w, globalHelp)
	}
}
