# Sweep Reference

Impact assessment for proposed changes. Advisory only — does NOT make changes.

## How It Works

```
Load Topology -> Identify Entities -> Parallel Assessment -> Synthesize
```

Use Task tool to spawn anonymous subagents for parallel per-entity assessment.

## Progress Checklist

```
Sweep Progress:
- [ ] Topology loaded
- [ ] Affected entities identified
- [ ] Per-entity assessment complete
- [ ] Constraint chain checked
- [ ] Synthesis delivered
```

---

## Step 1: Load Topology

```bash
bash <skill-dir>/bin/c3x.sh list --json
```

Get full system structure: entities, relationships, frontmatter.

## Step 2: Identify Affected Entities

From the proposed change, determine:
- Which containers are affected?
- Which components within those containers?
- Which refs govern the affected areas?
- Which ADRs are relevant?

Match by: entity title, relationships, code-map entries, ref scopes.

## Step 3: Per-Entity Assessment

For each affected entity, assess impact. Use subagents for parallelism when multiple containers are affected.

### Container Assessment

For each affected container:
1. Read container README (from JSON path)
2. Check: does the change affect container responsibilities?
3. Identify affected components within
4. Check container-level constraints

### Component Assessment

For each affected component:
1. Read component doc
2. Read `.c3/code-map.yaml` for file paths — inspect actual code
3. Check: does the change modify behavior, API, dependencies?
4. Check applicable refs for compliance
5. Identify downstream dependents (components that use this one)

### Ref Assessment

For each applicable ref:
1. Read ref doc
2. Check: does the proposed change comply with or violate conventions?
3. If violation: note severity and override requirements

## Step 4: Constraint Chain

For each affected component, trace constraints upward:
- Component constraints (from doc)
- Container constraints (from parent README)
- Context constraints (from `.c3/README.md`)
- Cited ref constraints (from Related Refs)

Flag any proposed violation.

## Step 5: Synthesize

Combine all assessments into unified report:

1. **Affected Entities** — which containers and components, with reasons
2. **Constraint Chain** — all conventions, refs, ADRs that apply
3. **File Changes** — specific files that would need modification
4. **Risks** — edge cases, relationship impacts, ADR conflicts
5. **Recommended Approach** — step-by-step plan respecting all constraints

---

## Output Format

```
**C3 Impact Assessment**

**Proposed Change:** [summary]

## Affected Entities

| Entity | Type | Impact | Reason |
|--------|------|--------|--------|
| c3-N | container | direct | [why] |
| c3-NNN | component | direct/indirect | [why] |
| ref-X | ref | compliance check | [why] |

## Constraint Chain

| Source | Constraint | Status |
|--------|-----------|--------|
| c3-0 | [system rule] | compliant/violated |
| c3-N | [container rule] | compliant/violated |
| ref-X | [pattern rule] | compliant/violated |

## File Changes Required

| File | Change | Component |
|------|--------|-----------|
| src/path/file.ts | [modification] | c3-NNN |

## Risks

- [Risk 1]: [impact and mitigation]
- [Risk 2]: [impact and mitigation]

## Recommended Approach

1. [Step respecting constraints]
2. [Step with rationale]
```

---

## Routing

- Want to implement after assessment? -> change operation
- Architecture questions -> query operation
- Pattern management -> ref operation
- Standalone audit -> audit operation
