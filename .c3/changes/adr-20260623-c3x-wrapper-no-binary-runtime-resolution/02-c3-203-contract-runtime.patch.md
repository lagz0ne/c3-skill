---
target: c3-203
scope: block
base: c3-203#n621@v1:sha256:15e3cfb61e104e29fd2539cb839702b05d81f1a506dd6754406bee5ee9bc4b86
---
| runtime exec | OUT | Bundled and source-built runtimes run with C3X_VERSION exported and the original argv forwarded; no-binary passive root help/version is answered locally; no-binary real commands run through the pinned npm manager package without wrapper-created C3X_VERSION; unsupported platform or a missing binary/source/npm path exits non-zero with a recovery hint that names npm, fat/portable skill artifacts, and source builds | Forwards normal C3 command argv unchanged; adds no architecture behavior of its own; keeps no-binary runtime-version resolution inside the npm manager | skills/c3/bin/c3x.sh export C3X_VERSION before exec "$bin" "$@" / exec "$portable_bin" "$@"; print_wrapper_help and version branches before npm fallback; npm exec --yes --package "@c3x/cli@${VERSION}" -- c3x "$@"; unsupported-platform and not-found error exits |
