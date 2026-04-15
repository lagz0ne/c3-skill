# C3 Design Agent Instructions

## Local C3 Source Rule

This repository is the C3 project source. All C3 operations here must use this checkout, not an installed/global C3 skill.

Hard rules:
- Do not use bare `c3x`; it may resolve to a global installed skill.
- Do not load C3 from `~/.agents/skills/c3`, `~/.claude/skills/c3`, `~/.codex/skills/c3`, or marketplace installs for this repo.
- Load C3 skill instructions from `skills/c3/SKILL.md` in this repository.
- Run C3 through the local built wrapper:

```bash
C3X_MODE=agent bash skills/c3/bin/c3x.sh <command>
```

If the local wrapper or binary is missing, build it first:

```bash
bash scripts/build.sh
```

Then create a session-local alias/function and use it for every C3 command:

```bash
alias c3local='C3X_MODE=agent bash skills/c3/bin/c3x.sh'
c3local verify
```

If command behavior, output format, or verification result seems inconsistent with this source tree, be doubtful. First prove which binary and skill path are being used, then continue.

## Workflow

- Use RED-GREEN-TDD for changes.
- Keep `.c3/c3.db` treated as disposable local cache, not submitted truth.
- Verify with `c3local verify` and the relevant project tests before done.
- For code changes in `cli/`, run `go test ./...` from `cli/`.

## CLI Flow Principles

- CLI commands should lead within their capabilities: inspect what they can prove, return the smallest useful answer, and include the next repair step when there is a failure.
- In `C3X_MODE=agent`, structured command output must be TOON, not JSON. Do not add command-local `json.NewEncoder` paths for agent-facing command results; route through the shared output helpers.
- Failure output is part of the workflow contract. Prefer grouped blockers and direct fix loops over one-error-at-a-time guessing, but keep operations tight enough that the command cannot leave canonical `.c3/` files in an impossible partial state.
