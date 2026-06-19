# Working in this repo

This is the C3 **product source** — the Go CLI and the Claude skill that ships it. This file is the repo-dev contract: how to invoke C3 here, how we work, and how we release. It teaches no product behavior.

**What C3 is, the operation set, and the change-unit / freeze / canvas model are owned by `skills/c3/SKILL.md`.** Read it there; do not duplicate it here.

---

## Local C3 Source Rule

This repository is the C3 project source. Work here is intentionally outside the installed/global C3 skill scope — never let the global skill answer for this checkout.

Hard rules:
- Do not use bare `c3x`; it may resolve to the global installed skill.
- Do not load C3 from `~/.agents/skills/c3`, `~/.claude/skills/c3`, `~/.codex/skills/c3`, or marketplace installs for this repo.
- Load C3 skill instructions from `skills/c3/SKILL.md` in this checkout.
- Run C3 through the local built wrapper: `C3X_MODE=agent bash skills/c3/bin/c3x.sh <command>`.
- If `skills/c3/bin/c3x.sh` or the matching local binary is missing, tell the user — CI builds binaries, not you. Only run `bash scripts/build.sh` when explicitly debugging the build.
- The wrapper rebuilds only when the platform binary is **absent**; after editing CLI source, delete the stale gitignored binary in `skills/c3/bin/` so you don't dogfood old code.
- At session start, create a local alias/function and use it for every C3 command:

```bash
alias c3local='C3X_MODE=agent bash skills/c3/bin/c3x.sh'
c3local check
```

If C3 output looks wrong, commands fail unexpectedly, or behavior differs from source changes, suspect the wrong C3 version is being used. Prove the path/version before continuing.

File lookup: `c3local lookup <file-or-glob>` maps files/directories to components + refs.
CLI: `c3local <command>` after the alias above; see `skills/c3/SKILL.md` for the operation set.

---

## Workflow

- Load `/superpowers:brainstorming` and `/superpowers-developing-for-claude-code:developing-claude-code-plugins` up front.
- Use `AskUserQuestionTool` where possible — it yields a better-grounded answer.
- Start with brainstorming to pin the intention; align the concept in prose before offering implementation menus.
- Once understood, write the plan, then implement in parallel using subagents.
- Before claiming work is done: run `/noslop` to strip AI slop, then run the local C3 flow (`c3local check` + the audit guidance in `skills/c3/SKILL.md`) to verify docs match code.
- Delegate to `/release` when done; confirm scope with the user. Patch by default.

---

## Plugin Structure

### plugin.json

Auto-discovery only. Do NOT add explicit component paths.

```json
{
  "name": "c3-skill",
  "version": "x.x.x"
}
```

### Repository Layout

```
c3-design/
├── .claude-plugin/           # Plugin metadata
│   ├── plugin.json
│   └── marketplace.json
├── cli/                      # Go CLI source
│   ├── main.go
│   ├── cmd/                  # Command implementations
│   └── internal/             # Core libraries (schema, store, changeset,
│       │                     #   codemap, frontmatter, index, walker, …)
│       └── templates/        # Embedded doc templates
├── packages/cli/             # npm @c3x/cli thin client (downloads the binary)
├── skills/c3/                # Unified skill (auto-discovered)
│   ├── SKILL.md              # Skill definition + intent router
│   ├── bin/                           # CLI wrapper + version (binaries built in CI)
│   │   ├── c3x.sh                    # Platform-detecting wrapper (committed)
│   │   ├── VERSION                   # Current version, read by c3x.sh (committed)
│   │   └── c3x-{version}-{os}-{arch} # Cross-compiled binaries (gitignored; local
│   │                                 #   builds accumulate here, only the matching
│   │                                 #   platform/version is used)
│   └── references/           # Operation-specific guidance (8 files)
│       ├── onboard.md
│       ├── query.md
│       ├── audit.md
│       ├── change.md
│       ├── canvas.md
│       ├── ref.md
│       ├── rule.md
│       └── sweep.md
└── scripts/
    └── build.sh              # Cross-compile Go CLI (debug-only; CI owns the build)
```

### Build System

**Do NOT run `bash scripts/build.sh` during releases.** CI owns the build — pushing a `v*` tag triggers `distribute.yml`, which tests, cross-compiles the supported platforms, and creates the GitHub Release automatically. Only run `build.sh` locally when debugging a build issue.

```bash
cd cli && go test ./...       # Run Go tests locally
```

### CI/CD

- **Push a `v{VERSION}` tag** -> `distribute.yml` (Build & Distribute): tests, cross-compiles `linux/amd64`, `linux/arm64`, `darwin/arm64` (VERSION is derived from the tag name), assembles the plugin zip + binaries, and cuts the GitHub Release. This is the release trigger.
- **Push to `main`** (a merge/`distribute dev` commit) -> `release.yml` checks `skills/c3/bin/VERSION` (fallback path).
- The npm `@c3x/cli` publish (`Publish @c3x/cli` workflow) needs an `NPM_TOKEN` secret — it currently fails `ENEEDAUTH`, independent of the GitHub Release.

### Release Process

1. Commit changes to `dev` (merge the work branch onto `dev`).
2. Add a `CHANGELOG.md` entry for the version.
3. Bump the version everywhere it appears: `skills/c3/bin/VERSION`, `.claude-plugin/plugin.json`, `.claude-plugin/marketplace.json`, and the npm client `packages/cli/{package.json, package-lock.json, src/version.ts}`.
4. Push `dev`, then `git tag -a v{VERSION} && git push origin v{VERSION}` — the tag triggers the build + GitHub Release.
5. Verify with `gh run watch` and `gh release view v{VERSION}`.

### Versioning

All version files must stay in sync:

| File | Purpose |
|------|---------|
| `skills/c3/bin/VERSION` | Source of truth — CI, c3x.sh, and build.sh all read this |
| `.claude-plugin/plugin.json` | Plugin metadata |
| `.claude-plugin/marketplace.json` | Marketplace listing |
| `packages/cli/package.json` | npm `@c3x/cli` thin-client version |
| `packages/cli/package-lock.json` | npm lockfile (two `version` fields) |
| `packages/cli/src/version.ts` | `C3X_VERSION` the npm wrapper pins + downloads |

Use the `/release` command to bump versions consistently. The `v{VERSION}` git tag must match `skills/c3/bin/VERSION` (CI derives the build version from the tag name).
