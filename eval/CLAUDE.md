# Eval System Development Guide

This eval system tests C3 skills using the Claude Agent SDK.

## Evaluation Approaches

### 1. Expectations-Based (Original)
Checks if specific expectations are met (structural).
```bash
bun eval/run.ts eval/cases/onboard-simple.yaml
```

### 2. Question-Based (Meaningful)
Tests if docs actually help understand the system:
```bash
bun eval/test-question-eval.ts <codebase-path> <docs-path>
```

**How it works:**
1. Analyzes codebase → extracts components, patterns, interactions
2. Generates questions a developer would ask
3. Answers questions using ONLY the docs (fresh LLM context)
4. Verifies answers against the actual codebase
5. Scores based on % of questions correctly answered

**Why it's better:**
- Grounded in THIS codebase, not a generic golden
- Tests actual utility, not structural similarity
- Clear debugging: know exactly which question failed

## Quick Commands

```bash
# Run a single test case (expectations-based)
bun eval/run.ts eval/cases/onboard-simple.yaml

# Run all test cases
bun eval/run.ts eval/cases/

# Run with verbose output (see tool calls, agent responses)
bun eval/run.ts eval/cases/onboard-simple.yaml --verbose

# Keep temp directory for debugging
bun eval/run.ts eval/cases/onboard-simple.yaml --keep

# Combine flags
bun eval/run.ts eval/cases/ --verbose --keep
```

## API Discovery

### Bun APIs

Bun types and documentation are in `node_modules/bun-types/`:

```bash
# Main type definitions
node_modules/bun-types/bun.d.ts        # Core Bun APIs (222k lines)
node_modules/bun-types/shell.d.ts      # Bun.$ shell API
node_modules/bun-types/test.d.ts       # bun:test APIs

# Documentation
node_modules/bun-types/docs/           # MDX docs for all features
node_modules/bun-types/CLAUDE.md       # Quick reference
```

Key Bun APIs used in this eval:
- `Bun.file(path)` - Read files
- `Bun.write(path, content)` - Write files
- `Bun.$\`command\`` - Shell execution (template literal)
- `Bun.spawn()` - Process spawning
- `Bun.sleep(ms)` - Async sleep

### Claude Agent SDK APIs

SDK types are in `node_modules/@anthropic-ai/claude-agent-sdk/`:

```bash
# Main SDK types (read this to understand the full API)
node_modules/@anthropic-ai/claude-agent-sdk/sdk.d.ts

# Key exports to look for:
# - query() function and Options type
# - HookInput, HookJSONOutput types
# - PreToolUseHookInput, PostToolUseHookInput
# - SDKMessage types (SDKAssistantMessage, SDKResultMessage, etc.)
# - SdkPluginConfig for loading plugins
```

Key SDK features used:
```typescript
import { query, type HookInput, type HookJSONOutput } from "@anthropic-ai/claude-agent-sdk";

query({
  prompt: "...",
  options: {
    cwd: "/path/to/workdir",
    plugins: [{ type: "local", path: "/path/to/plugin" }],
    permissionMode: "bypassPermissions",
    allowDangerouslySkipPermissions: true,
    hooks: {
      PreToolUse: [{ hooks: [async (input) => ({ ... })] }],
      PostToolUse: [{ hooks: [async (input) => ({ ... })] }],
    },
  },
});
```

## Architecture

```
eval/
├── run.ts           # Main runner - SDK query with inline hooks
├── lib/
│   ├── judge.ts     # LLM judge - evaluates output vs expectations
│   └── types.ts     # Shared TypeScript types
├── cases/           # Test case YAML files
│   ├── onboard-simple.yaml
│   └── onboard-greenfield.yaml
├── fixtures/        # Test project fixtures
│   ├── simple-express-app/
│   └── greenfield-saas/
└── results/         # Timestamped test results (JSON)
```

## Test Cases

| Case | Skill Tested | Description |
|------|--------------|-------------|
| `onboard-simple.yaml` | c3:onboard | Brownfield Express API |
| `onboard-greenfield.yaml` | c3:onboard | Greenfield SaaS design |
| `onboard-monorepo.yaml` | c3:onboard | Turborepo with apps/packages |
| `query-architecture.yaml` | c3:query | Query existing C3 docs |
| `audit-stale-refs.yaml` | c3:audit | Detect stale file references |
| `alter-add-feature.yaml` | c3:alter | Add feature via ADR workflow |

## Fixtures

| Fixture | Type | Contents |
|---------|------|----------|
| `simple-express-app/` | Brownfield | Express API, no C3 docs |
| `greenfield-saas/` | Greenfield | Just package.json + README |
| `documented-api/` | Documented | Hono API with full C3 docs |
| `react-monorepo/` | Monorepo | Turborepo (web, api, packages) |
| `stale-project/` | Stale | C3 docs with missing file refs |

## Test Case Format

```yaml
name: "Test name"
fixtures: "fixtures/directory-name"
command: |
  The prompt to send to the agent...

goal: |
  Description of what should be achieved...

constraints:
  - Constraint 1
  - Constraint 2

expectations:
  - "Expected outcome 1"
  - "Expected outcome 2"
```

## Development Loop

### 1. Type Check (fast)
```bash
bunx @typescript/native-preview --noEmit eval/run.ts eval/lib/*.ts
```
Note: Bun-specific errors are expected (tsc doesn't know Bun types).

### 2. Syntax Check (instant)
```bash
bun eval/run.ts --help 2>&1 | head -1
# Should output usage, not a syntax error
```

### 3. Run Single Test (medium)
```bash
bun eval/run.ts eval/cases/onboard-simple.yaml --verbose
```

### 4. Run All Tests (slow)
```bash
bun eval/run.ts eval/cases/
```

### 5. Debug Failed Test
```bash
# Keep the temp directory
bun eval/run.ts eval/cases/failing-test.yaml --keep --verbose

# Inspect the output
ls /tmp/tmp.XXXXXX/.c3/
cat /tmp/tmp.XXXXXX/.c3/README.md
```

## Adding New Test Cases

1. Create fixture directory in `eval/fixtures/`
2. Create YAML test case in `eval/cases/`
3. Run with `--verbose --keep` to debug
4. Iterate until passing

## Hooks Architecture

The eval uses SDK hooks for:

1. **Tracing** - PreToolUse/PostToolUse hooks record all tool calls
2. **Sandboxing** - PreToolUse checks file paths are within temp dir
3. **Human Simulation** - PreToolUse intercepts AskUserQuestion and uses a separate Claude call to generate answers based on the test goal

Hook output format:
```typescript
// Allow tool call
{ hookSpecificOutput: { hookEventName: "PreToolUse", permissionDecision: "allow" } }

// Deny tool call
{ hookSpecificOutput: { hookEventName: "PreToolUse", permissionDecision: "deny", permissionDecisionReason: "..." } }

// Allow with modified input (for AskUserQuestion)
{ hookSpecificOutput: { hookEventName: "PreToolUse", permissionDecision: "allow", updatedInput: { ...modified } } }
```

## Troubleshooting

### "Cannot find module" errors
```bash
bun install
```

### SDK type errors
Read `node_modules/@anthropic-ai/claude-agent-sdk/sdk.d.ts` to understand the actual types.

### Test hangs
The agent might be waiting for input. Run with `--verbose` to see what's happening.

### All tests fail with same error
Check if the c3 plugin is being loaded correctly:
```typescript
plugins: [{ type: "local", path: PLUGIN_ROOT }]
```
