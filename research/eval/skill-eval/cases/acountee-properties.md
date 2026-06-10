# Acountee diverse-property cases

Ground truth was established with the local C3 wrapper:

```bash
C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 <cmd>
```

The fixture cache must be rebuilt with fixture `check` before search/read. The
fixture still has known canonical drift; do not repair it for this eval.

These cases intentionally avoid the notification-heavy cross-cut property. A
complete answer must trace real C3 ids and name the specific emergent property.

## PROPERTY-AUDIT-ATOMICITY-1: mass approval audit consistency

Question: If mass approval partially fails, is the audit trail still consistent?

Grounding commands:

```bash
c3 search "mass approval partial failure audit trail consistency"
c3 read c3-105 --full
c3 read c3-205 --full
c3 read recipe-approval-workflow --full
c3 read ref-audit-trail --full
c3 read recipe-audit-and-compliance --full
c3 read c3-208 --full
c3 read c3-202 --full
c3 graph ref-audit-trail --depth 2 --direction reverse
```

Expected trace:

| Segment | Expected ids and facts |
| --- | --- |
| User action | `c3-105` owns PaymentRequestsScreen Approvals mode and bulk approval entry. |
| Bulk mutation | `c3-205` owns `approveAll`; it iterates `pr_ids` and collects `approved` and `failed` arrays. |
| Approval semantics | `ref-approval-chain` and `recipe-approval-workflow` govern PR lifecycle and step approval semantics. |
| Audit mechanism | `recipe-approval-workflow` says PR mutations are audit-captured by DB trigger on the `pr` table and must not double-write explicit audit entries. |
| Audit contract | `ref-audit-trail` says explicit audit writes participate in the active transaction and trigger-covered tables use `log_change()` so no silent data change escapes capture. |
| Audit query surface | `c3-208` exposes audit history/list/export/stats for compliance review. |
| Execution context | `c3-202` supplies `transactionTag`; `ref-audit-trail` says transaction setup sets `app.current_user` for triggers. |
| Emergent property | **Audit atomicity/consistency under partial failure**: committed PR mutations get audit entries; failed/rolled-back PR attempts should not create orphan audit records. Consistency is per persisted mutation, not a claim that the whole bulk batch is all-or-nothing. |

Evidence snippets:

```text
search: recipe-approval-workflow ... Approval mutations are audit-captured via DB trigger on `pr` table
read c3-105: Approvals mode ... Approve, reject, bulk approve
read c3-205: approveAll | Bulk approval: iterates pr_ids, approves each, collects approved/failed arrays
read recipe-approval-workflow: do NOT also call createAuditEntry (ref-audit-trail)
read ref-audit-trail: audit writes are atomic with the mutation; if mutation rolls back, audit entry should too
read ref-audit-trail: DB triggers guarantee no silent data change escapes capture
read recipe-audit-and-compliance: DB trigger audit on invoices, pr, invoice_services
read c3-202: transactionTag | DrizzleTransaction | Active DB transaction
graph ref-audit-trail reverse: cited by c3-208, recipe-approval-workflow, recipe-audit-and-compliance
```

## PROPERTY-CONFIG-BLAST-RADIUS-1: NATS subject prefix blast radius

Question: What is affected if I change `NATS_SUBJECT_PREFIX` away from `sync`?

Grounding commands:

```bash
c3 search "configuration scope mechanism blast radius sync subject prefix"
c3 read ref-scope-controlled-config --full
c3 graph ref-scope-controlled-config --depth 2 --direction reverse
c3 read ref-sync --full
c3 graph ref-sync --depth 2 --direction reverse
c3 read c3-202 --full
c3 read c3-203 --full
c3 read c3-209 --full
c3 read c3-101 --full
c3 read c3-4 --full
c3 read c3-211 --full
c3 read recipe-realtime-sync --full
```

Expected trace:

| Segment | Expected ids and facts |
| --- | --- |
| Config scope owner | `ref-scope-controlled-config` governs config access through runtime scope tags. Reverse graph shows `c3-202`, `c3-203`, `c3-204`, `c3-209`, `c3-211`, and `recipe-backend-foundations` depend on it. |
| Subject contract | `ref-sync` defines `{prefix}.broadcast` and `{prefix}.user.{escaped_email}` and says the current default prefix is `sync` / `NATS_SUBJECT_PREFIX`. |
| Frontend lockstep | `c3-101` currently subscribes to `sync.broadcast` and `sync.user.<email>` directly; if the prefix changes, frontend subscription wiring must change in lockstep. |
| Credential permissions | `c3-209` uses `natsConfig.subjectPrefix` to grant browser subscribe permissions for `{prefix}.broadcast` and `{prefix}.user.{escaped_email}`. |
| Broker permissions | `c3-4` NATS config permits `{prefix}.>` with default prefix `sync`; broker permissions must match the new prefix. |
| Notification path | `c3-211` also uses scope-controlled config and publishes in-app notifications on user subjects. |
| Sync dependents | Reverse graph for `ref-sync` includes `c3-101`, `c3-205`, `c3-206`, `c3-207`, `c3-210`, `c3-212`, `recipe-approval-workflow`, and `recipe-realtime-sync`. |
| Emergent property | **Blast radius/scope of impact**: one config value crosses backend scope tags, JWT permission generation, NATS broker permissions, frontend subscriptions, sync flows, and user notification subjects. |

Evidence snippets:

