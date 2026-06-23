---
target: c3-301
scope: insert
base: c3-301#n519@v1:sha256:b864d4b312482965caf6eb3982731efe8143f700e211d130e65d671ea235ac02
---
| project runtime metadata | IN | `.c3/runtime.json` may select only a version or `latest` plus update timestamp | Rejects paths, URLs, commands, checksums, release hosts, and unknown fields before execution | manager.ts readProjectRuntimeConfig / writeProjectRuntimeConfig; package tests |
