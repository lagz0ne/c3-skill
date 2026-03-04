# C3 Knowledge Index

## Problem

Every C3 operation pays a cold-start tax: the LLM must discover which entities, files, refs, and constraints are relevant before it can reason about the developer's intent. A typical query costs 6-15 CLI calls. A multi-component change costs 40-70. This discovery+assembly work is repeated from scratch in every conversation.

Current storage is entity-oriented, but usage is task-oriented. The developer doesn't want "the c3-101 document" — they want "everything I need to know to safely change auth."

## Design

Two layers. Layer 1 eliminates discovery. Layer 2 eliminates assembly.

### Layer 1: Structural Index (CLI-computed, deterministic)

A readable text file projecting the full graph + code-map into a scannable format. Rebuilt as a side-effect of any command that loads the graph (`check`, `list`, `add`). Sub-second, no dedicated command needed.

Stored at `.c3/_index/structural.md`. Default format is human/LLM-readable text. JSON available via `--json` for programmatic access (jq, scripts, CI).

```markdown
# C3 Structural Index
<!-- hash: sha256:abc... -->

## c3-101 — auth (component)
container: c3-1 | context: c3-0
refs: ref-jwt
reverse deps: c3-102, c3-103
files: src/auth/middleware.ts, src/auth/jwt.ts
constraints from: c3-0, c3-1, ref-jwt
blocks: Goal ✓, Dependencies ✓, Related Refs ○

## c3-102 — session (component)
container: c3-1 | context: c3-0
refs: ref-jwt
reverse deps: c3-104
files: src/session/*.ts
constraints from: c3-0, c3-1, ref-jwt
blocks: Goal ✓, Dependencies ○

## ref-jwt — JWT Authentication (ref)
citers: c3-101, c3-102, c3-205, c3-312
scope: c3-1
blocks: Goal ✓, Choice ✓, Why ✓

## File Map
src/auth/middleware.ts → c3-101 | refs: ref-jwt
src/auth/jwt.ts → c3-101 | refs: ref-jwt
src/session/*.ts → c3-102 | refs: ref-jwt
```

This replaces the discovery phase: instead of list → get → lookup → get refs (6-15 calls), the LLM reads one file and has the full structural picture.

The index is a projection of data already in memory during `c3x list`. No new data sources needed — just a different output shape. JSON format (`--json`) available when the content needs to be accessed programmatically.

### Layer 2: Topic Notes (LLM-generated, narrative + metadata)

Cross-cutting narrative notes stored in `.c3/_index/notes/`. The LLM identifies topics (not one-per-entity, but themes like "authentication flow" or "data persistence strategy") and writes short notes with provenance.

Each note has YAML frontmatter for machine checks and a freeform markdown body:

```markdown
---
topic: authentication-flow
sources:
  - c3-101#Goal
  - c3-101#Dependencies
  - c3-102#Dependencies
  - c3-103#Dependencies
  - c3-1#Responsibilities
  - c3-0#Abstract Constraints
  - ref-jwt#Choice
  - ref-jwt#Why
source_hash: sha256:def...
generated_at: 2026-03-04
status: current
---

Auth validates JWT tokens at the API gateway (c3-101). The stateless
pattern comes from ref-jwt — all 4 citing components must validate on
every request, never cache auth state.

Changing auth: ref-jwt's Choice section defines the contract. c3-102
(session) and c3-103 (permissions) consume claims from auth — check
their Dependencies tables for coupling points. Container constraint
from c3-1: all auth paths under 100ms.

Risk: ref-jwt is cited across 2 containers. A change here has blast
radius beyond the API container.
```

Notes are 200-400 words. The `sources` field uses anchored citations (`entity#section`) for verifiable provenance. The `source_hash` is computed from the concatenation of all cited sections — the CLI can compare this against current content to flag stale notes.

## Usage Flow

```
Intent arrives ("add rate limiting to auth")
    │
    ├── 1. Read structural index
    │      → instant: c3-101 is auth, in c3-1, cites ref-jwt,
    │        files are src/auth/*.ts, reverse deps: c3-102, c3-103
    │
    ├── 2. Check topic notes (any note whose sources overlap targets?)
    │      → found "authentication-flow" — pre-assembled context
    │        with constraint chain, coupling points, blast radius
    │
    ├── 3. Read source docs for targets
    │      → verify note accuracy against current state
    │
    └── 4. Act (answer question, plan change, assess impact)
```

