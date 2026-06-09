---
id: c3-206
c3-version: 3
c3-seal: 701f74023237b0bb700d6cf91ade3351de576381ce23a803e7cae7a69238425f
title: Invoice Flows
type: component
category: feature
parent: c3-2
goal: Invoice business logic - import XML/ZIP, link/unlink PRs, manage status
uses:
    - ref-file-handling
    - ref-pumped-fn
    - ref-query-services
    - ref-server-functions
    - ref-structured-logging
    - ref-sync
---

# Invoice Flows

## Goal

Invoice business logic - import XML/ZIP, link/unlink PRs, manage status

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-2 |
| Role | Own Invoice Flows behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep Invoice Flows decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for Invoice Flows so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before Invoice Flows behavior is changed. | ref-file-handling |
| Inputs | Accept only the files, commands, data, or calls that belong to Invoice Flows ownership. | ref-file-handling |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-file-handling |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-file-handling |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks Invoice Flows to deliver its documented responsibility. | ref-file-handling |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-file-handling |
| Alternate paths | When a request falls outside Invoice Flows ownership, hand it to the parent or sibling component. | ref-file-handling |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-file-handling |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-file-handling | ref | Governs Invoice Flows behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| Invoice Flows input | IN | Callers must provide context that matches the component goal and parent fit. | c3-2 boundary | c3x lookup plus targeted tests or review. |
| Invoice Flows output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-2 boundary | c3x check and project test suite. |

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
| invoiceService | Link/unlink PR, update payment details, mark redundant/imported |
| fileStorage | Store and retrieve uploaded files |
| invoiceQueries | Insert invoices, deduplication hash check, insert services |
| sync | Client notification via NATS after mutations |

## Operations

| Flow | Effect | Side Effects |
| --- | --- | --- |
| linkPaymentRequest | Links invoice to PR via invoiceService.linkToPr | sync |
| unlinkPaymentRequest | Removes PR link via invoiceService.unlinkFromPr | sync |
| updatePaymentDetail | Updates invoice payment info (type, bank name, account, account name) | sync |
| changePrToOther | Atomic unlink old PR + link new PR | sync |
| markInvoiceAsRedundant | Sets invoice as redundant via invoiceService.markAsRedundant | sync |
| markInvoiceAsImported | Reverts invoice to imported via invoiceService.unmarkAsComplete | sync |
| importFiles | Processes uploaded XML/ZIP files, parses invoices, deduplicates, inserts. Uses mapServicesToInsertArgs helper for bulk service insertion | sync |
| getInvoice | Returns single invoice by id | - |
| listInvoices | Returns all invoices | - |

## Import Flow

1. Store each uploaded file via `fileStorage.store` (skip if duplicate file)
2. Retrieve stored file, detect type (XML or ZIP)
3. **XML**: Parse invoice, compute MD5 hash, check for duplicate hash in DB, insert invoice + line-item services
4. **ZIP**: Extract entries, process each XML file individually (same parse/hash/insert logic)
5. Deduplication uses MD5 of raw XML content against `invoiceQueries.countInvoiceByHashValue`

Import results per file: `success` (inserted), `skipped` (duplicate hash), or `failure` (parse/insert error). ZIP imports can produce `partial` state when some entries succeed and others fail.

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| IN (uses) | See component wiring and code sections in this document. | c3-2 |
| OUT (provides) | Exposes behavior described by this component to dependent callers. | c3-2 |
