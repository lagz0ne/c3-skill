---
target: c3-301
scope: block
base: c3-301#n712@v1:sha256:042acac0c2d862a4cb9c3f3f3cab9d8d5ecd8d5374f2a49670ac093b2af9ea3d
---
| project runtime metadata | IN | `.c3/runtime.json` may select only a version or `latest` | Rejects paths, URLs, timestamps, commands, checksums, release hosts, and unknown fields before execution | manager.ts readProjectRuntimeConfig / writeProjectRuntimeConfig; package tests |
