---
target: c3-203
scope: insert
base: c3-203#n616@v1:sha256:0ab994bab6c71d3f692691e55c8cec5bdb1f3e89dd3ff6e0e358c326f58a31cd
---
| ref-fat-thin-distribution | ref | The wrapper behavior required by fat skill ZIPs and no-binary skill/plugin artifacts | Fat artifacts exec bundled binaries; no-binary artifacts delegate to the pinned npm runtime manager | c3x.sh tries bundled binary, source build, then npm exec @c3x/cli@VERSION. |
