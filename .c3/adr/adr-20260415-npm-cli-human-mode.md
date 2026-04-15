---
id: adr-20260415-npm-cli-human-mode
c3-seal: 8e312655fd8051272272b0b1d777c94d12075fd395655dd0ccebd09a09bc478d
title: npm-cli-human-mode
type: adr
goal: Document the npm shim as a C3 component, map its package files, and make the shim clear inherited `C3X_MODE` before delegating to `c3x.sh` so npm callers keep human/default output unless they pass explicit output flags. The skill wrapper remains responsible for agent mode. The npm package can publish as a wrapper-only patch while the plugin release is promoted to 9.1.0 for the broader strict component-doc impact.
status: proposed
date: "2026-04-15"
---

## Goal

Document the npm shim as a C3 component, map its package files, and make the shim clear inherited `C3X_MODE` before delegating to `c3x.sh` so npm callers keep human/default output unless they pass explicit output flags. The skill wrapper remains responsible for agent mode. The npm package can publish as a wrapper-only patch while the plugin release is promoted to 9.1.0 for the broader strict component-doc impact.
