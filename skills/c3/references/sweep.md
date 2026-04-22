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
bash <skill-dir>/bin/c3x.sh list
```

## Step 2: Affected Entities

From proposed change, identify:
- Containers, components, refs, ADRs
- Match by: title, relationships, code-map entries, ref scopes

## Step 3: Per-Entity Assessment

Subagents for parallelism when multiple containers affected.

**Container:** `c3x read <container-id>` → change affect responsibilities? → identify affected components.

**Component:**
1. `c3x read <component-id>`
2. Per code-map file: `c3x lookup <file>` — loads constraint chain before inspecting code
3. Check code against constraints
4. Change modify behavior, API, dependencies?
5. Check applicable refs. Identify downstream dependents.

**Ref:** `c3x read <ref-id>` → proposed change comply or violate? → note severity + override requirements.

**Rule:** `c3x read <rule-id>` → proposed change violate golden pattern? Note severity + remediation.

## Step 4: Constraint Chain

Per affected component, trace upward:
- Component constraints → container → context → cited refs + rules

Flag any proposed violation.

## Step 5: Synthesize

**Impact Graph:** Include `c3x graph <target-entity> --direction reverse --format mermaid` as mermaid code block atop report (reverse direction = who depends on the changed entity). Graph from most specific affected entity (component > container). For ref/rule impact, graph ref/rule itself to show all citers.

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