Step 2 is the shortcut. If no relevant note exists, fall back to step 3 with more entities — same as today, but with step 1 already done.

When index is absent: full fallback to existing CLI call chain + warning. No degradation in correctness, only in speed.

## Lifecycle

### Creation

**Structural index**: Rebuilt automatically by any graph-loading command. Deterministic, sub-second.

**Topic notes**: LLM generates after onboard completes or after significant changes (new component, new ref, ADR accepted). The skill instructs: "read the structural index, identify cross-cutting themes, write a note for each."

### Staleness

**Structural index**: No staleness concern. Rebuilt from source every time. Hash in file allows consumers to detect if it's current.

**Topic notes**: Dual mechanism:
1. **Mechanical**: `c3x index check` computes current hash of each note's `sources` sections, compares against `source_hash`. Mismatches set `status: stale`.
2. **LLM reasoning**: When consulting a note, the LLM also reads the cited source docs. If content contradicts the note, the LLM knows to re-derive.

### Sunset

Notes whose `sources` reference entities that no longer exist are flagged by `c3x index check` as orphaned. The LLM or user can delete them.

## CLI Additions

No new commands. Everything piggybacks on existing commands:

| Existing command | Added behavior |
|-----------------|---------------|
| `c3x check` | Rebuild structural index + flag stale/orphaned topic notes |
| `c3x list` | Rebuild structural index |
| `c3x add` | Rebuild structural index |

The structural index rebuild is a projection of in-memory data — no extra I/O beyond writing the JSON file. It piggybacks on the graph load these commands already pay for.

Note health checks (stale `source_hash`, orphaned sources) fold into `c3x check` as an additional validation layer alongside the existing schema and code-map checks.

Topic notes are written by the LLM through the skill's normal file write operations. The CLI does not generate them.

### `c3x query` — eliminated

The structural index replaces `query`'s discovery/routing function (catalog, chain traversal, file resolution). The LLM reads source `.c3/` markdown files directly for content — structured block extraction adds no value over what LLMs already do with markdown natively.

| Former query mode | Replaced by |
|-------------------|------------|
| Catalog (entity list + fill status) | `structural.json` → `entities` with `block_fill` |
| Snapshot (all blocks for entity) | LLM reads the `.c3/` markdown file directly |
| Chain (hierarchy + refs) | `structural.json` → `container`, `context`, `refs`, `constraints_from` |
| File resolution (file → entity) | `structural.json` → `files` map |

## What This Replaces

| Scenario | Today | With index |
|----------|-------|-----------|
| Any question about a component | 6-15 CLI calls to discover and assemble context | 2-3 reads (index + note + source verification) |
| Cross-cutting question | 5-8 doc reads, mental stitching | 1 topic note read + selective source verification |
| Pre-change impact check | Walk graph manually via serial CLI calls | Read structural index for instant blast radius |
| "What refs apply to this file?" | lookup → get entity → get refs → read each ref | Read structural index files section |

## Design Decisions

1. **CLI computes structure, LLM reasons about meaning.** The structural index is a deterministic projection. Topic notes require understanding.

2. **Topic-oriented, not entity-oriented notes.** Cross-cutting themes ("authentication flow") are more useful than per-entity summaries because developer questions are task-oriented.

3. **Narrative body + structured metadata.** Notes are freeform markdown for expressiveness. YAML frontmatter (`sources`, `source_hash`, `status`) enables machine checks without constraining the narrative.

4. **Anchored citations for provenance.** `c3-101#Goal` is verifiable. `c3-101` alone is too coarse to detect whether the specific content the note drew from has changed.

5. **Graceful degradation.** When index is absent, all existing CLI functionality works unchanged. The index is pure acceleration, not a dependency.

## Open Questions

- **Topic granularity**: How many notes is right for a 20-entity system? 50-entity? Need real usage data.
- **Note governance**: When notes overlap or contradict, how should the LLM resolve? Start with skill-prompt discipline, add tooling if needed.
