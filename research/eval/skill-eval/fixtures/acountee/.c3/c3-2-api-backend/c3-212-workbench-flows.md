---
id: c3-212
c3-version: 3
c3-seal: 47efa3018be5952e583e5a126b14c675c5360c0eb95d505ce22aa2c7f593d359
title: Workbench Flows
type: component
category: feature
parent: c3-2
goal: Finance team workbench - invoice cleanup, PR export with customizable templates, paid PR import
uses:
    - ref-pumped-fn
    - ref-query-services
    - ref-server-functions
    - ref-structured-logging
    - ref-sync
---

# Workbench Flows

## Goal

Finance team workbench - invoice cleanup, PR export with customizable templates, paid PR import

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-2 |
| Role | Own Workbench Flows behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep Workbench Flows decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for Workbench Flows so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before Workbench Flows behavior is changed. | ref-pumped-fn |
| Inputs | Accept only the files, commands, data, or calls that belong to Workbench Flows ownership. | ref-pumped-fn |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-pumped-fn |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-pumped-fn |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks Workbench Flows to deliver its documented responsibility. | ref-pumped-fn |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-pumped-fn |
| Alternate paths | When a request falls outside Workbench Flows ownership, hand it to the parent or sibling component. | ref-pumped-fn |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-pumped-fn |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-pumped-fn | ref | Governs Workbench Flows behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| Workbench Flows input | IN | Callers must provide context that matches the component goal and parent fit. | c3-2 boundary | c3x lookup plus targeted tests or review. |
| Workbench Flows output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-2 boundary | c3x check and project test suite. |

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
| drizzleDb | Direct Drizzle queries (respects transaction tag) |
| invoiceService | markAsRedundant for bulk obsolete |
| prService | complete for paid PR import |
| prQueries | getPr, updatePaymentReference |
| exportTemplateQueries | CRUD for user export templates |
| fileStorage | Store/retrieve uploaded XLSX template files |
| sync | Client notification via NATS after mutations |
| templateGenerator | Generate XLSX/CSV from template definitions |
| templateProcessor | Process uploaded XLSX templates with placeholder scanning |

## Operations

| Flow | Input | Effect | Side Effects |
| --- | --- | --- | --- |
| listInvoicesForCleanup | year, month | Queries invoices where status='imported' within the month. Returns summary (id, invoiceNo, invoiceName, supplierName, amount, invoiceDate) | - |
| bulkMarkObsolete | invoiceIds[] | Iterates IDs, calls invoiceService.markAsRedundant for each. Returns per-item results | sync |
| listApprovedPrsForExport | year, month | Queries PRs where status='approved' with month filtering via invoice dates. Resolves bank info: advanced PRs use own fields, direct PRs use linked invoice payment fields | - |
| importPaidPrs | items[{prId, paymentReference}] | For each: validates PR exists + status='approved', sets payment_reference, completes PR. Returns per-item results | sync |
| saveTemplate | name, format, definition, templateId? | Create or update a CSV/XLSX export template definition. Validates via ExportTemplateDefinitionSchema | - |
| listUserTemplates | (none) | Returns all export templates owned by current user | - |
| uploadXlsxTemplate | name, templateId?, file | Upload XLSX file, scan for {{field}} placeholders, store file and metadata | - |
| exportWithTemplate | templateId, year, month, prIds? | Export approved PRs using template. templateId=0 uses golden default | - |
| deleteTemplate | templateId | Delete user template and clean up stored source file if present | - |

## Export Template System

### Template Types

1. **Golden default** (templateId=0): Built-in XLSX with formatted headers, borders, frozen rows, SUM formula
2. **Definition-based**: JSON schema defining headers, data columns, summaries, cell styles
3. **XLSX upload**: User uploads XLSX with `{{field}}` placeholders in data band rows; placeholders are expanded per PR

### Export Fields (13)

`rowNumber`, `id`, `name`, `type`, `amount`, `supplierName`, `bankName`, `bankAccount`, `accountName`, `note`, `maker`, `paymentReference`, `static`

### Template Processing

- **Definition templates**: `generateFromTemplate()` creates workbook from `ExportTemplateDefinition` schema
- **XLSX uploads**: `processXlsxTemplate()` finds data band (rows with `{{field}}`), expands per PR, preserves cell styles and formula offsets
- **Placeholder scanning**: `scanPlaceholders()` extracts `{{fieldName}}` patterns from uploaded XLSX

## Bank Info Resolution (listApprovedPrsForExport)

Uses SQL CASE to resolve bank details based on PR type:

- **advanced**: Uses `pr.advanced_bank_name/account/account_name`
- **direct**: Uses linked invoice's `payment_bank_name/account/account_name`

Fallback for PRs without invoices: checks audit table for activity within the date range.

## Import Paid PRs Pipeline

1. Look up PR, validate status is `approved`
2. Set `payment_reference` on PR record
3. Complete PR via `prService.complete`
4. Collect results (success/failure with reason per item)

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| IN (uses) | See component wiring and code sections in this document. | c3-2 |
| OUT (provides) | Exposes behavior described by this component to dependent callers. | c3-2 |
