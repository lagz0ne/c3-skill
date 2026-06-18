# Dogfood findings — onboarding a fresh doc-only project into c3

A subagent onboarded a lean stage-1 **single-user TODO web app** (from the matched-abstraction
`todo-web-app-complexity-growth` eval topic) into a fresh `c3 init` project, using the local CLI.
End state: 5 facts (c3-0 → {web, api}; web → task-list-ui; api → {task-api, task-store}),
`c3 check` green, lean rung-1 throughout. Reaching that took ~8 source-diving detours. The friction
is the deliverable. Ranked by severity.

## A. Switch-gated double-V is too strict for docs-ahead-of-code (the biggest)
1. **The Derived Materials inspection gate makes a doc-only project unbuildable.** A non-`N.A`
   `Material` row is a derivation obligation; `change apply` refuses to land it without a
   territory-grounded `*.inspect.md` whose code-map resolves to **real files**. With no code yet
   (onboarding / docs-first), that's impossible — the only escape is `N.A` on every Material cell,
   hollowing out a *required* section. **And the gate is invisible in `onboard.md`/`change.md`.**
   This is a gap in the just-shipped feature: it assumes code exists. Fix: when a touched fact has
   obligations but **no resolved code-map territory**, treat it as docs-ahead-of-code — **defer/warn,
   don't hard-block** (the inspection fires later when the code is bound + changed). Reserve the
   hard gate for facts that DO have territory. Also: document the gate in the onboarding flow.

## B. Genesis-flip flow: docs describe a different flow than the CLI does
2. **create-first-to-learn-ids is impossible in the one-flip genesis model** — no pre-apply id
   reveal, so parent `Components`/`Containers` tables need `PENDING` placeholders + a second
   fix-up ADR. `onboard.md §1.2` warns against exactly the churn the flow forces.
3. **Create-patch `target:` becomes the literal id — there is no `c3-101`/`c3-110` numbering.**
   Ids became the slugs (`web`, `api`, `task-api`). `onboard.md` describes numbered allocation; the
   two id models are never reconciled.
9. **Genesis ADR filename ≠ internal id**: `c3 init` writes `adr-20260618-c3-adoption.md` with
   `id: adr-00000000-c3-adoption` inside; commands key off the internal id, so the dated filename
   misleads.
10. **"resumable ledger / staged patches persist" is false** — `change apply` deletes the patch
    files; only the ADR prose survives.
11. **"`check --fix` auto-dones the genesis ADR" isn't automatic** — the latch keys on Affected
    Topology Evidence cites, which a genesis ADR must author as `N.A` (facts don't exist yet), so it
    stays `accepted` until you rewrite Evidence with real cites post-flip (a step never documented).
12. **`write <adr> --section` can't create a section on the bodyless genesis ADR** — must write the
    whole body once first; `onboard.md §0.1` says otherwise.
13. Genesis ADR is **guaranteed noisy** (unknown-entity / N.A-evidence warnings) until the post-flip
    Evidence rewrite.

## C. Table editing via patch is finicky + undocumented (explains the c3-104 retire blocker)
4. **ADR Affected Topology Evidence (`cite`) cannot cite a table-row block** — the required
   `"exact snippet"` contains `|` pipes that break the ADR's own table; escaping makes it "not
   exact." Only text-block sections (no pipes) are citable. (= my c3-104 finding #4.)
5. **A `block` patch on a table row must be the bare inner cells (`a | b | c`)** — no outer pipes,
   no header. Natural attempts ("| a | b |", full header+row) fail "invalid required table." Only
   discoverable from `parse.go`/`render.go` internals. **This explains the c3-104 retire blocker** —
   my README patch supplied a full table, not the bare-cell row form.
6. **Block cite handles are invalidated by node-id renumbering even when the hash is identical** —
   `web#n28`→`web#n115` (same sha256), apply rejects "anchor block not found; rebase." The hash-drift
   design doesn't protect against the integer node-id changing.

## D. Real bugs
7. **`c3 check` reports `BROKEN_SEAL` on staged `*.patch.md` files** — `snapshotCanonicalTree`
   walks every `.md` under `.c3/` incl. staged patches (no seal → BROKEN_SEAL → "run repair").
   Mid-change-unit `check` looks alarmingly broken. Staged change-unit files should be exempt.
8. **The placeholder denylist bans "TODO" and "later"** — `\b(TBD|TODO|maybe|optional|later|if
   applicable)\b`. A **TODO app cannot write "TODO"**; "later" is natural architecture prose
   ("later rungs"). Both rejected in strict sections.

## E. Cosmetic
14. Slug-as-id + title-slugification mangles filenames (`.c3/api-api-/task-api-task-api-.md`).
15. `graph --format mermaid` is ignored under `C3X_MODE=agent` (returns TOON).
16. `read <id> --full --include-adr` together → "entity not found" (each flag alone works).

## What fit well
Schemas are excellent + self-describing; lean-vs-full rung sizing was clear and never pressured
enrichment; `change apply` dry-run + atomic all-or-nothing worked; the frozen-fact guard + creation
window are coherent; a doc-only project CAN reach genuine zero-issue `check`.

## Verdict
The lean rung-1 **shape** fits a fresh simple project; the **doc-only / genesis-flip path** fought
hard. Top fixes: (A1) docs-ahead-of-code path for the inspection gate + document it; (B) reconcile
`onboard.md`'s id/ledger/auto-done claims with actual genesis behavior; (C) document + smooth
table-row patching/citing (also unblocks the c3-104 retire); (D7) exempt staged patches from the
seal walk; (D8) loosen the placeholder denylist (word-boundary false positives).
