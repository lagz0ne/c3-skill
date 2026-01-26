# Alter Evaluation Design

## Problem

Current alter evaluation uses expectations-based checking (structural). Need outcome-focused evaluation that tests whether alter output enables engineers to use the docs for their intended purposes.

## Core Insight

Alter outputs (ADR, Plan, Component docs) exist for specific purposes. Evaluate by checking if concerns for each purpose are addressed, **grounded in the actual templates**.

---

## Rating Scale (Action-Focused)

| Level | Meaning | Reviewer Action |
|-------|---------|-----------------|
| **reject** | Fundamentally flawed or missing | Block, request rewrite |
| **rework** | Major gaps, can't proceed safely | Return with specific fixes |
| **conditional** | Usable with caveats | Approve with follow-ups |
| **approve** | Solid, ready to proceed | Approve |
| **exemplary** | Reference quality | Approve, use as example |

---

## Purpose-Driven Evaluation

### Purpose 1: Approve the Change

*Engineer reviewing ADR before accepting*

| Concern | What Must Be Addressed |
|---------|------------------------|
| Problem validity | Does stated problem match the change request? Not invented? |
| Decision clarity | Is the decision specific and actionable? |
| Alternatives/Rationale | Are alternatives realistic? (Only if template requires) |
| Scope appropriateness | Not over-scoped? Not under-scoped? |
| Breaking changes/Dependencies | Are impacts identified? |

### Purpose 2: Execute the Change

*Engineer implementing after approval*

| Concern | What Must Be Addressed |
|---------|------------------------|
| File identification | Which files to create/modify? Paths clear? |
| Sequence clarity | Sequence of changes explicit? Dependencies respected? |
| Content detail | What goes in each file? Enough detail to write? |
| Integration/Hand-offs | How does new code connect to existing? |

### Purpose 3: Verify Correctness

*Engineer checking after implementation*

| Concern | What Must Be Addressed |
|---------|------------------------|
| Concrete checks | Verification steps listed? |
| Independence | Can verify without knowing implementation details? |
| Coverage | All affected areas have verification? |

### Purpose 4: Maintain Later

*Future engineer understanding the change*

| Concern | What Must Be Addressed |
|---------|------------------------|
| Rationale traceability | Why was this change made? Link to ADR? |
| Impact tracking | What does this affect? What affects this? |
| Future evolution | (Only if template requires - usually NOT required) |

---

## Key Design Decision: Template Grounding

**Problem:** Initial judge produced false negatives by applying generic "good documentation" standards.

**Solution:** Judge receives templates as input and evaluates against what templates actually require:

```typescript
interface AlterJudgeInput {
  changeRequest: string;
  output: { adr, plan, componentDocs };
  templates: {  // Ground truth
    adrTemplate: string;
    planTemplate: string;
    componentTemplate: string;
  };
}
```

**Impact:**
- Before: 50% (conditional) with 4 false negatives
- After: 75% (approve) with 0 false negatives

**What this means:**
- Only flag missing content if template explicitly requires it
- Don't penalize for "future evolution" if not in template
- Don't penalize for "runtime verification" if change is doc-only
- Don't penalize for implementation details if dependencies documented

---

## Test Results

| Scenario | Type | Verdict | Score | Notes |
|----------|------|---------|-------|-------|
| Add notifications | Feature | approve | 75% | New component, container update |
| Fix session timeout | Bugfix | approve | 75% | Minimal scope, single component |
| Add caching layer | Container | approve | 75% | Cross-cutting, affects multiple components |

All three change types pass with consistent scoring.

---

## Implementation

### Files

| File | Purpose |
|------|---------|
| `eval/lib/alter-judge.ts` | Purpose-driven LLM judge with template grounding |
| `eval/lib/types.ts` | Extended TestCase with `eval_type` and `change_request` |
| `eval/run.ts` | Integrated alter judge for `eval_type: alter` cases |
| `eval/test-alter-judge.ts` | Standalone test with good input |
| `eval/test-alter-judge-bad.ts` | Standalone test with bad input |

### Test Cases

| File | Type | Tests |
|------|------|-------|
| `eval/cases/alter-add-feature.yaml` | Feature | New component integration |
| `eval/cases/alter-bugfix.yaml` | Bugfix | Minimal scope change |
| `eval/cases/alter-container-change.yaml` | Container | Cross-cutting concern |

### Usage

```bash
# Run single alter eval
bun eval/run.ts eval/cases/alter-add-feature.yaml --verbose

# Run all alter cases
bun eval/run.ts eval/cases/alter-*.yaml

# Run standalone judge tests
bun eval/test-alter-judge.ts --verbose
bun eval/test-alter-judge-bad.ts --verbose
```

### Test Case Format

```yaml
name: "Test name"
fixtures: "fixtures/documented-api"
eval_type: alter  # Triggers alter judge

change_request: |  # Ground truth for judge
  The original user request...

command: |
  The prompt sent to the agent...

goal: |
  What should be achieved...

constraints:
  - Constraint 1
```

---

## Learnings

1. **Template grounding is essential** - Generic "good documentation" standards cause false negatives
2. **Purpose-driven > structural** - Test what docs enable, not what sections exist
3. **Change type doesn't affect judge** - Same concerns apply to bugfix, feature, container changes
4. **False negative review is valuable** - Using subagent to review gaps catches judge issues

### From Granularity Testing

5. **Judge correctly handles partial quality** - `rework` verdict for incomplete output, with mix of addressed/partial/missing
6. **Judge catches workflow violations** - When ADR doesn't match request, correctly marks `problem_validity` as missing
7. **Soften structural requirements** - Changed "alternatives table required" to "any reasoning counts" to avoid false positives

### Known Limitations (v1)

| Limitation | Impact | Fix Later |
|------------|--------|-----------|
| No reference validation | Won't catch broken component references | Add cross-reference checking |
| No ID sequencing check | Won't catch non-sequential IDs | Add ID validation |
| Threshold could be stricter | `rework` vs `reject` borderline for fundamental mismatches | Tune aggregation logic |

---

## Granularity Test Results

| Test | Input Quality | Expected | Actual | Pass |
|------|---------------|----------|--------|------|
| Good | Fully correct | approve | approve | ✓ |
| Bad | Fundamentally wrong | reject | reject | ✓ |
| Partial | Some good, some missing | rework/conditional | rework | ✓ |
| Violation | ADR wrong feature | reject/rework | rework | ✓ |

All test scenarios produce appropriate verdicts.
