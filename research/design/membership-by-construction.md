# Membership by construction

## Problem

A parent↔child membership is stored **twice**: the child's `parent:` frontmatter
(the DB truth — `Entity.ParentID`, used by graph/numbering/everything) *and* a
hand-authored row in the parent's membership table (`## Components` for a
container, `## Containers` for a system/context). `checkLayerDisconnectsStore`
(`cli/cmd/check_enhanced.go`) exists only to police divergence between them, and
the eval's `grow-todo-app` lap-1 hit it as 8 `layer disconnect` WARN: the agent
set `parent:` on each child but could not (lap-1) / had to separately author
(lap-2) the matching parent row.

## Model (aligned with user)

> "change should be a saga, all or nothing … once we enforce that, the integrity
> of the frozen is always correct, so the rest is just ceremony — do what's most
> optimal because nobody cares about things in the middle."

- The change-unit is already an **atomic saga** (`Apply` runs every patch + the
  membership reconcile in one `WithTx`; all-or-nothing).
- **Enforce frozen integrity at the switch.** The committed result must be
  membership-consistent, or the unit doesn't commit.
- The committed boundary is the only observed state, so the middle is ceremony:
  **synthesize the parent rows on apply** from the children's `parent:` edges.
  The agent declares `parent:` *once*; the unit makes the frozen result correct.
- `checkLayerDisconnectsStore` becomes a **should-never-fire assert** (divergence
  is now impossible by construction), retained as defense-in-depth.

## The derived/authored split (column policy)

The real table is `| ID | Name | Category | Status | Goal Contribution |`. Per
column, matched case-insensitively on the trimmed header:

- **Identity (derived, always refreshed from the child entity):** `ID`←child.ID
  (the row key), `Name`/`Title`←child.Title, `Category`←child.Category,
  `Status`←child.Status. These can never go stale.
- **Descriptive (authored, preserved):** any other column (e.g.
  `Goal Contribution`, `Responsibility`). Preserved across reconciles (matched by
  ID); a newly-synthesized row defaults it to the child's Goal-section first line
  (fallback child.Title), so the row is born canvas-valid with zero authoring.
  The author may refine it with a block patch; the refinement is preserved.

## Reconcile (mechanism)

`changeset.ReconcileMembershipBody(s, parentID, section, childType) (changed, err)`:

1. Read the parent body; `ExtractTableFromSection(body, section)`. No table ⇒ no-op.
2. `children = s.Children(parentID)` of `childType`.
3. Rebuild rows keyed by the ID column: keep existing rows whose ID still maps to
   a child (refresh identity columns, preserve descriptive), **in existing order**;
   append rows for new children (id order); drop orphan rows.
4. If the table changed: `SetTableInSection` → `WriteEntity` (replaces nodes +
   reseals). Idempotent — an already-correct parent costs nothing.

## Wiring

- `changeset.Apply` gains a bundled `*ApplyHooks{ SyncEdges, ReconcileMembership }`
  (replaces the bare `syncEdges` param; `nil` hooks skip — so test callers passing
  `nil` are unaffected, only `change.go`/`overlay.go` pass the struct).
- Apply computes **affected parents** in-tx: the old parent of every
  parent-changing (`frontmatter`+`parent:`) / `retire` patch (pre-captured before
  the patch overwrites `ParentID`), the post-apply `ParentID` of every touched
  entity, and every touched entity itself. Each is reconciled via the hook.
- The hook is cmd-provided (`membershipReconciler(c3Dir)`): reconcile, then
  **canvas-validate** the parent's new body — a violation fails the tx (integrity
  at the boundary). Passed to both the real apply and the preview overlay so the
  maintained tables show in `graph --unit` / `change inspect`.

## Canvas: header-only membership tables are valid

A required table can't be header-only under the canvas gate ("empty required table").
But a membership table's rows are tool-owned, so the author can't pre-fill them — and
shouldn't. Fix: `isToolMaintainedTable(type, section)` exempts the membership section
(Containers/Components) from the **0-rows** rule in both the canvas gate
(`validateBodyContentWithDefinition`) and `c3 check` — the header row stays required
(the reconciler fills into it; `change scaffold` already emits exactly that), only the
data-rows requirement is lifted. Identity columns covered: id, name/title, category,
status, **boundary**.

## Agent-facing consequence

Membership becomes a single declaration: set the child's `parent:` (a `frontmatter`
patch, or `parent:` in a `whole` create). The parent's row appears, fully formed,
in the same atomic unit. No insert-row bookkeeping, no drift to dodge. The
insert-row primitive (`3f65303`) remains for genuinely author-owned tables.

## Validation

Go unit tests on the reconciler (presence, orphan-drop, reparent old-parent heal,
identity-refresh, descriptive-preserve, idempotency) + the gate path; then a
`grow-todo-app` eval lap — the agent should reach 0 `layer disconnect` with **no**
membership row patches authored.