```text
search: c3-101 ... current frontend atom subscribes to `sync.*` directly
search: c3-209 ... natsConfig.subjectPrefix | sync
read ref-scope-controlled-config: all config access through scope tags; define tags for config: natsConfig.subjectPrefix
graph ref-scope-controlled-config reverse: c3-202, c3-203, c3-204, c3-209, c3-211, recipe-backend-foundations
read ref-sync: current default prefix is `sync` (`NATS_SUBJECT_PREFIX`)
read ref-sync: if prefix changes from `sync`, frontend subscription wiring must change in lockstep
graph ref-sync reverse: c3-101, c3-205, c3-206, c3-207, c3-210, c3-212, recipe-approval-workflow, recipe-realtime-sync
read c3-4: pub/sub permissions include `{prefix}.>` with default prefix: sync
```

## PROPERTY-TRANSPORT-SYNC-COUPLING-1: NATS JWT auth to sync coupling

Question: How does changing NATS JWT auth affect the sync layer?

Grounding commands:

```bash
c3 search "NATS JWT auth sync layer coupling WebSocket"
c3 read ref-nats-jwt-auth --full
c3 graph ref-nats-jwt-auth --depth 2 --direction reverse
c3 read c3-209 --full
c3 read c3-4 --full
c3 read c3-101 --full
c3 read ref-sync --full
c3 read recipe-realtime-sync --full
c3 read recipe-auth-and-access --full
c3 read adr-20260113-nats-jwt-resolver --full
c3 read adr-20260112-nats-auth-callout --full
c3 graph c3-209 --depth 1
c3 graph c3-101 --depth 1
```

Expected trace:

| Segment | Expected ids and facts |
| --- | --- |
| Auth contract | `ref-nats-jwt-auth` selects JWT resolver with memory preload, no auth callout service, and credentials generated at page load. |
| Credential owner | `c3-209` generates per-session JWT + nkey and returns `{ jwt, seed }` / loader credentials to the browser. |
| Broker enforcement | `c3-4` validates signatures, expiration, and permissions, then accepts browser WSS connections. |
| Client sync receiver | `c3-101` uses JWT auth to connect via WebSocket and subscribes to `sync.broadcast` and `sync.user.<email>`. |
| Sync contract | `ref-sync` and `recipe-realtime-sync` define delta/ack delivery, executionId correlation, and subject naming. |
| Historical decisions | `adr-20260113-nats-jwt-resolver` supersedes the auth-callout direction in `adr-20260112-nats-auth-callout`. |
| Emergent property | **Transport-auth/sync coupling**: sync delivery is gated by transport identity. Changing JWT signing, resolver config, expiry, or permissions can break sync even if HTTP application auth still works. |

Evidence snippets:

```text
search: ref-nats-jwt-auth ... NATS validates directly using a configured account public key - no auth callout service needed
read ref-nats-jwt-auth: browser connects with JWT credentials; NATS verifies signature, expiration, and embedded permissions
graph ref-nats-jwt-auth reverse: cited by adr-20260113-nats-jwt-resolver, c3-209, recipe-auth-and-access
read c3-209: creates ephemeral user keypair; signs JWT; loader passes credentials to client
read c3-4: WSS 8080 browser connection; JWT resolver; permission enforcement
read c3-101: natsSync connects via WebSocket with JWT auth
read ref-sync: `{prefix}.broadcast` carries deltas + acks; executionId correlation drives client wait/notify
read recipe-realtime-sync: executionId is threaded HTTP response -> execution context tag -> sync.emit/ack -> client tracker
```

## PROPERTY-FILE-IDEMPOTENCY-1: invoice import partiality and idempotency

Question: If an invoice ZIP import has duplicates and parse failures, what keeps import/file state coherent?

Grounding commands:

```bash
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

Expected trace:

| Segment | Expected ids and facts |
| --- | --- |
| User action | `c3-104` owns InvoiceScreen import via drag-and-drop dialog and server `importFiles`. |
| Import owner | `c3-206` owns `importFiles`, stores files, parses XML/ZIP, deduplicates by raw XML MD5 hash, inserts invoices/services, and emits sync. |
| File contract | `ref-file-handling` stores files in PostgreSQL BYTEA, uses content-hash deduplication, and returns `success | failure | skipped`. |
| Database boundary | `c3-204` owns schema and lists `files`, `invoices`, and related tables under Drizzle/PostgreSQL. |
| Sync update | `ref-sync` applies full-record deltas after successful mutations. |
| Bulk result precedent | `adr-20260212-workbench-feature` says bulk operations with per-item results match the existing `importFiles` pattern. |
| Emergent property | **Import idempotency/partial-success boundary**: duplicates become `skipped`, parse/insert errors become `failure`, successful XML entries commit and sync, and PostgreSQL file storage keeps file data transactional with invoice data. |

Evidence snippets:

```text
search: c3-206 ... success, skipped, failure; ZIP imports can produce partial state
read c3-104: Import invoices via drag-and-drop dialog; parsed server-side via importFiles
read c3-206: importFiles processes XML/ZIP, deduplicates, inserts
read c3-206: deduplication uses MD5 of raw XML content against countInvoiceByHashValue
read ref-file-handling: PostgreSQL storage keeps files transactional with the rest of the data
read ref-file-handling: result type uses success | failure | skipped tri-state
read c3-204: key tables include invoices and files
read adr-20260212-workbench-feature: bulk operations with per-item results matches existing importFiles pattern
graph ref-file-handling reverse: cited by c3-206
```
