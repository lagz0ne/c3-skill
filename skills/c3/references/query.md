# Query Reference

Answer where/how/why about architecture. Full context = docs + code.

## Flow

`Discover Candidates → Read Detail → Lookup Files → Graph if Needed → Explore Code`

## Progress

- [ ] Candidates discovered (`c3 search "<question>"`, `c3 lookup <file>`, or `c3 read <id>`)
- [ ] Intent clarified (or skipped if specific)
- [ ] Entity matched from ranked candidates
- [ ] `c3 lookup` run on every file path surfaced
- [ ] Code explored
- [ ] Response delivered

---

## Step 0a: Discover Candidates

Choose the narrowest discovery tool that matches what the user gave you:

| Input shape | First command | Outcome |
|-------------|---------------|---------|
| Concept, capability, paraphrase, "where is X", "what handles Y" | `c3 search "<question>"` | Ranked candidate entities by semantic + keyword + graph signals |
| Known file or glob | `c3 lookup <file-or-glob>` | Owning component plus governing refs/rules |
| Known C3 ID | `c3 read <id>` | Entity detail |

For natural-language questions, reach for search before topology/title matching:

```bash
c3 search "how do users sign in and get permissions"
```

Use `match_sources` to understand why a candidate ranked: `semantic` means meaning matched even when wording differs; `content_fts` / `entity_fts` means keyword match; `graph:*` means relationship context helped.

Anti-goal: do not fall back to `c3 list` plus title matching when the query is conceptual. Search it. Use `c3 list` when you need topology-wide inventory, coverage, or relationship overview after candidates are known.

Never manually Glob/Read `.c3/`. Read only after identifying specific entities.

For code→entity mapping use `c3 lookup <file-or-glob>`. For doc body text, identify candidate entities from search/lookup, then read targeted bodies with `c3 read <id>` or `c3 read <id> --full`.

## Step 0a+: Check Recipes

Recipes are candidate entities too. If search returns a matching `recipe`, read it first and serve its sources as a narrative trace. If you already needed `c3 list` for inventory, filter `recipe` rows by title + description before deeper reads.

## Step 0a++: Cross-Cutting Flow Expansion

For questions like "how does X happen and inform users", "trace end-to-end",
"what breaks across sync/notifications", or any flow with side effects, do not
stop at the first owner. Build the path:

1. **Action / command owner**: read the screen/flow/component that accepts the
   user action or external command.
2. **State mutation owner**: read the component/service/ref that changes domain
   state and governs semantics.
3. **Sync mechanism**: search/read `ref-sync`, graph the mutating component, and
   name the concrete transport (`NATS WebSocket`, deltas/acks, `executionId`).
4. **Notification mechanism**: search/read notification candidates (`c3-211`,
   notification ADRs/refs/channels), then graph them to find dispatcher/channel
   touchpoints. Copy the exact queue/subject/channel names from the read docs:
   examples include `NATS JetStream`, `sync.user.<email>`, and `slackChannel`.
   Do not collapse these to generic "notification system" wording when the
   question is about who gets informed.
5. **Emergent property**: state what falls out of the combined mechanisms, such
   as async/non-blocking notification, targeted user subjects vs broadcast sync,
   step-advance-only notification, or flow entry preserving side effects.
6. **Failure boundary**: state how the combined path degrades — what happens if
   the sync/notification/transport leg fails or is bypassed, which side effects
   are preserved vs lost, and who observes the failure. If the docs don't say,
   report that as an explicit gap, never as a guess.

When saying no rules apply, say "no `rule-*` entities found" instead of naming
made-up negative ids like `rule-auth` or `rule-sync`.

## Step 0a+++ Diverse Property Expansion

When the question asks about a property that emerges across mechanisms, name the
property explicitly and trace the mechanism that creates it:

- **Partial failure / bulk operation / rollback / consistency:** trace action
  owner -> mutation owner -> transaction/audit/storage contract -> observation
  surface. Name the property as atomicity, consistency, partial-success
  boundary, or idempotency. For audit questions, read the audit ref/recipe and
  the execution/transaction owner; say whether consistency is per committed
  mutation or all-or-nothing for the whole batch.
- **Config / env / prefix / scope changes:** after search, read the config/scope
  ref and run `c3 graph <ref-or-component> --direction reverse` on both the
  config mechanism and the domain contract it feeds. Enumerate the concrete ids
  from the reverse graph; call this the blast radius or scope of impact.
- **Transport/auth feeding sync or another mechanism:** trace credential
  generation -> enforcement point -> dependent runtime path. Name the coupling,
  the shared token/permission/subject/config, and the way the dependent path
  fails even if the upstream feature still appears healthy.
- **Duplicate/import/file processing:** trace UI/command -> import/storage flow
  -> result taxonomy -> sync/audit/database boundary. Name idempotency,
  deduplication, transactionality, and partial-success behavior when present.

## Step 0b: Clarify Intent

Ask when (skip if ASSUMPTION_MODE):
- Vague ("how does X work?")
- Multiple interpretations ("authentication" — login? tokens?)
- Scope unclear

Skip when: C3 ID given, specific query, "show me everything about X".

## Step 1: Navigate Layers

Top-down: Context → Container → Component.

Start from the best search/lookup candidate. `c3 read <id>` when body content is needed beyond the candidate snippet. Move up to parent context or down to components when the answer needs ownership, boundaries, or implementation detail.

