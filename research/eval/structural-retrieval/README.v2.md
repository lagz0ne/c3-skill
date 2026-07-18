# Real C3 structural-retrieval adapter v2

This development adapter exercises the exported `cmd.RunSearch` boundary on a fresh synthetic C3 store. It does not rank fixture documents itself.

```text
controller: fixture + oracle + authority
              |
              | redacted corpus + raw query
              v
confined arm: Store + content + relationships -> cmd.RunSearch -> raw rows
              |
              v
controller: direct probes + logical DB dump + score + provenance + history
```

The arm response contains only case ids, queries, and ordered search rows. The controller owns the database path, direct FTS probes, canonical logical dump, oracle, metrics, provenance, report, and history.

External execution confinement requires all three layers:

- a user cgroup created by `systemd-run` for `TasksMax`, `MemoryMax`, and runtime;
- `prlimit` for CPU time and open files;
- `bwrap` for filesystem, environment, PID/session, and network isolation.

Missing capability invalidates an arm. An in-process unit replay proves the real C3 loader and scoring seam, but is not labeled externally confined.

The built controller also launches the built tool's `--arm` mode through that full stack. Before spawn, a strict controller-owned authority file must match the runtime, controller, fixture, scorer, source capsule, dirty patch, environment, module graph, action envelope, budget contract, and external threshold authority. The controller rebuilds the source capsule with `go build -trimpath` and requires the rebuilt bytes to match. This is a local rebuild proof; it is not a portable bundle-verification claim.

A live test proves the real binary cannot see controller-only oracle, report, history, repository, or network surfaces. Another live mutation strips the accepted binary and proves the changed bytes are rejected before work or output directories exist. Prefix logs, suffix bytes, second objects, unknown fields, and oversized output are rejected.

The controller writes separate reports for every isolated case, the combined development corpus, and the deterministic scale corpus. It never averages metrics across those modes. Owner recall is averaged only across wrong-layer queries; relationship-route recall and MRR are separate metrics. Expansion-route credit requires the frozen logical relationship witness, the expected top-five row and match source, and a controller-observed direct entity/content FTS miss. Route coverage uses each fixture's frozen `required_route_fields` only.

Before the first arm starts, the controller exclusive-creates an empty history and output snapshot. After each successful arm, its exact stdout bytes are exclusive-created under `results/` and hashed, its controller report is exclusive-created under `reports/` and hashed, its history row is appended, and a new output snapshot is exclusive-created. A later failure cannot erase earlier run evidence. The failure path retains any raw bytes and actual measurements available, writes controller-owned error evidence, appends a `crash` or `invalid` history row linked to the prior tail, writes another output snapshot, and points stdout and stderr to that durable evidence.

The append-only history binds result and report hashes, controller-scored metrics, actual wall/CPU/max-RSS/process/stdout/stderr/case/SQLite/logical-dump measurements, and the external threshold authority. Controller stdout contains only those refs, hashes, measured actuals, and failure evidence when present. Every run remains `diagnostic_unadmitted`.

Live controller tests capture stdout and stderr separately. Their output directory is created outside `t.TempDir`; passing tests remove it, while failing tests retain it and print its absolute path. Failure stderr names the absolute durable output snapshot, its SHA-256, and the failure history record, so the refs-only stdout can be replayed after Go removes its own test directories.

The wall envelope comes from the hash-bound `BudgetLimits.WallTimeMillis`, not a second command literal. One shared registered value freezes that authority field at 60,000 milliseconds, which becomes uniform `RuntimeMaxSec=60s` enforcement for every arm. A self-consistent authority with another wall is rejected. Confinement validation keeps CPU at 10 seconds and requires wall time to exceed CPU plus a five-second startup margin. For the frozen four fixtures, the controller derives six arm slots plus four minutes of controller overhead: `6 * 60s + 240s = 600s`. CPU, memory, tasks, network, runtime behavior, source inputs, fixtures, scorer, threshold, diagnostics, and `diagnostic_unadmitted` admission stay unchanged.

`canonical_row_bytes` is the UTF-8 length of compact JSON for the exact returned `SearchResultRow` slice. It is a context-size surrogate, not token or money cost. The 1.05 ratio is admissible only when a controller-supplied authority hash matches the external steering check-in recorded in `benchmark.v2.json`. Editing the benchmark cannot create that controller provenance.

The fixtures are generic development cases. They contain no product repository facts. Their results test this adapter and its retrieval measurements only. They do not prove answer correctness, product value, paper claims, or paired-evaluation advantage.
