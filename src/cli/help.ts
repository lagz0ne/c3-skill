// src/cli/help.ts

const GLOBAL_HELP = `
c3x — Architecture-aware CLI for C3 projects

Commands:
  list                   Topology view with relationships
  check                  Doc integrity (broken links, orphans, duplicates)
  init                   Scaffold .c3/ skeleton
  add <type> <slug>      Create entity with auto-numbering + wiring
  sync                   Generate guard skills from component docs

Options:
  -h, --help             Show help (with command for details)
  -v, --version          Print version
  --json                 Machine-readable output
  --c3-dir <path>        Override .c3/ detection

Run 'npx c3-kit <command> --help' for details and examples.
`.trim();

const COMMAND_HELP: Record<string, string> = {
  list: `
Usage: c3x list [options]

Topology view of all C3 architecture docs with relationships

Options:
  --flat                 Simple file list (no topology)
  --json                 Output as JSON

Examples:
  c3x list
  c3x list --flat
  c3x list --json
`.trim(),

  check: `
Usage: c3x check [options]

Doc integrity check: broken links, orphans, duplicates, missing parents

Options:
  --json                 Output as JSON

Examples:
  c3x check
  c3x check --json
`.trim(),

  init: `
Usage: c3x init

Scaffold a new .c3/ directory skeleton

Creates:
  .c3/config.yaml
  .c3/README.md              (context template)
  .c3/refs/                  (empty)
  .c3/adr/adr-00000000-c3-adoption.md

Examples:
  c3x init
`.trim(),

  add: `
Usage: c3x add <type> <slug> [options]

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
  c3x add adr oauth-support
`.trim(),

  sync: `
Usage: c3x sync

Generate guard skills from .c3/ component docs into .claude/skills/

Reads Code References from each component doc and generates
per-component Claude Code skills that trigger on matching file paths.

Examples:
  c3x sync
`.trim(),

};

export function showHelp(command?: string): void {
  if (command && COMMAND_HELP[command]) {
    console.log(COMMAND_HELP[command]);
  } else {
    console.log(GLOBAL_HELP);
  }
}
