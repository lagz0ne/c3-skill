---
id: c3-205
c3-version: 3
c3-seal: be86adfa4330a21e8cf07a9a1c2b2ac6aca7bbe3785578d3d8823698aa09b120
title: PR Flows
type: component
category: feature
parent: c3-2
goal: Payment request business logic - create, approve, reject, complete
uses:
    - ref-approval-chain
    - ref-pumped-fn
    - ref-query-services
    - ref-rbac
    - ref-server-functions
    - ref-structured-logging
    - ref-sync
---

# PR Flows

## Goal

Payment request business logic - create, approve, reject, complete

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-2 |
| Role | Own PR Flows behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep PR Flows decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for PR Flows so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before PR Flows behavior is changed. | ref-approval-chain |
| Inputs | Accept only the files, commands, data, or calls that belong to PR Flows ownership. | ref-approval-chain |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-approval-chain |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-approval-chain |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks PR Flows to deliver its documented responsibility. | ref-approval-chain |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-approval-chain |
| Alternate paths | When a request falls outside PR Flows ownership, hand it to the parent or sibling component. | ref-approval-chain |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-approval-chain |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-approval-chain | ref | Governs PR Flows behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| PR Flows input | IN | Callers must provide context that matches the component goal and parent fit. | c3-2 boundary | c3x lookup plus targeted tests or review. |
| PR Flows output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-2 boundary | c3x check and project test suite. |

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
| prService | Core PR mutations (create, approve, complete, revert, attachments) |
| notificationService | Notify next approvers on step advance |
| sync | Client notification via NATS after mutations |

## Operations

| Flow | Effect | Side Effects |
| --- | --- | --- |
| createPr | Creates PR in draft with name, type (direct/advanced), amount, note, approval flow config, optional bank details and attachments | sync |
| updatePr | Updates PR fields (name, type, amount, note, bank details, approval config) | sync |
| requestApprovals | Transitions draft to pending via prService.requestForApprovals | sync, notifies next approvers |
| approvePr | Records current user's approval; if step advances, notifies next approvers | sync, conditional notification |
| unapprovePr | Removes current user's approval | sync |
| rejectPr | Reverts PR from pending back to draft via prService.revertFromRequestingForApprovals | sync |
| recallPr | Same as reject -- maker reverts their own PR to draft | sync |
| completePr | Marks PR as completed (after all approvals) | sync |
| uncompletePr | Reverses completion via prService.unmarkAsComplete | sync |
| removePr | Deletes PR via prService.remove | sync |
| approveAll | Bulk approval: iterates pr_ids, approves each, collects approved/failed arrays | sync, conditional notifications per PR |
| updatePrAttachments | Replaces PR attachments (filters empty files) | sync |
| removePrAttachment | Removes single attachment by name | sync |
| getPr | Returns single PR by id via prQueries.getPr | - |
| listPrs | Returns all PRs via prQueries.listPr | - |

## State Machine

```
draft --> pending --> approved --> completed
  ^         |            ^           |
  |         v            |           v
  +--- rejected          +--- uncompleted
  +--- recalled
draft --> [deleted]
```

## PR Types

- **direct**: Bank info resolved from linked invoices
- **advanced**: Bank info stored directly on PR (`advanced_bank_name`, `advanced_bank_account`, `advanced_account_name`)

## Approval Integration

Each PR carries an approval flow config (embedded JSON). On `requestApprovals`, the PR moves to pending and first-step approvers are notified. On `approvePr`, if a step completes (`stepAdvanced`), next-step approvers are notified. Notifications fire async with error suppression (logged, not thrown).

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| IN (uses) | See component wiring and code sections in this document. | c3-2 |
| OUT (provides) | Exposes behavior described by this component to dependent callers. | c3-2 |
