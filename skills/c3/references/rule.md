# Rule

A **rule** is a frozen fact whose identity is an **enforceable standard** — a project-wide invariant a reviewer can check a file against, with the canonical code to check it by. Shape: `## Goal` (the standard) · `## Rule` (one-line, present-tense, enforceable) · `## Golden Example` (literal code from a real file). `## Not This` / `## Scope` / `## Override` are optional. Run `c3 schema rule` for the full contract — it leads with `REJECT IF:` bullets that are the rejection floor (a paraphrased Golden Example, a multi-clause Rule, or a Goal that describes one component instead of a standard all fail).

**Rule vs ref** is one test, owned by `references/ref.md`: strip the rationale and ask if the doc still tells you *what must be true*. Yes → rule (enforcement). No → ref (the rationale was the point). When in doubt, run it.

Two operations.

## Create a rule

A rule is a fact, so it **freezes the moment it has a body** — there is no scaffold-then-fill. Find the canonical code first (if no real instance exists, the rule is premature: author the instance, then extract the rule). Assemble the whole body — `## Goal`, `## Rule`, `## Golden Example`, optional sections, code fences and all — into one file and create it in a single call:

```bash
c3 schema rule > body.md     # the contract to draft against
# ...author the body...
c3 add rule structured-logging --file body.md
```

`c3 add` is the unguarded create path. After it lands, `c3 write` / `c3 set` on the rule are **refused** (see the frozen-fact contract in SKILL.md) — so `origin:`, `scope:`, and every field go **in the body file**, never a follow-up `set`. The lone exception is `c3 set rule-<slug> codemap '<glob>'`: the code binding is verified, not frozen, and may be set directly.

**Cite the rule from each component it governs.** A citation is a row in the component's edge column — for a freshly-seeded project that's the `Governance` table, `Reference` column (the one tagged `edge: uses`), with `Type: rule`. Ask `c3 schema component` where that column lives rather than memorizing it. A brand-new citer carries the row in its body at `c3 add` time; an existing citer is a frozen fact, so the citation rides a change-unit patch — that whole flow (cite the block, author the patch, apply) is owned by `references/change.md`.

## Change a rule

Editing, deprecating, or re-citing an existing rule is a **change-unit** — author patches in `.c3/changes/<unit-id>/`, then `c3 change apply`. The full saga (cite → `change new` → patch → `apply`), the adoption ADR, and its `accepted → done` latch are owned by `references/change.md`. Do not re-walk them here; classify the work as a change and hand off.

## Migrate refs to rules

When auditing existing refs for rule candidates (or splitting a dual-nature ref):

1. **Classify** each ref with the Separation Test (`references/ref.md`). Pure rationale → leave it. Tells you what to do → convert. Both → split: narrow the ref to rationale, extract a new rule with `origin: [ref-<slug>]` in its body.
2. **Author the rule(s) whole** via `c3 add rule … --file` (above).
3. **Rewire and retire in one change-unit.** Re-cite each affected component to the rule, and if a ref is being removed, retire it in the *same* unit. The destruction gate refuses a retire that would orphan a child or dangle a citer unless that same unit heals it — so you cannot leave a dangling citation behind; you do not hand-check for orphans, the gate does. Mechanics: `references/change.md`.

To find what a rule governs and what a migration touches, lead with `c3 search "<concept>"` and `c3 graph rule-<slug> --direction reverse` (add `--format mermaid` for a citation diagram) — never read `.c3/` files directly.
