---
target: c3-301
scope: block
base: c3-301#n518@v1:sha256:44b0b6837d66ad2153d1d86448fbc9cf8bfb77c8a8643d32529e9444e422dc35
---
| c3x CLI entry | IN | Invoked with argv; handles npm-only `runtime` subcommands locally, otherwise prepares the selected runtime and execs the cached binary, propagating its exit status | Consumes only the `runtime` namespace; normal C3 command argv is forwarded unchanged | packages/cli/src/cli.ts; manager.ts runManagerCommand / runCli / defaultExec |