| Source | Use For |
|--------|---------|
| Component name | Class/module names |
| code-map (cache-backed) | Direct file paths, symbols |
| Technology | Framework patterns |

## Step 2: Extract + Lookup

For every file path:
1. **Run `c3 lookup <file>` before reading source** — returns component + governing refs/rules. Directory-level: `c3 lookup 'src/auth/**'`.
2. Check relationships via `c3 read <id>` or `c3 graph <id> --format mermaid`. Always include mermaid output as code block — graph matched entity (container or component, never c3-0).
3. Find `ref-*` and `rule-*` from topology. `c3 read <id>` for body.

Lookup-returned refs/rules = constraints governing that file.

## Step 3: Explore Code

```bash
# Glob patterns
src/auth/**/*.ts
# Grep class/function names
# Read specific files from code-map lookup
```

---

## Query Types

| Type | When | Response |
|------|------|---------|
| Docs | "where is X", "explain X" | Docs + suggest code |
| Code | "show me code for X" | Full flow through code |
| Deep | "explore X thoroughly" | Docs + Code + Related |
| Constraints | "what rules/refs apply to X" | Full constraint chain |

## Constraint Chain Query

1. Identify target (c3-NNN, c3-N, or c3-0)
2. Read upward: component → container → context
3. Extract: explicit constraints (MUST/MUST NOT), boundaries, layer rules
4. Collect cited refs/rules from the component's reference column (`c3 schema component` — today `Governance`), read key constraints

```
**Constraint Chain for c3-NNN (Name)**

**Context (c3-0):** [system-wide rule]
**Container (c3-N):** [container boundary]
**Refs:** ref-X: [key patterns]
**Rules:** rule-X: [key constraints]
**Layer Boundaries:** MAY: [...] MUST NOT: [...]
```

**Constraint Chain Graph:**
(mermaid from `c3 graph <target-component> --direction reverse --format mermaid`)

## ADR Handling

ADRs = ephemeral work orders, not architectural truth. Default listings exclude them.

**Use `--include-adr` ONLY when:**
- Working on specific ADR (implementing tracked change)
- Pre-staging feature (reviewing past decisions)
- Explicitly asked for historical decisions

**Never `--include-adr` for:**
- General architecture questions
- Code navigation
- Understanding current state

ADRs in results → historical context only. Verify against current entity docs before acting. Any ADR you cite in an answer gets an explicit label — current / superseded / historical — derived from its `status` and from newer ADRs or entity docs that cover the same ground.

## Edge Cases

| Situation | Action |
|-----------|--------|
| Topic not in C3 | Search code directly, suggest documenting |
| Spans containers | List all affected, explain relationships |
| Docs seem stale | Note, suggest audit |

## Answer Depth Contract

An answer is judged on depth, not term coverage. Naming the right entities while
skipping the causal chain is a failing answer. Hold every delivered answer to
these bars:

1. **Every material claim is bound to evidence you actually collected.** A claim
   about an entity's behavior, contract, or governance must come from a `read`/
   `graph`/`lookup`/`search` output you ran, and the answer says which (entity +
   section). Listing evidence commands up front grounds nothing by itself — if a
   claim has no read behind it, run the read or drop the claim. Cite entities by
   their exact c3 id as c3 output shows it — never a truncated id or a
   directory/path-derived name.
2. **Cross-cutting answers state the full causal chain** (Step 0a++ items 1–6)
   as a chain — action owner -> state mutation -> mechanism -> dependent/observer
   -> emergent property -> failure boundary — not as a flat entity list. Each
   arrow says WHY the next hop follows: which contract, ref, subject, or
   permission carries it.
3. **Direct vs indirect dependents are separated.** A reverse-graph neighbor is
   a candidate, not a conclusion. Assign behavior to a dependent only after
   reading it; label each one direct (cites/consumes the changed thing) or
   transitive (reached through another entity).
4. **Cited ADRs carry a status label** — current, superseded, or historical —
   checked against `status` and newer ADRs/entity docs. Never present an old
   decision as the live mechanism.
5. **Negatives and caveats are evidence-backed.** "No rules apply" requires the
   search/list output that showed it. A caveat ("may drift", "not enforced")
   requires the doc row or check output suggesting it; otherwise omit it.
6. **Change-usefulness means concrete checks, not advice.** When the question
   implies a change, end with verifiable checks: which owner files/entities to
   touch, which config/permission/runtime values to confirm, which sync or
   notification observable to assert, and how to probe the failure mode.
7. **UI/layout answers name pattern semantics, not ref lists**: which ref owns
   which concrete behavior (layout regions, sticky/empty/responsive states),
   what visibly breaks if the pattern changes, and how consistency is verified
   across the screens that share it.
8. **Side effects belong to a layer — verify which one before any bypass claim.**
   When a capability has multiple entry paths (UI, chat/bot action, API, direct
   service call), read where each side effect — sync emit, notification dispatch,
   audit write — is actually attached: flow layer, service layer, or storage
   trigger. Then state explicitly which side effects each entry path triggers
   and which it skips. Never assume a side effect survives entering below its
   attachment layer; the attachment point decides, and the doc that declares it
   is the evidence.

## Response Format

```
**Layer:** <c3-id> (<name>)

<Architecture from docs>

**Graph:**
(mermaid from c3 graph <matched-entity> --format mermaid)

**Code Map:** `path/file.ts` - <role>

**Key Insights:** <Observations>

**Related:** <navigation hints>
```
