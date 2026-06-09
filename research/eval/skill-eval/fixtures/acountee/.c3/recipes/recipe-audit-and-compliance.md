---
id: recipe-audit-and-compliance
c3-seal: 149cef76c685b0ea7227ddd3a68b162444edef26b8b85117602e23ea547378be
title: Audit and Compliance
type: recipe
goal: Trace how mutations are captured for compliance — which tables use triggers vs explicit audit, and where security events fit.
sources:
    - c3-0
    - c3-208
    - ref-audit-trail
    - ref-rbac
---

# Audit and Compliance

## Goal

Trace how mutations are captured for compliance — which tables use triggers vs explicit audit, and where security events fit.

## Narrative

Every business mutation must be auditable (c3-0 constraint). The system
uses a hybrid audit strategy (ref-audit-trail):

**Explicit audit**: Flow code calls `createAuditEntry` for tables without
triggers — `teams`, `roles`, `user_roles`, `approval_flows`. Captures
semantic actions (e.g., `assign_role`) with before/after snapshots.

**DB trigger audit**: `log_change()` trigger on `invoices`, `pr`,
`invoice_services`. Fires on INSERT/UPDATE/DELETE. Reads actor from
PostgreSQL session variable: `set_config('app.current_user', base64email, true)`,
set by `executeInDrizzleTransaction`.

**Critical rule**: If a table has `log_change()` trigger, do NOT also
call `createAuditEntry` — this creates duplicates.

Each audit entry has: action, table_name, record_id, before/after JSONB,
triggered_at, triggered_by (email or 'system'), user_agent, ip_address,
md5 checksum, and optional metadata.

Security events (ref-rbac) are a separate audit surface for RBAC mutations
(role assignment, revocation, permission changes). These go to
`security_events` table, not the general audit log.

Audit flows (c3-208) expose the audit data to the UI — timeline views
with before/after diffs for compliance review.
