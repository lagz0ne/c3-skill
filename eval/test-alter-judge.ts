import { runAlterJudge, type AlterJudgeInput } from "./lib/alter-judge";
import { join } from "path";

const PLUGIN_ROOT = join(import.meta.dir, "..");

// Sample test case: Add notifications to documented-api
const testInput: AlterJudgeInput = {
  changeRequest: `Add notification routes for sending task reminders.

The notification routes should:
- GET /api/notifications - list user notifications
- POST /api/notifications/:id/read - mark as read
- Depend on auth middleware (c3-102) and database (c3-105)`,

  output: {
    adr: `---
id: adr-20260113-notifications
title: Add Notification Routes
status: accepted
date: 2026-01-13
affects: [c3-1]
---

# Add Notification Routes

## Status

**Accepted** - 2026-01-13

## Problem

Users need to receive task reminders and view their notifications. Currently there is no notification system in the API.

## Decision

Add a new notification component (c3-106) to handle notification CRUD operations.

## Rationale

| Considered | Rejected Because |
|------------|------------------|
| WebSocket notifications | Adds complexity, polling sufficient for MVP |
| Email-only notifications | Users need in-app view |

## Affected Layers

| Layer | Document | Change |
|-------|----------|--------|
| Container | c3-1-api/README.md | Add c3-106 to components table |
| Component | c3-106-notifications.md | New component |

## Verification

- [ ] GET /api/notifications returns user notifications
- [ ] POST /api/notifications/:id/read marks notification as read
- [ ] Auth middleware protects routes
`,

    plan: `---
adr: adr-20260113-notifications
status: pending
---

# Plan: Add Notification Routes

> Implements [adr-20260113-notifications](./adr-20260113-notifications.md)

## Pre-execution checklist

- [ ] Update References in affected components

## Changes

### c3-1 (Container)

**File:** \`.c3/c3-1-api/README.md\`

| Section | Action | Change |
|---------|--------|--------|
| Components | Add row | c3-106 | notifications | feature | Notification CRUD |

### c3-106 (Component) - NEW

**File:** \`.c3/c3-1-api/c3-106-notifications.md\`

Action: Create from component template

| Section | Content |
|---------|---------|
| Goal | Handle notification CRUD for task reminders |
| Dependencies | c3-102 (auth), c3-105 (database) |
| Endpoints | GET /api/notifications, POST /api/notifications/:id/read |

## Order

1. Create c3-106-notifications.md
2. Update c3-1-api/README.md components table

## Verification

| From ADR | How to Verify |
|----------|---------------|
| c3-106 exists | \`ls .c3/c3-1-api/c3-106-*.md\` |
| Component in inventory | \`grep "c3-106" .c3/c3-1-api/README.md\` |
`,

    componentDocs: [
      `---
id: c3-106
c3-version: 3
title: Notifications
type: component
category: feature
parent: c3-1
goal: Handle notification CRUD for task reminders
summary: REST endpoints for listing and managing user notifications
---

# Notifications

## Goal

Handle notification CRUD for task reminders, allowing users to view and acknowledge notifications.

## Container Connection

Extends c3-1 (API) with user-facing notification endpoints. Without this, users cannot see or interact with task reminders.

## Hand-offs

| Direction | What | From/To |
|-----------|------|---------|
| IN | userId | c3-102 (auth-middleware) |
| IN | db connection | c3-105 (database) |
| OUT | notification list | HTTP response |
| OUT | read confirmation | HTTP response |

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/notifications | List user's notifications |
| POST | /api/notifications/:id/read | Mark notification as read |

## Dependencies

- c3-102 (auth-middleware) - for userId extraction
- c3-105 (database) - for notification persistence
`,
    ],
  },
};

async function main() {
  const verbose = process.argv.includes("--verbose");

  console.log("Testing alter judge with sample notification feature...\n");
  console.log("Change Request:");
  console.log(testInput.changeRequest);
  console.log("\n---");

  // Load templates as ground truth
  try {
    testInput.templates = {
      adrTemplate: await Bun.file(join(PLUGIN_ROOT, "references/adr-template.md")).text(),
      planTemplate: await Bun.file(join(PLUGIN_ROOT, "references/plan-template.md")).text(),
      componentTemplate: await Bun.file(join(PLUGIN_ROOT, "templates/component.md")).text(),
    };
    console.log("Loaded templates as ground truth\n");
  } catch (e) {
    console.log(`Warning: Could not load templates: ${e}\n`);
  }

  const result = await runAlterJudge(testInput, verbose);

  // Summary
  console.log("\n=== Summary ===");
  console.log(`Verdict: ${result.verdict}`);
  console.log(`Purposes evaluated: ${result.purposes.length}`);

  for (const p of result.purposes) {
    const concerns = p.concerns;
    const addressed = concerns.filter((c) => c.status === "addressed").length;
    const partial = concerns.filter((c) => c.status === "partial").length;
    const missing = concerns.filter((c) => c.status === "missing").length;
    console.log(`  ${p.purpose}: ${p.rating} (${addressed}✓ ${partial}~ ${missing}✗)`);
  }
}

main().catch(console.error);
