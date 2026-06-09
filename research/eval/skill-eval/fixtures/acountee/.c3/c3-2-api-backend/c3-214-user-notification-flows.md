---
id: c3-214
c3-version: 3
c3-seal: bc4bd49444161673895d8ee4643599427da2d89529f810cd449e760bf743be86
title: User Notification Flows
type: component
category: feature
parent: c3-2
goal: User-facing notification operations - fetch, read, dismiss
uses:
    - ref-pumped-fn
    - ref-query-services
    - ref-server-functions
    - ref-structured-logging
---

# User Notification Flows

## Goal

User-facing notification operations - fetch, read, dismiss

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-2 |
| Role | Own User Notification Flows behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep User Notification Flows decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for User Notification Flows so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before User Notification Flows behavior is changed. | ref-pumped-fn |
| Inputs | Accept only the files, commands, data, or calls that belong to User Notification Flows ownership. | ref-pumped-fn |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-pumped-fn |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-pumped-fn |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks User Notification Flows to deliver its documented responsibility. | ref-pumped-fn |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-pumped-fn |
| Alternate paths | When a request falls outside User Notification Flows ownership, hand it to the parent or sibling component. | ref-pumped-fn |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-pumped-fn |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-pumped-fn | ref | Governs User Notification Flows behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| User Notification Flows input | IN | Callers must provide context that matches the component goal and parent fit. | c3-2 boundary | c3x lookup plus targeted tests or review. |
| User Notification Flows output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-2 boundary | c3x check and project test suite. |

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
| notificationQueries | All notification data access (list logs, read/dismiss state) |

## Operations

| Flow | Input | Effect |
| --- | --- | --- |
| getNotificationsFlow | (none) | Fetches up to 50 non-dismissed in-app notifications for current user. Returns hasMore flag |
| markNotificationReadFlow | notificationId | Marks single notification as read |
| markAllNotificationsReadFlow | notificationIds[] | Marks multiple notifications as read. Returns count |
| dismissNotificationFlow | notificationId | Dismisses single notification (filtered out, not deleted) |
| dismissAllNotificationsFlow | notificationIds[] | Dismisses multiple notifications. Returns count |

## Fetch Strategy

`getNotificationsFlow` compensates for dismissed notifications by over-fetching:

1. Fetch dismissed IDs for user
2. Fetch `min(51 + dismissedCount, 200)` notification logs (channel='in_app')
3. Fetch read IDs for user
4. Filter out dismissed, cap at 50, enrich with read state

Read and dismiss are independent states -- a notification can be read but not dismissed, or dismissed without being read.

All flows require authenticated user via `currentUserTag`. All results include `executionId` for sync coordination.

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| IN (uses) | See component wiring and code sections in this document. | c3-2 |
| OUT (provides) | Exposes behavior described by this component to dependent callers. | c3-2 |
