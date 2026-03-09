# Audit Reference

Validate C3 docs for consistency, drift, and completeness.

Two stages: **automated scan** (CLI tools, fast) → **deep exploration** (reasoning, thorough).

## Progress

- [ ] Stage 1: Automated scan (c3x check + list + coverage)
- [ ] Stage 2: Deep exploration against goals
- [ ] Stage 3: Synthesis and recommendations

---

## Stage 1: Automated Scan

Run these CLI commands to establish a factual baseline:

```bash
bash <skill-dir>/bin/c3x.sh check --json          # structural validation
bash <skill-dir>/bin/c3x.sh list --json            # full entity inventory
bash <skill-dir>/bin/c3x.sh coverage               # code-map coverage stats
```

Record the results. `check` catches broken links, orphans, duplicate IDs, missing fields. `coverage` shows mapped vs unmapped files. These are your starting facts — don't repeat checks that the CLI already covered.

---

## Stage 2: Deep Exploration

With the inventory and structural results in hand, investigate each of these goals. Spend real time on each — read source code, grep for patterns, cross-reference claims. The point is to find things the CLI can't catch.

### Goal A: Do docs match reality?

Compare what the docs say against what the code actually does:
- Do container/component descriptions match the code's actual behavior?
- Are CLI commands, API endpoints, or key features documented accurately?
- Do code-map entries point to files that still exist and still serve that purpose?
- Are there significant code modules or features with no corresponding docs?

Use `c3x lookup <file>` to spot-check whether files map to the right components.

### Goal B: Are refs (patterns) accurate and wired?

For each ref:
- Does it have Choice + Why sections? (Required minimum)
- Is it cited by at least one component? (Orphan refs = WARN)
- Does the described pattern match how the code actually implements it?
- Has the underlying tech or approach changed since the ref was written?

### Goal C: Are boundaries respected?

- Do components stay within their container's responsibility?
- Are there cross-container imports that violate boundaries?
- Is orchestration logic in the right layer (container, not component)?
- Do components duplicate ref content instead of citing it?

### Goal D: Is anything stale or missing?

- Recent code changes not reflected in docs
- Deleted features still documented
- New modules/packages with no component docs
- CI/CD workflow changes not reflected in relevant refs
- ADRs stuck in `accepted` status without being implemented

### Goal E: CLAUDE.md integration

- Do relevant directories have `<!-- c3-generated -->` blocks?
- Do existing blocks reference correct component IDs?
- Any orphan blocks pointing to deleted components?

**Dig deeper on anomalies.** If something looks off during any goal, investigate it fully. Read the actual source code, check git history if needed, verify claims against reality. The best audit findings come from following threads that don't add up.

**Code navigation: LSP first.** When verifying code against docs, use LSP tools (go-to-definition, find-references, hover) to precisely trace function calls, type definitions, and usages. Only fall back to Grep/Glob when LSP is unavailable. LSP catches discrepancies that text search misses.

---

## Stage 3: Synthesis

Present findings in a structured report:

```
**C3 Audit Results**

| Area | Status | Issues |
|------|--------|--------|
| Structural (CLI) | PASS/WARN/FAIL | [from c3x check] |
| Code-map coverage | PASS/WARN/FAIL | [from c3x coverage] |
| Docs vs reality | PASS/WARN/FAIL | [from Goal A] |
| Ref validation | PASS/WARN/FAIL | [from Goal B] |
| Boundaries | PASS/WARN/FAIL | [from Goal C] |
| Freshness | PASS/WARN/FAIL | [from Goal D] |
| CLAUDE.md | PASS/WARN/FAIL | [from Goal E] |

**Summary:** N passes, M warnings, K failures

**Action Items:**
1. [specific fix with concrete steps]
2. ...
```

Severity guide:
- **FAIL**: Factual error, broken reference, or docs contradict code behavior
- **WARN**: Missing coverage, stale description, orphan entity
- **PASS**: No issues found

---

## Audit Scope

| Scope | Focus |
|-------|-------|
| Full | All layers, all goals |
| Single container | Goals A-D scoped to one container |
| Targeted | Specific goal or area only |

---

## Drift Resolution

| Situation | Action |
|-----------|--------|
| Code changed, docs outdated | Create ADR, update docs |
| Docs describe removed code | Remove stale sections |
| New module not in inventory | Add to inventory |
| Orphan ADR (accepted, never implemented) | Close with reason |

Intentional arch change → ADR. Doc rot → direct fix.
