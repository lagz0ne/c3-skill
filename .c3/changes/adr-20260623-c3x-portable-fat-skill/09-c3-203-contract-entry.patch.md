---
target: c3-203
scope: block
base: c3-203#n620@v1:sha256:d7edf5d047fec416c5b5cb1e1839482c5106168dca444a7db45ebb9a547d97be
---
| c3x.sh entrypoint | IN | Invoked with C3 args; resolves VERSION + platform to a full asset name and, on Linux, a portable asset name; execs the full bundled binary when present, then the portable bundled binary when present, source-builds the full binary when running from a checkout with Go installed, otherwise delegates to the pinned npm runtime manager | No-binary npm delegation happens only after local bundled/source paths are absent; normal C3 command argv is forwarded unchanged | skills/c3/bin/c3x.sh asset_name / portable_bin / bundled exec / go build / npm exec branches |
