# Audit

**Question:** is the sealed truth intact and consistent?

The facts froze at Act 1 and change only through Act-2 change-units (see SKILL.md). Audit checks that what is frozen still holds together — it never repairs by hand. The one fix loop is to author a change-unit and `c3 change apply` (change.md). Audit reads; the change-unit writes.

Work three layers, outermost first. Stop and report at the first layer that fails — a broken seal makes every deeper finding unreliable.

## Layer 1 — Seal

`.c3/` markdown is the canonical truth; `.c3/c3.db` is a rebuildable cache sealed to match it. A branch switch, selective merge, or conflict resolution can desync the two.

```
c3 check
```

If `check` reports seal drift or cache divergence:

```
c3 repair
```

`repair` rebuilds the cache from canonical markdown and re-exports so seals match. It realigns the seal only — it invents no content fixes. If `check` still fails after `repair`, the canonical files themselves are wrong: that is a Layer-2 finding, fixed through a change-unit, not `repair`.

## Layer 2 — Structural

Run `c3 check` and read its output. Do not hand-walk membership tables against directories — the tool already validates:

- broken links, orphans, duplicate ids, missing parents
- required sections empty or missing, per each entity's canvas (the canvas definition is the contract — a project that edited a definition changed what is enforced; canvas.md)
- code refs resolve on disk, cited entity ids exist in the graph, cite consistency holds
- coverage signal `mapped / (total − excluded)` — `_exclude` patterns don't penalize the score; low coverage → WARN; suggest `_exclude` for test/config files and map the rest

Two structural facts the tool guarantees, so audit must never flag them as gaps:

- **Membership is synthesized**, not authored — every parent's membership rows are derived from children's `parent:` edges on `add` and `check --fix`. Never report "missing membership row"; a real disconnect is a missing-parent error, which `check` already raises.
- **The retire gate holds the graph closed** — a retire that would orphan a live child or dangle a live citer is refused unless the same change-unit heals it (change.md). So a clean `check` means no removal left a dangling reference behind.

## Layer 3 — Semantic

What `check` cannot judge — read a sample and assess:

- **Orphan refs/rules:** a ref or rule cited by zero components is dead weight → WARN. Confirm via `c3 graph <id> --direction reverse`.
- **Actionable rationale:** spot-check a ref's `## How` / a rule's `## Golden Example` — can you derive a YES/NO compliance question from it, and does a cited component's code hold to it? If the guidance is too vague to check, that's the finding (the standard needs rework), not the code. Compliance specifics live in ref.md and rule.md.

## ADR lifecycle (`--include-adr` only)

ADRs are hidden from default `c3` ops; audit them only on request or `c3 check --include-adr`. Canonical status set: `[open, accepted, done, superseded]`. Terminal docs (`done`, `superseded`) are content-frozen and check-exempt by design — leave them. The one signal worth surfacing: a unit stuck at `accepted`, long unapplied — its After-cites never resolved through `apply`. Surface it; do not hand-close it. Closing it is its own change-unit.

## Output

End in a verdict.

```
**C3 Audit Results**

| Layer       | Status         | Findings |
|-------------|----------------|----------|
| Seal        | PASS/WARN/FAIL | …        |
| Structural  | PASS/WARN/FAIL | …        |
| Semantic    | PASS/WARN/FAIL | …        |

**Summary:** N passes, M warnings, K failures
**Fixes:** each requires a change-unit (c3 change apply) — see change.md
```
