# Test Codebases

Add real projects here using git subtree for authentic evaluation.

## Adding a Codebase

```bash
# From repo root
git subtree add --prefix evals/codebases/<name> <repo-url> <branch> --squash

# Example
git subtree add --prefix evals/codebases/sample-api https://github.com/you/sample-api main --squash
```

## Updating a Codebase

```bash
git subtree pull --prefix evals/codebases/<name> <repo-url> <branch> --squash
```

## Requirements

Good test codebases should:
- Have clear structure (multiple modules/services)
- Not already have .c3/ (for ADOPT testing)
- OR have .c3/ (for ADJUST testing)
- Be representative of real-world projects

## Suggested Codebases

- Small API project (3-5 containers)
- Monorepo with multiple services
- Frontend + backend project
- CLI tool with plugins
