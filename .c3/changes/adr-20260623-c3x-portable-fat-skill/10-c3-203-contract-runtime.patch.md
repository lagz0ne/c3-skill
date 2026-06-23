---
target: c3-203
scope: block
base: c3-203#n621@v1:sha256:5590624df15d6f7d2529b10163a1fca08ea7fd67e002b84fd24a254f3e5dec85
---
| runtime exec | OUT | The selected runtime runs with C3X_VERSION exported and the original argv forwarded; unsupported platform or a missing binary/source/npm path exits non-zero with a recovery hint that names npm, fat/portable skill artifacts, and source builds | Forwards argv unchanged; adds no architecture behavior of its own | skills/c3/bin/c3x.sh exec "$bin" "$@"; exec "$portable_bin" "$@"; npm exec --yes --package "@c3x/cli@${VERSION}" -- c3x "$@"; unsupported-platform and not-found error exits |
