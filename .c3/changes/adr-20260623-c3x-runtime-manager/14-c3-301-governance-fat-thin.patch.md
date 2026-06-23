---
target: c3-301
scope: block
base: c3-301#n514@v1:sha256:489c8e80276c9f381e7976e27c8e8ab75c8ba1efc2b43acd294e1533661f8dd5
---
| ref-fat-thin-distribution | ref | Why the npm package ships only a runtime manager and pulls verified release assets on demand instead of bundling binaries | Thin client fetches; the binary stays the single source | manager.ts downloads from the GitHub Release base URL for the selected version, the thin-client half of the split. |
