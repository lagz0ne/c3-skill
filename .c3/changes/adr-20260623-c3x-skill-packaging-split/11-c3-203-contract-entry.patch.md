---
target: c3-203
scope: block
base: c3-203#n620@v1:sha256:7c5f74610b6c8813a5a2de28b4d0787336f28042a5ecb4209f18909fa458d844
---
| c3x.sh entrypoint | IN | Invoked with C3 args; resolves VERSION + platform to one asset name, execs the matching bundled binary when present, source-builds it when running from a checkout with Go installed, otherwise delegates to the pinned npm runtime manager | No-binary npm delegation happens only after local binary/source paths are absent; normal C3 command argv is forwarded unchanged | skills/c3/bin/c3x.sh asset_name / bundled exec / go build / npm exec branches |
