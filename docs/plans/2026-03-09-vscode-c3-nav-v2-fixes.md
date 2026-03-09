# Fix: vscode-c3-nav branch base, CI, and install mechanism

**Date:** 2026-03-09

## Changes

### 1. Rebase onto dev
Branch `feat/vscode-c3-nav-v2` currently derives from `main`. Rebase onto `origin/dev` so it flows through the existing CI pipeline (push to dev -> build -> merge to main -> release).

### 2. Replace install script with README
- Delete `scripts/install-vscode-ext.sh`
- Create `vscode-c3-nav/README.md` with manual install instructions (download .vsix from releases, or build from source)

### 3. Update CI to build and release .vsix
In `distribute.yml`:
- **distribute job**: Add Node.js setup, `npm ci`, `npm run package` to build `c3-nav.vsix`. Copy it to main alongside Go binaries.
- **release job**: Include `c3-nav.vsix` as a release asset.

### 4. Extension versioning
Stays independent — version managed in `vscode-c3-nav/package.json` (currently `0.1.0`), not tied to main `VERSION` file.
