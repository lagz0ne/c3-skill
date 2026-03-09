# Fix: vscode-c3-nav branch base, CI, and install mechanism

**Date:** 2026-03-09

## Changes

### 1. Rebase onto dev
Branch `feat/vscode-c3-nav-v2` currently derives from `main`. Rebase onto `origin/dev` so it flows through the existing CI pipeline (push to dev -> build -> merge to main -> release).

### 2. Replace install script with README
- Delete `scripts/install-vscode-ext.sh`
- Create `vscode-c3-nav/README.md` with manual install instructions (download .vsix from releases, or build from source)

### 3. Dedicated CI for extension
- New workflow: `.github/workflows/vscode-extension.yml`
- Triggers on push to `dev` only when `vscode-c3-nav/**` files change
- Reads version from `vscode-c3-nav/package.json`, creates `c3-nav-v{version}` tag
- Releases `.vsix` as a standalone GitHub Release (separate from skill releases)
- `distribute.yml` unchanged — skill release does NOT include the extension

### 4. Extension versioning
Stays independent — version managed in `vscode-c3-nav/package.json` (currently `0.1.0`), not tied to main `VERSION` file.

### 5. README
Main `README.md` updated with VS Code extension section explaining usage and install-from-release instructions.
