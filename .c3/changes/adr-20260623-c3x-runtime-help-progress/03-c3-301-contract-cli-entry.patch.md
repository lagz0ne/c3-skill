---
target: c3-301
scope: block
base: c3-301#n518@v1:sha256:3fe363a7645e311e1b41b889e42a73033ffa2d925b1befe13bc725b9616fef02
---
| c3x CLI entry | IN | Invoked with argv; handles root help, empty argv, root version, and npm-only `runtime` subcommands locally, otherwise prepares the selected runtime with automatic download progress and execs the cached binary, propagating its exit status | Consumes only root discovery commands and the `runtime` namespace; normal C3 command argv is forwarded unchanged | packages/cli/src/cli.ts; manager.ts runManagerCommand / runCli / defaultExec |
