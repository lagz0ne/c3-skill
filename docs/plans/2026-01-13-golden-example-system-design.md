# Golden Example System Design

## Overview

A reliability improvement system for C3 skill evaluation that uses **golden examples** as quality benchmarks instead of structural checks alone.

## Problem

The original eval system checked for structure (files exist, certain patterns present) but couldn't verify **quality** - whether the output was actually good architecture documentation. This led to:

- Tests passing when output was technically present but poorly structured
- Inconsistent results due to stochastic LLM behavior
- No baseline to compare "good enough" vs "needs improvement"

## Solution: Golden Examples + LLM Similarity Judge

### Golden Examples

A golden example is a captured, refined output that represents the ideal quality for a skill + scenario combination. Each golden includes:

```
eval/goldens/<skill>-<scenario>/
├── GOLDEN.yaml          # Metadata + evaluation dimensions
├── README.md            # Actual C3 docs
├── TOC.md
├── adr/
├── c3-N-*/              # Container directories
└── refs/                # Reference files
```

### GOLDEN.yaml Structure

```yaml
name: "Human-readable name"
skill: "c3:onboard"
category: "greenfield|existing|monorepo"

description: |
  What this golden represents

input_context:
  project_name: "..."
  domain: "..."
  features: [...]
  tech_stack: {...}

evaluation_dimensions:
  structure:
    - "Structural checks"
  content_quality:
    - "Quality checks"
  references:
    - "Ref checks"
  balance:
    - "Balance checks"

created: "YYYY-MM-DD"
source_session: "how it was created"
refinements:
  - "What was fixed before capture"
```

### Capture and Refine Workflow

```
┌─────────────────┐
│ Define scenario │
│ (Q&A session)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Run skill in    │
│ isolated env    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Audit output    │
│ (C3 accuracy)   │
└────────┬────────┘
         │
         ▼
┌───────────────────────┐
│ Issues found?         │
│                       │
│ Yes → Fix → Re-audit  │
│ No  → Capture golden  │
└───────────┬───────────┘
            │
            ▼
┌─────────────────┐
│ Store in        │
│ eval/goldens/   │
└─────────────────┘
```

### LLM Similarity Judge (To Implement)

The judge compares eval output against golden across dimensions:

```typescript
interface JudgeResult {
  overall_score: number;      // 0-100
  dimension_scores: {
    structure: number;
    content_quality: number;
    references: number;
    balance: number;
  };
  gaps: string[];             // Specific missing/incorrect items
  verdict: 'pass' | 'fail';   // Based on threshold
}
```

**Judge Prompt Pattern:**

```
You are evaluating C3 architecture documentation quality.

## Golden Example (The Standard)
<golden content>

## Output to Evaluate
<eval output>

## Evaluation Dimensions
<from GOLDEN.yaml>

For each dimension, score 0-100 and list specific gaps.
```

## Learnings from Session

### What Works

1. **Explicit file requirements in test commands** - Don't rely on skill instructions alone; specify exact files expected in the test case YAML
2. **"Deny with message" hook pattern** - When simulating human interaction, use `permissionDecision: "deny"` with `permissionDecisionReason` containing the answer, forcing agent to continue
3. **Isolated subagent sessions** - Run onboarding in fresh context to avoid pollution from current conversation
4. **C3 audit before capture** - Use the audit skill to verify C3 concept accuracy before saving as golden

### What Doesn't Work

1. **Adding "CRITICAL" instructions to skills** - LLM doesn't reliably follow emphasis markers
2. **"Allow with modified input" hook pattern** - Agent treats this as "done" and may stop
3. **Structural checks alone** - Doesn't catch quality issues

## Golden Examples Roadmap

| Skill | Scenario | Status |
|-------|----------|--------|
| onboard | greenfield-saas | ✅ Captured |
| onboard | simple-api | Pending |
| onboard | monorepo | Pending |
| query | existing-docs | Pending |
| alter | add-feature | Pending |
| audit | stale-refs | Pending |

## Next Steps

1. **Create eval judge** - Build the LLM similarity judge that compares against goldens
2. **Create test case linking** - Map eval cases to their golden examples
3. **Capture remaining goldens** - Run capture-and-refine for other skill/scenario combinations
4. **Threshold tuning** - Determine pass/fail thresholds based on dimension scores
