---
name: codex
description: >
  Delegate well-scoped coding, analysis, and review tasks to the OpenAI Codex CLI
  (gpt-5.5) as a cross-model collaborator. Use for implementation chunks you want a
  second model to write or cross-check, independent second-opinion review of a design
  or diff, and applying Codex-authored changes. Triggers: "ask codex", "have codex
  implement/review X", "second opinion from codex", "cross-check with codex".
---

# Codex — cross-model delegation

`codex` (OpenAI Codex CLI, model `gpt-5.5`, `approval_policy = never` so it never blocks
on prompts) is a second coding agent that reads this repo and reasons over it precisely
(verified). Consult it for a different model's perspective — it catches what a single
model misses.

## When to consult Codex
- A well-scoped implementation chunk you want a second model to write, or to cross-check
  your own implementation against.
- An independent design/diff review — different model, different failure modes caught.
- Adversarial verification of a claim ("is X actually true in the shipped code?").

## How — always non-interactive `codex exec`
| Goal | Command |
|------|---------|
| Analyze / review (read-only, safe — the default) | `codex exec --sandbox read-only "<tight prompt>" 2>/dev/null` |
| Let Codex EDIT files | `codex exec --sandbox workspace-write "<prompt>" 2>/dev/null` |
| Independent review of the working tree | `codex review` |
| Apply Codex's last proposed diff | `codex apply` |
| Cheaper/deeper | add `-c model_reasoning_effort=low\|medium\|high\|xhigh` |
| Capture output | append `-o out.txt`, or `--json` for structured events |
| Continue a session | `codex exec resume --last "<follow-up>"` |

## Discipline (non-negotiable)
- **Scope tightly + hand it concrete material** — point Codex at exact `file:line`, the
  locked contract (`tasks/v11-locked-contract.md`), or the spec. It works best on
  concrete input, not vague asks.
- **`2>/dev/null`** to suppress thinking tokens unless you're debugging the run.
- **Verify every workspace-write run with the repo's own gates** — `go -C cli build ./...`,
  `go -C cli test ./...`, `c3local check`. Codex's edits are a *proposal*, never ground
  truth; treat its diff like an untrusted PR.
- **For high-stakes work, run it twice across models** — have Codex and a Claude subagent
  do the same task independently, then reconcile the diff. Cross-model disagreement is
  signal.
