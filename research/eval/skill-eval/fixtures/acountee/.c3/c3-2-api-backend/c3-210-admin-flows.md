---
id: c3-210
c3-version: 3
c3-seal: 501252f6bf2205cec1c9fbf7832158e4ced9da4e87d881c9a8da7864e4e0f8cc
title: Admin Flows
type: component
category: feature
parent: c3-2
goal: Admin management flows - teams, roles, users, approval configuration
uses:
    - ref-pumped-fn
    - ref-query-services
    - ref-rbac
    - ref-server-functions
    - ref-structured-logging
    - ref-sync
---

# Admin Flows

## Goal

Admin management flows - teams, roles, users, approval configuration

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-2 |
| Role | Own Admin Flows behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep Admin Flows decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for Admin Flows so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before Admin Flows behavior is changed. | ref-pumped-fn |
| Inputs | Accept only the files, commands, data, or calls that belong to Admin Flows ownership. | ref-pumped-fn |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-pumped-fn |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-pumped-fn |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks Admin Flows to deliver its documented responsibility. | ref-pumped-fn |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-pumped-fn |
| Alternate paths | When a request falls outside Admin Flows ownership, hand it to the parent or sibling component. | ref-pumped-fn |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-pumped-fn |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-pumped-fn | ref | Governs Admin Flows behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| Admin Flows input | IN | Callers must provide context that matches the component goal and parent fit. | c3-2 boundary | c3x lookup plus targeted tests or review. |
| Admin Flows output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-2 boundary | c3x check and project test suite. |

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
| rbacQueries | Owner check, role CRUD, user-role assignment |
| teamQueries | Team CRUD, member count |
| userQueries | User CRUD, associated records check |
| approvalConfigQueries | Approval flow CRUD |
| auditQueries | Explicit audit entries for team/role/approval mutations |

## Authorization

Every admin flow: extract user from `currentUserTag` -> check `rbacQueries.isOwner` -> reject with `NOT_OWNER` if not.

## Team Operations

| Flow | Effect | Guards | Audit |
| --- | --- | --- | --- |
| listTeams | Returns all teams with capabilities | owner | - |
| createTeam | Creates team (name, description, capabilities), records creator | owner | create on teams |
| updateTeam | Updates team fields | owner, team must exist | update on teams |
| deleteTeam | Deletes team | owner, team must exist, blocks if has members (TEAM_HAS_USERS) | delete on teams |

## Role Operations

| Flow | Effect | Guards | Audit |
| --- | --- | --- | --- |
| listRoles | Returns all roles with permissions | owner | - |
| createRole | Creates role (name, description, permissions, parent_role_id) | owner, name must be unique (ROLE_EXISTS) | create on roles |
| updateRole | Updates role fields | owner, role must exist, blocks owner role (CANNOT_MODIFY_OWNER_ROLE) | update on roles |
| deleteRole | Deletes role | owner, role must exist, blocks owner role, blocks if assigned (ROLE_HAS_USERS) | delete on roles |
| assignRole | Assigns role to user by email | owner, role must exist | assign_role on user_roles |
| revokeRole | Revokes role from user | owner, role must exist, protects last owner (CANNOT_REVOKE_LAST_OWNER) | revoke_role on user_roles |

## User Operations

| Flow | Effect | Guards |
| --- | --- | --- |
| listUsers | Returns all users enriched with roles (batch fetched, no N+1) | owner |
| createUser | Creates user (email, permissions, team) | owner |
| updateUser | Updates permissions, team, and/or status independently | owner |
| deleteUser | Soft (set inactive) or hard delete (checks hasAssociatedRecords) | owner, protects last owner |
| transferOwnership | Assigns owner role to target user | owner |
| removeOwnership | Removes owner role from target | owner, target must be owner, protects last owner |

## Approval Config Operations

| Flow | Effect | Guards | Audit |
| --- | --- | --- | --- |
| listApprovalFlows | Returns all flows with steps | owner | - |
| getApprovalFlow | Returns single flow by id | owner, flow must exist | - |
| updateApprovalFlow | Updates flow name, description, and steps (stepNumber, name, mode: anyof/allof, userEmails) | owner, flow must exist | update on approval_flows |
| toggleApprovalFlowActive | Flips is_active flag | owner, flow must exist | update on approval_flows |

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| IN (uses) | See component wiring and code sections in this document. | c3-2 |
| OUT (provides) | Exposes behavior described by this component to dependent callers. | c3-2 |
