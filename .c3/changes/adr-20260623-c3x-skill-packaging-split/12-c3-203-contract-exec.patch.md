---
target: c3-203
scope: block
base: c3-203#n621@v1:sha256:e90b8f5f12df1dbf3a2dbaeda7c58b43268805a94bb2884e2917a2ee613e6dbc
---
| runtime exec | OUT | The selected runtime runs with C3X_VERSION exported and the original argv forwarded; unsupported platform or a missing binary/source/npm path exits non-zero with a recovery hint | Forwards argv unchanged; adds no architecture behavior of its own | skills/c3/bin/c3x.sh exec "$bin" "$@"; npm exec --yes --package "@c3x/cli@${VERSION}" -- c3x "$@"; unsupported-platform and not-found error exits |
