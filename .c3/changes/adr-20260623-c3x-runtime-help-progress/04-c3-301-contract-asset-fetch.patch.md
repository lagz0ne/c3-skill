---
target: c3-301
scope: block
base: c3-301#n519@v1:sha256:9fb5abf2f5227820426435edb792c3dd9f964187f9757922767ae3b9883cc7a2
---
| asset fetch + cache | OUT | Each asset download reports bounded stderr progress, matches its sha256 against the published checksum, then atomically renames into the per-version cache; a mismatch aborts | Help/version and local runtime metadata operations do not fetch assets; a checksum-failing or partially written asset never becomes the live binary | packages/cli/src/manager.ts ensureCachedAsset / createProgressReporter; checksum-mismatch throw |
