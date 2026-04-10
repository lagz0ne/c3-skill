# Rule + Graph Injection Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the c3 skill automatically load rule content and graph context after every `c3x lookup`, so the LLM absorbs constraints before editing code and verifies compliance after.

**Architecture:** Instruction-only changes to three skill markdown files. No CLI code, no schema changes, no new commands. The `c3x lookup` → `c3x read` → `c3x graph` sequence is encoded in skill instructions that the LLM follows during change and audit operations.

**Tech Stack:** Markdown (skill instruction files)

---

### Task 1: Update SKILL.md — File Context section

**Files:**
- Modify: `skills/c3/SKILL.md:130-137` (the "File Context — MANDATORY" block)

- [ ] **Step 1: Replace the File Context block**

Replace lines 130-137 of `skills/c3/SKILL.md` (the block starting with `**File Context — MANDATORY`  through `No match = uncharted, proceed with caution.`):

Old content:
```markdown
**File Context — MANDATORY before reading or altering any file:**
```bash
bash <skill-dir>/bin/c3x.sh lookup <file-path>
bash <skill-dir>/bin/c3x.sh lookup 'src/auth/**'   # glob for directory-level context
```
Returned refs = hard constraints, every one MUST be honored.
Run the moment any file path surfaces. Use glob when working across a directory.
No match = uncharted, proceed with caution.
```

New content:
````markdown
**File Context — MANDATORY before reading or altering any file:**

**Step 1 — Lookup:**
```bash
bash <skill-dir>/bin/c3x.sh lookup <file-path>
bash <skill-dir>/bin/c3x.sh lookup 'src/auth/**'   # glob for directory-level context
```
Run the moment any file path surfaces. Use glob when working across a directory.
No match = uncharted, proceed with caution.

**Step 2 — Load rules:** For every `rule-*` returned by lookup:
```bash
bash <skill-dir>/bin/c3x.sh read <rule-id>
```
Extract `## Rule`, `## Golden Example`, and `## Not This`. These are hard constraints — code MUST match the golden pattern. Deviations require an Override section in the rule or a new ADR.

**Step 3 — Graph context:** For the first component returned (or each, if few):
```bash
bash <skill-dir>/bin/c3x.sh graph <component-id> --depth 1
```
Shows providers (what this component depends on) and consumers (what depends on it). If your change affects the component's interface, check consumers before proceeding.

**Result:** Returned refs + loaded rule content + graph = the full constraint set. All MUST be honored.
````

- [ ] **Step 2: Verify the edit**

Open the file and confirm:
- The old 7-line block is gone
- The new block has three numbered steps (Lookup, Load rules, Graph context)
- No duplicate content — the old "Returned refs = hard constraints" single line is replaced by the richer "Result:" line
- Surrounding content (the "Layer Navigation" heading above and "Graph Output" section below) is unchanged

- [ ] **Step 3: Commit**

```bash
git add skills/c3/SKILL.md
git commit -m "feat(skill): add rule loading + graph context to File Context gate

After c3x lookup, the LLM now auto-reads each returned rule and
graphs the component before touching code. This turns ID-only
pointers into loaded constraints in the LLM's working context."
```

---

### Task 2: Update change.md — Phase 3 (Execute)

**Files:**
- Modify: `skills/c3/references/change.md:81-88` (the Phase 3 lookup block)

- [ ] **Step 1: Expand the Phase 3 lookup block**

Replace lines 81-88 of `skills/c3/references/change.md` (from `**REQUIRED before touching any file:**` through `Per task: verify code correct, docs updated (code-map entries, Related Refs), no regressions.`):

Old content:
```markdown
**REQUIRED before touching any file:**
```bash
bash <skill-dir>/bin/c3x.sh lookup <file-path>
```
Returned refs = hard constraints. Every one must be honored. No exceptions.

Parallel subagents: decompose tasks, each runs `c3x read` on component docs + refs before touching code.

Per task: verify code correct, docs updated (code-map entries, Related Refs), no regressions.
```

New content:
````markdown
**REQUIRED before touching any file — the 3-step context gate:**

**1. Lookup:**
```bash
bash <skill-dir>/bin/c3x.sh lookup <file-path>
```

**2. Load rules:** For every `rule-*` in the lookup response:
```bash
bash <skill-dir>/bin/c3x.sh read <rule-id>
```
Extract `## Rule`, `## Golden Example`, and `## Not This`. Hold these in working context while editing. Code MUST match the golden pattern — deviations require the rule's `## Override` process or a new ADR.

**3. Graph blast radius:** For each matched component:
```bash
bash <skill-dir>/bin/c3x.sh graph <component-id> --depth 1
```
If changes affect the component's interface, check consumers before proceeding.

Returned refs + loaded rule content + graph = hard constraints. No exceptions.

Parallel subagents: decompose tasks, each runs the 3-step gate on their files before touching code.

Per task: verify code correct, docs updated (code-map entries, Related Refs/Rules), no regressions.
````

- [ ] **Step 2: Verify the edit**

Open the file and confirm:
- Phase 3 heading and scaffold commands above are untouched
- The new block has three numbered sub-steps matching SKILL.md's pattern
- Phase 3b below is untouched

- [ ] **Step 3: Commit**

```bash
git add skills/c3/references/change.md
git commit -m "feat(change): expand Phase 3 lookup into 3-step context gate

Phase 3 now mandates: lookup → read each rule → graph component.
Subagents must run the same gate on their files."
```

---

### Task 3: Update change.md — Phase 3b (Ref Compliance Gate)

