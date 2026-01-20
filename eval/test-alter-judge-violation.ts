import { runAlterJudge, type AlterJudgeInput } from "./lib/alter-judge";
import { join } from "path";

const PLUGIN_ROOT = join(import.meta.dir, "..");

// Workflow violation test case:
// - ADR says one thing (add caching)
// - Component doc says something different (add notifications)
// - Plan doesn't match ADR
// This tests if judge catches when output doesn't follow the ADR
const testInput: AlterJudgeInput = {
  changeRequest: `Add notification routes for sending task reminders.

The notification routes should:
- GET /api/notifications - list user notifications
- POST /api/notifications/:id/read - mark as read
- Depend on auth middleware (c3-102) and database (c3-105)`,

  output: {
    // ADR is about a DIFFERENT feature (caching instead of notifications)
    adr: `---
id: adr-20260113-caching
title: Add Redis Caching
status: accepted
date: 2026-01-13
affects: [c3-1]
---

# Add Redis Caching

## Status

**Accepted** - 2026-01-13

## Problem

API needs caching to reduce database load.

## Decision

Add Redis caching layer for frequently accessed data.

## Rationale

Caching improves performance.

## Affected Layers

| Layer | Document | Change |
|-------|----------|--------|
| Container | c3-1-api/README.md | Add c3-106 |
| Component | c3-106-cache.md | Create |

## Verification

- [ ] Cache component created
`,

    // Plan also doesn't match change request
    plan: `---
adr: adr-20260113-caching
status: pending
---

# Plan: Redis Caching

1. Create cache component
2. Update container README
`,

    // Component is about notifications but doesn't match ADR
    componentDocs: [
      `---
id: c3-106
title: Notifications
type: component
category: feature
parent: c3-1
goal: Handle notifications
---

# Notifications

## Goal

Notification routes for task reminders.

## Endpoints

| Method | Path |
|--------|------|
| GET | /api/notifications |
| POST | /api/notifications/:id/read |
`,
    ],
  },
};

async function main() {
  const verbose = process.argv.includes("--verbose");

  console.log("Testing alter judge with WORKFLOW VIOLATION...\n");
  console.log("Change Request: Add notification routes");
  console.log("\n---");
  console.log("\nViolations:");
  console.log("- ADR is about CACHING (not notifications)");
  console.log("- Plan references caching ADR");
  console.log("- Component is about notifications but doesn't match ADR");
  console.log("- ADR problem doesn't match change request");
  console.log("\nThis tests: Does judge catch when ADR doesn't match request?");
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
  console.log(`\nExpected: 'reject' or 'rework' (ADR doesn't match request)`);

  for (const p of result.purposes) {
    const concerns = p.concerns;
    const addressed = concerns.filter((c) => c.status === "addressed").length;
    const partial = concerns.filter((c) => c.status === "partial").length;
    const missing = concerns.filter((c) => c.status === "missing").length;
    console.log(`  ${p.purpose}: ${p.rating} (${addressed}✓ ${partial}~ ${missing}✗)`);
  }

  // Check if judge caught the mismatch
  const problemValidity = result.purposes
    .flatMap((p) => p.concerns)
    .find((c) => c.concern.includes("problem") || c.concern.includes("validity"));

  console.log("\n=== Mismatch Detection ===");
  if (problemValidity) {
    console.log(`Problem validity status: ${problemValidity.status}`);
    if (problemValidity.status === "missing" || problemValidity.status === "partial") {
      console.log("✓ Judge correctly caught ADR/request mismatch");
    } else {
      console.log("✗ WARNING: Judge missed the ADR/request mismatch!");
    }
    if (problemValidity.gap) {
      console.log(`Gap noted: ${problemValidity.gap}`);
    }
  }
}

main().catch(console.error);
