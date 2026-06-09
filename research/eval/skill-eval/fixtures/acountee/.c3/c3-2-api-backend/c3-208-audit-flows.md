---
id: c3-208
c3-version: 3
c3-seal: 25ca50629da1d035653471ff48f6864c10a899a19facb8b5ccf3e84bb084e9c2
title: Audit Flows
type: component
category: feature
parent: c3-2
goal: Audit trail querying - history lookup, paginated list, export, statistics
uses:
    - ref-audit-trail
    - ref-pumped-fn
    - ref-query-services
    - ref-server-functions
    - ref-structured-logging
---

# Audit Flows

## Goal

Audit trail querying - history lookup, paginated list, export, statistics

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-2 |
| Role | Own Audit Flows behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep Audit Flows decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for Audit Flows so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before Audit Flows behavior is changed. | ref-audit-trail |
| Inputs | Accept only the files, commands, data, or calls that belong to Audit Flows ownership. | ref-audit-trail |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-audit-trail |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-audit-trail |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks Audit Flows to deliver its documented responsibility. | ref-audit-trail |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-audit-trail |
| Alternate paths | When a request falls outside Audit Flows ownership, hand it to the parent or sibling component. | ref-audit-trail |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-audit-trail |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-audit-trail | ref | Governs Audit Flows behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| Audit Flows input | IN | Callers must provide context that matches the component goal and parent fit. | c3-2 boundary | c3x lookup plus targeted tests or review. |
| Audit Flows output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-2 boundary | c3x check and project test suite. |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Contract drift | Goal, boundary, or derived material changes without matching component docs. | Compare Goal, Parent Fit, Contract, and Derived Materials. | Run c3x check and relevant project tests. |
| Governance drift | Cited references, rules, or parent responsibilities change. | Re-read Governance rows and parent container docs. | Run c3x verify plus targeted lookup for changed files. |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| Code, docs, tests, prompts | Goal, Governance, Contract, and Change Safety sections. | Names and framework shape may vary; behavior and boundaries may not. | c3x check, c3x verify, and relevant tests. |

## Architecture Details

## Uses

| Component | For |
| --- | --- |
| Flow Pattern | All operations use flow() with namespace + Zod schema |
| auditResource | All audit data access (history, list, export, stats, count) |

## Operations

| Flow | Effect |
| --- | --- |
| getAuditHistory | Returns audit entries for a specific table + record_id. For invoices table, also fetches related invoice_services changes |
| listAuditEntries | Paginated audit log with filters (table_name, action, triggered_by, date range). Returns entries + total count |
| exportAuditTrail | Exports filtered audit data as JSON or CSV string. Default limit 10000 |
| getAuditStats | Returns aggregate stats: count, first/last change grouped by table_name + action |

## Related Tables

The `getAuditHistory` flow automatically includes changes from related tables:

- `invoices` -> includes `invoice_services` changes (via `invoice_id` foreign key)

## Filter Options (listAuditEntries / exportAuditTrail)

table_name, action (CREATE/UPDATE/DELETE), triggered_by (user email), from_date, to_date, limit, offset (list only).

## Audit Entry Shape

id, action, table_name, record_id, record_before (JSON), record_after (JSON), triggered_by (email), triggered_at (timestamp).

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| IN (uses) | See component wiring and code sections in this document. | c3-2 |
| OUT (provides) | Exposes behavior described by this component to dependent callers. | c3-2 |