**Files:**
- Modify: `skills/c3/references/change.md:109` (the Rule Compliance paragraph in Phase 3b)

- [ ] **Step 1: Connect Phase 3b to Phase 3's loaded rules**

Replace the single line at line 109 of `skills/c3/references/change.md`:

Old content:
```markdown
**Rule Compliance (strict):** For each `rule-*` entity returned by lookup, compare code against `## Golden Example` and `## Not This` for exact compliance. Rules use strict enforcement — code must match the golden pattern. Flag any deviation as a violation.
```

New content:
```markdown
**Rule Compliance (strict):** For each `rule-*` entity returned by lookup, compare code against the `## Golden Example` and `## Not This` content loaded in Phase 3 step 2. Do NOT re-read — the constraints are already in context. Rules use strict enforcement — code must match the golden pattern. Flag any deviation as a violation.
```

- [ ] **Step 2: Verify the edit**

Open the file and confirm:
- Only the Rule Compliance paragraph changed
- It now references "Phase 3 step 2" and says "Do NOT re-read"
- The ref compliance table above and adversarial framing below are untouched

- [ ] **Step 3: Commit**

```bash
git add skills/c3/references/change.md
git commit -m "feat(change): connect Phase 3b rule check to Phase 3 loaded content

Phase 3b now references the rule content already loaded in Phase 3
step 2, avoiding redundant c3x read calls."
```

---

### Task 4: Update audit.md — Phase 7b (Rule Compliance)

**Files:**
- Modify: `skills/c3/references/audit.md:103-115` (the Rule Compliance sub-section of Phase 7b)

- [ ] **Step 1: Enhance Phase 7b Rule Compliance with explicit read + compliance questions**

Replace lines 103-115 of `skills/c3/references/audit.md` (from `**Rule Compliance:**` through the `Rules use STRICT enforcement...` line):

Old content:
```markdown
**Rule Compliance:** For each rule with `## Golden Example`:
1. Find citing components via `c3x list --json`
2. For each citing component, spot-check 1-2 mapped files from code-map
3. Compare code against `## Golden Example` pattern

| Result | Meaning |
|--------|---------|
| COMPLIANT | Code matches golden pattern exactly |
| VIOLATION | Code deviates from required pattern |
| INCOMPLETE | Rule lacks Golden Example section |
| NOT CHECKED | No code-map mapping or no citing components |

Rules use STRICT enforcement (code must match golden pattern exactly), unlike refs which use directional alignment. A rule VIOLATION is always FAIL severity — rules are non-negotiable constraints.
```

New content:
```markdown
**Rule Compliance:** For each rule with `## Golden Example`:
1. **Load rule content:**
   ```bash
   bash <skill-dir>/bin/c3x.sh read <rule-id>
   ```
   Extract `## Rule`, `## Golden Example`, and `## Not This` sections.
2. **Derive compliance questions:** From `## Rule` + `## Golden Example`, write 1-3 YES/NO questions that code must satisfy (e.g., "Does the error return use CmdError struct?" / "Is slog used with component context?").
   - If questions can't be derived → WARN: rule is too vague for enforcement, needs rework.
3. Find citing components via `c3x list --json`
4. For each citing component, spot-check 1-2 mapped files from code-map
5. Apply the YES/NO questions to the spot-checked code

| Result | Meaning |
|--------|---------|
| COMPLIANT | All compliance questions answered YES |
| VIOLATION | One or more questions answered NO |
| INCOMPLETE | Rule lacks Golden Example or questions can't be derived |
| NOT CHECKED | No code-map mapping or no citing components |

Rules use STRICT enforcement (code must match golden pattern exactly), unlike refs which use directional alignment. A rule VIOLATION is always FAIL severity — rules are non-negotiable constraints.
```

- [ ] **Step 2: Verify the edit**

Open the file and confirm:
- The ref compliance section above (lines 87-101) is untouched
- The rule compliance section now has 5 numbered steps instead of 3
- Steps 1-2 are new (load rule content, derive compliance questions)
- Steps 3-5 map to the old steps 1-3
- The result table and severity note are preserved

- [ ] **Step 3: Commit**

```bash
git add skills/c3/references/audit.md
git commit -m "feat(audit): add rule reading + compliance questions to Phase 7b

Audit Phase 7b now loads rule content via c3x read and derives
YES/NO compliance questions before spot-checking code. Rules that
can't produce compliance questions are flagged as too vague."
```

---

### Task 5: End-to-end verification

**Files:**
- Read: `skills/c3/SKILL.md`, `skills/c3/references/change.md`, `skills/c3/references/audit.md`

- [ ] **Step 1: Cross-reference consistency**

Read all three modified files and verify:
1. SKILL.md File Context references "Step 1/2/3" (Lookup, Load rules, Graph)
2. change.md Phase 3 uses the same "1/2/3" numbering and same commands
3. change.md Phase 3b references "Phase 3 step 2"
4. audit.md Phase 7b step 1 uses the same `c3x read <rule-id>` command
5. No orphan references to old single-line "Returned refs = hard constraints"

- [ ] **Step 2: Verify no unintended changes**

```bash
git diff HEAD~4 --stat
```

Expected: exactly 3 files changed:
- `skills/c3/SKILL.md`
- `skills/c3/references/change.md`
- `skills/c3/references/audit.md`

- [ ] **Step 3: Run c3x check to ensure no structural issues**

```bash
C3X_MODE=agent bash <skill-dir>/bin/c3x.sh check
```

Expected: no errors (these are markdown instruction files, not `.c3/` entity docs, but verify the skill still loads cleanly).
