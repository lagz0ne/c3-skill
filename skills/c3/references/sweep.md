# Sweep Reference

Impact assessment. Advisory only — no changes.

Flow: `Topology → Affected Entities → Parallel Assessment → Synthesize`

Spawn subagents via Task tool for parallel per-entity assessment.

## Progress

- [ ] Topology loaded
- [ ] Affected entities identified
- [ ] Per-entity assessment complete
- [ ] Constraint chain checked
- [ ] Synthesis delivered

---

## Step 1: Topology

```bash
bash <skill-dir>/bin/c3x.sh list --json
```

## Step 2: Affected Entities

From proposed change, identify:
- Containers, components, refs, ADRs
- Match by: entity title, relationships, code-map entries, ref scopes

## Step 3: Per-Entity Assessment

Use subagents for parallelism when multiple containers affected.

**Container:** Read README → does change affect responsibilities? → identify affected components.

**Component:**
1. Read component doc
2. For each file in code-map: `c3x lookup <file>` — loads constraint chain before inspecting code
3. Check code against constraints
4. Does change modify behavior, API, dependencies?
5. Check applicable refs. Identify downstream dependents.

**Ref:** Read ref → does proposed change comply or violate? → note severity + override requirements.

## Step 4: Constraint Chain

For each affected component, trace upward:
- Component constraints → container → context → cited refs

Flag any proposed violation.

## Step 5: Synthesize

```
**C3 Impact Assessment**

**Proposed Change:** [summary]

## Affected Entities
| Entity | Type | Impact | Reason |
|--------|------|--------|--------|
| c3-N | container | direct | [why] |

## Constraint Chain
| Source | Constraint | Status |
|--------|-----------|--------|
| c3-0 | [rule] | compliant/violated |

## File Changes Required
| File | Change | Component |
|------|--------|-----------|
| src/path/file.ts | [mod] | c3-NNN |

## Risks
- [Risk]: [impact + mitigation]

## Recommended Approach
1. [Step respecting constraints]
```

---

## Routing

- Implement after assessment → change
- Architecture questions → query
- Pattern management → ref
- Standalone audit → audit
