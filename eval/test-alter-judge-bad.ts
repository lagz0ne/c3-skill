import { runAlterJudge, type AlterJudgeInput } from "./lib/alter-judge";
import { join } from "path";

const PLUGIN_ROOT = join(import.meta.dir, "..");

// Bad test case: Incomplete/poor quality alter output
const testInput: AlterJudgeInput = {
  changeRequest: `Add notification routes for sending task reminders.

The notification routes should:
- GET /api/notifications - list user notifications
- POST /api/notifications/:id/read - mark as read
- Depend on auth middleware (c3-102) and database (c3-105)`,

  output: {
    // ADR with problems:
    // - Problem doesn't match request (invents "real-time" requirement)
    // - No alternatives considered
    // - Missing affected layers detail
    // - No verification steps
    adr: `---
id: adr-20260113-notifications
title: Notifications
status: proposed
---

# Notifications

## Problem

We need real-time notifications with websockets.

## Decision

Add notifications.

## Affected Layers

Will affect some stuff.
`,

    // No plan provided
    plan: undefined,

    // Component doc with problems:
    // - Wrong dependencies (doesn't mention c3-102 or c3-105)
    // - No hand-offs
    // - Missing endpoints detail
    componentDocs: [
      `# Notifications

## Purpose

Handle notifications.

## Location

src/routes/notifications.ts
`,
    ],
  },
};

async function main() {
  const verbose = process.argv.includes("--verbose");

  console.log("Testing alter judge with BAD alter output...\n");
  console.log("Change Request:");
  console.log(testInput.changeRequest);
  console.log("\n---");
  console.log("\nExpected problems:");
  console.log("- ADR problem doesn't match request (invents websockets)");
  console.log("- No alternatives in rationale");
  console.log("- No plan provided");
  console.log("- Component missing dependencies (c3-102, c3-105)");
  console.log("- No hand-offs or endpoints detail");
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

  for (const p of result.purposes) {
    const concerns = p.concerns;
    const addressed = concerns.filter((c) => c.status === "addressed").length;
    const partial = concerns.filter((c) => c.status === "partial").length;
    const missing = concerns.filter((c) => c.status === "missing").length;
    console.log(`  ${p.purpose}: ${p.rating} (${addressed}✓ ${partial}~ ${missing}✗)`);
  }
}

main().catch(console.error);
