---
target: c3-203
scope: block
base: c3-203#n1139@v1:sha256:73742c797989859803234a3ff522fdc623f3e499bab09d59517636c8584141c6
---
| ref-fat-thin-distribution | ref | The wrapper behavior required by full-fat skill ZIPs, Linux portable fat skill ZIPs, and no-binary skill/plugin artifacts | Full/portable fat artifacts exec bundled binaries; no-binary artifacts delegate to the pinned npm runtime manager | c3x.sh tries full bundled binary, Linux portable bundled binary, source build, then npm exec @c3x/cli@VERSION. |
