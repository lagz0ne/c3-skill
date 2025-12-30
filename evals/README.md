# C3 Skill Evaluation Harness

Iteratively improve C3 skills through automated evaluation.

## Quick Start

```bash
# 1. Install dependencies
bun install

# 2. Add a codebase to test against (git subtree)
git subtree add --prefix evals/codebases/sample-api https://github.com/you/sample-api main --squash

# 3. Copy skill to workbench for iteration
cp -r ../skills/c3/* workbench/skills/c3/
cp -r ../agents/c3.md workbench/agents/

# 4. Run evaluation
./runner.sh adopt sample-api

# 5. Review results, tweak skill, repeat
# When satisfied, copy back to main skills/
```

## Directory Structure

```
evals/
├── codebases/           # Git subtrees of real projects to test against
├── scenarios/           # Test scenario definitions
│   ├── adopt.yaml       # ADOPT scenario (new project)
│   └── adjust.yaml      # ADJUST scenario (create ADR)
├── checklists/          # Structural requirements (Phase 1)
│   ├── adopt.yaml
│   └── adjust.yaml
├── rubric.yaml          # Quality scoring criteria (Phase 2)
├── workbench/           # Plugin with skill under test
│   ├── .claude-plugin/
│   ├── hooks/           # File operation tracking
│   ├── scripts/         # Hook scripts (Bun)
│   ├── skills/c3/       # Skill being iterated
│   └── agents/
├── scripts/             # Evaluation scripts
│   ├── check.ts         # Checklist runner
│   └── evaluate.ts      # Rubric scorer
├── results/             # Eval run outputs (git-ignored)
└── runner.sh            # Main entry point
```

## The Loop

```
┌─────────────────────────────────────────────┐
│ 1. Copy skill to workbench/skills/c3/       │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│ 2. Run: ./runner.sh <scenario> <codebase>   │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│ 3. Phase 1: Checklist (automated)           │
│    - Structural requirements                │
│    - Pass/fail per check                    │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│ 4. Phase 2: Rubric (LLM-scored)             │
│    - Quality assessment                     │
│    - Weighted score 0-100%                  │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│ 5. Review results/run-*/summary.json        │
│    - If score low → tweak skill → goto 2   │
│    - If score good → copy to main skills/   │
└─────────────────────────────────────────────┘
```

## Commands

### Quick Feedback (Recommended Start)

```bash
# Run eval and get feedback
./runner.sh adopt spree --skip-rubric
./feedback.sh
```

### Interactive Loop

```bash
# Start improvement loop
./loop.sh adopt spree

# With target score
./loop.sh adopt spree --target 80

# Auto-improve mode (uses Claude to fix skill)
./loop.sh adopt spree --auto --max-iter 3
```

### Single Evaluation

```bash
# Basic
./runner.sh adopt sample-api

# With options
./runner.sh adopt sample-api --model opus --budget 1.00

# Skip rubric (faster iteration)
./runner.sh adopt sample-api --skip-rubric
```

### Run Checklist Only

```bash
bun scripts/check.ts checklists/adopt.yaml results/run-xxx/c3-output
```

### Run Rubric Only

```bash
bun scripts/evaluate.ts results/run-xxx
```

## Adding Codebases

Use git subtree to add real projects:

```bash
# Add a codebase
git subtree add --prefix evals/codebases/my-project https://github.com/org/my-project main --squash

# Update a codebase
git subtree pull --prefix evals/codebases/my-project https://github.com/org/my-project main --squash
```

## Results

Each run creates:

```
results/run-YYYYMMDD-HHMMSS/
├── summary.json           # Overall result
├── conversation.json      # Raw Claude output
├── session.json           # Session metadata from hooks
├── file-operations.jsonl  # All file operations logged
├── file-list.txt          # Files created in .c3/
├── c3-output/             # Snapshot of .c3/ directory
├── checklist-result.json  # Phase 1 results
└── rubric-result.json     # Phase 2 results
```

## Customization

### Adding Scenarios

Create `scenarios/<name>.yaml`:

```yaml
name: my-scenario
description: What this tests

prompt: |
  The prompt to send to Claude

setup:
  - rm -rf .c3/  # Commands to run before

checklist: my-checklist.yaml
```

### Adding Checklists

Create `checklists/<name>.yaml`:

```yaml
name: my-checklist
checks:
  - id: check-id
    type: path_exists|glob_exists|frontmatter|sections|regex|...
    path: relative/path
    # ... type-specific options
```

### Modifying Rubric

Edit `rubric.yaml` to change quality criteria and weights.
