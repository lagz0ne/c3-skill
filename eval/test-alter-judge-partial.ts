import { runAlterJudge, type AlterJudgeInput } from "./lib/alter-judge";
import { join } from "path";

const PLUGIN_ROOT = join(import.meta.dir, "..");

// Partially correct test case:
// - ADR is good (problem matches, decision clear)
// - BUT: No plan provided
// - BUT: Component doc is incomplete (missing hand-offs)
// - BUT: Dependencies mentioned but not detailed
const testInput: AlterJudgeInput = {
  changeRequest: `Add notification routes for sending task reminders.

The notification routes should:
- GET /api/notifications - list user notifications
- POST /api/notifications/:id/read - mark as read
- Depend on auth middleware (c3-102) and database (c3-105)`,

  output: {
    // ADR is reasonably good
    adr: `---
id: adr-20260113-notifications
title: Add Notification Routes
status: accepted
date: 2026-01-13
affects: [c3-1, c3-106]
---

# Add Notification Routes

## Status

**Accepted** - 2026-01-13

## Problem

Users need to receive task reminders and view their notifications. Currently there is no notification system.

## Decision

Add notification routes:
- GET /api/notifications - list notifications
- POST /api/notifications/:id/read - mark as read

New component c3-106 for notification functionality.

## Rationale

| Considered | Rejected Because |
|------------|------------------|
| WebSockets | Too complex for MVP |
| Email-only | Users need in-app view |

## Affected Layers

| Layer | Document | Change |
|-------|----------|--------|
| Container | c3-1-api/README.md | Add c3-106 |
| Component | c3-106-notifications.md | Create |

## Verification

- [ ] Component doc exists
- [ ] Container README updated
`,

    // NO PLAN - this is a gap
    plan: undefined,

    // Component doc exists but is incomplete
    componentDocs: [
      `---
id: c3-106
title: Notifications
---

# Notifications

## Purpose

Handle notification routes.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/notifications | List notifications |
| POST | /api/notifications/:id/read | Mark as read |

## Dependencies

Uses auth and database.
`,
      // Note: Missing hand-offs section, missing detailed dependencies (c3-102, c3-105 not specified)
    ],
  },
};

async function main() {
  const verbose = process.argv.includes("--verbose");

  console.log("Testing alter judge with PARTIAL quality output...\n");
  console.log("Change Request:");
  console.log(testInput.changeRequest);
  console.log("\n---");
  console.log("\nWhat's good:");
  console.log("- ADR problem matches request");
  console.log("- ADR has decision with endpoints");
  console.log("- ADR has rationale with alternatives");
  console.log("- Component doc exists with endpoints");
  console.log("\nWhat's missing/weak:");
  console.log("- NO PLAN provided");
  console.log("- Component missing hand-offs section");
  console.log("- Dependencies vague ('auth and database' not 'c3-102, c3-105')");
  console.log("- Missing verification detail");
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
  console.log(`\nExpected: 'conditional' or 'rework' (not 'approve' or 'reject')`);

  for (const p of result.purposes) {
    const concerns = p.concerns;
    const addressed = concerns.filter((c) => c.status === "addressed").length;
    const partial = concerns.filter((c) => c.status === "partial").length;
    const missing = concerns.filter((c) => c.status === "missing").length;
    console.log(`  ${p.purpose}: ${p.rating} (${addressed}✓ ${partial}~ ${missing}✗)`);
  }

  // Check if judge has proper granularity
  const hasPartial = result.purposes.some((p) =>
    p.concerns.some((c) => c.status === "partial")
  );
  const hasMissing = result.purposes.some((p) =>
    p.concerns.some((c) => c.status === "missing")
  );

  console.log("\n=== Granularity Check ===");
  console.log(`Has 'partial' ratings: ${hasPartial ? "YES ✓" : "NO ✗"}`);
  console.log(`Has 'missing' ratings: ${hasMissing ? "YES ✓" : "NO ✗"}`);

  if (result.verdict === "approve" || result.verdict === "exemplary") {
    console.log("\n⚠️  WARNING: Judge may be too lenient - this should NOT be approve");
  } else if (result.verdict === "reject") {
    console.log("\n⚠️  WARNING: Judge may be too strict - this should NOT be reject");
  } else {
    console.log("\n✓ Judge granularity looks correct");
  }
}

main().catch(console.error);
