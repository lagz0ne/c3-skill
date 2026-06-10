## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "invoice ZIP import partial failure duplicate transactional file storage"
c3 read c3-104 --full
c3 read c3-206 --full
c3 read ref-file-handling --full
c3 read c3-204 --full
c3 read ref-sync --full
c3 read adr-20260212-workbench-feature --full
c3 graph ref-file-handling --depth 2 --direction reverse
c3 graph c3-206 --depth 1
```

## Answer

Emergent property: import **idempotency** with a partial-success boundary. Duplicates and failures are explicit result states, not hidden exceptions that make the whole ZIP result ambiguous.

Trace:

| Segment | Entities | Mechanism |
| --- | --- | --- |
| Import action | `c3-104`, `c3-206` | InvoiceScreen starts import; `importFiles` handles XML/ZIP parsing, MD5 duplicate checks, inserts, and sync. |
| Storage/database boundary | `ref-file-handling`, `c3-204` | PostgreSQL BYTEA file storage, content-hash deduplication, and schema ownership for `files`/`invoices`. |
| Result/sync context | `ref-sync`, `adr-20260212-workbench-feature` | Successful mutations emit sync; per-item bulk results match the existing importFiles pattern. |

For a ZIP with duplicates and parse failures, `c3-206` can produce partial state: successful XML entries insert, duplicate MD5 entries become skipped, and parse/insert errors become failure. `ref-file-handling` governs because it defines PostgreSQL storage plus the `success | failure | skipped` tri-state. `ref-sync` governs because successful invoice mutations need full-record deltas to clients.

## Grounding

`ref-file-handling` governs file handling and deduplication. `ref-sync` governs how successful import mutations become visible to connected clients.

## Caveats

The fixture has known `c3 check` drift. No `rule-*` entities found in the fixture.
