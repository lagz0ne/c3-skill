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
- Match by: entity title, relationships, code-map entries, ref scopes

## Step 3: Per-Entity Assessment

Use subagents for parallelism when multiple containers affected.

**Container:** `c3x read <container-id>` → does change affect responsibilities? → identify affected components.

**Component:**
1. `c3x read <component-id>`
2. For each file in code-map: `c3x lookup <file>` — loads constraint chain before inspecting code
3. Check code against constraints
4. Does change modify behavior, API, dependencies?
5. Check applicable refs. Identify downstream dependents.

**Ref:** `c3x read <ref-id>` → does proposed change comply or violate? → note severity + override requirements.

**Rule:** `c3x read <rule-id>` → does proposed change violate the golden pattern? Note severity + remediation.

## Step 4: Constraint Chain

For each affected component, trace upward:
- Component constraints → container → context → cited refs + rules

Flag any proposed violation.

## Step 5: Synthesize

**Impact Graph:** Include `c3x graph <target-entity> --direction forward --format mermaid` output as a mermaid code block at the top of the report. Graph from the most specific affected entity (component > container). For ref/rule impact, graph the ref/rule itself to show all citers.

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
