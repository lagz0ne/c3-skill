# Ref

A **ref** is one of the eight frozen-fact types (SKILL.md §The shared contract). Its identity is **rationale**: a ref exists to record *why this choice over the realistic alternatives*, so `## Why` is the load-bearing section. Strip it and the doc has no reason to be a fact.

**Identity test — can't name a concrete file → ref, not component.** A ref captures a pattern that recurs across components; if it lives in one file, it is that component's concern, not a shared fact.

## Ref vs Rule — the Separation Test

A ref carries *rationale*; a rule carries an *enforceable standard*. When a doc could be either, ask:

> **Remove the `## Why` section. Is the doc now useless?**

| Answer | Type | Where |
|--------|------|-------|
| Yes — useless without the rationale | **Ref** | here |
| No — it still tells you what to do | **Rule** | `references/rule.md` |
| Both — rationale *and* an enforced standard | **Split** | a ref for the *why* + a rule for the *enforcement*; `references/rule.md` §Migrate |

A doc that is primarily golden code, anti-patterns, or a coding standard is a **rule**, not a ref — `C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" schema ref` rejects it on exactly this (`Pattern is primarily about enforcement … that's a rule`).

## Create a ref

A ref is created whole — author the full body into a file, then add it. Create is unguarded (the freeze applies only once a fact has a body).

```bash
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" schema ref          # leads with REJECT IF — draft to the contract, don't freehand
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" add ref <slug> --file ref.md
```

`C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" schema ref` is the authoring spec: it names every section, its `fill:` guidance, and the `reject_if` bullets that gate the doc (chiefly: `## Why` must give rationale, not restate `## Choice`; `## How`, if present, must cite a real file). There is **no scaffold-then-fill** — the first body freezes the fact, so everything goes in the file. Discover the pattern first: if you can't answer "why this over the alternatives," you don't have a ref yet.

## Cite, change, and adopt a ref

These are all change-unit operations — owned by `references/change.md`, not re-taught here:

- **Cite a ref from a component** (the citation *is* the edge the canvas marks `→ edge: uses`) → change.md §Phase 3.2.
- **Edit an existing ref** — refused (`<id> is a fact — facts are frozen and change only through a change-unit`); ride the edit as a patch in a change-unit and `C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" change apply` → change.md §Phase 3.2.
- **Adoption ADR + the `accepted → done` auto-done latch** → change.md §Status. (Never type or `set` a terminal status; the latch actualizes it.)

## See a ref's reach

```bash
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" graph ref-<slug> --direction reverse              # what cites this ref
C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh" graph ref-<slug> --direction reverse --format mermaid   # as a diagram
```

Reverse graph is the canonical "who uses this" — don't hand-walk citers or read raw `.c3/` files.
Use any `route:` block in that output to pick first inspection anchors, lanes, and drift labels, but
read each citer before claiming it is affected.
