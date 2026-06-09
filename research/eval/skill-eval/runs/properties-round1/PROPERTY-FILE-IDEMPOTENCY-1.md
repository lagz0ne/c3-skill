## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "invoice ZIP import partial failure duplicate transactional file storage"
c3 read c3-104 --full
c3 read c3-206 --full
c3 read ref-file-handling --full
c3 read ref-sync --full
c3 graph ref-file-handling --depth 2 --direction reverse
c3 graph c3-206 --depth 1
```

## Answer

Invoice import starts in `c3-104` and reaches `c3-206` `importFiles`. The flow stores uploads, detects XML or ZIP, parses invoice XML, checks duplicate raw XML content by MD5, inserts invoice data, and emits sync.

`ref-file-handling` is the core file mechanism. It stores files in PostgreSQL, uses content hashes for duplicate detection, and returns explicit result states: success, failure, or skipped. `c3-206` says ZIP imports can have partial state when some entries succeed and others fail.

So duplicate entries should become skipped, parse or insert errors should become failure, and successful entries should produce sync through `ref-sync`. This keeps the import result understandable without treating expected duplicates as exceptions.

## Grounding

`ref-file-handling` governs file storage and result handling. `ref-sync` governs successful mutation propagation to clients.

## Caveats

The fixture has known `c3 check` drift. I did not inspect source code beyond C3 docs.
