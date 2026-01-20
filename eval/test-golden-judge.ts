#!/usr/bin/env bun
/**
 * Test the golden judge by comparing the golden against itself (should score ~100)
 * and against a minimal/empty output (should score low)
 */

import { runGoldenComparison } from "./lib/golden-judge";
import { mkdtemp, rm } from "fs/promises";
import { join } from "path";
import { tmpdir } from "os";

const GOLDEN_PATH = join(import.meta.dir, "goldens/onboard-greenfield");
const GOLDEN_METADATA = join(import.meta.dir, "goldens/onboard-greenfield.yaml");

async function main() {
  const args = process.argv.slice(2);
  const testType = args[0] || "self";
  const verbose = args.includes("--verbose") || args.includes("-v");

  console.log("🧪 Golden Judge Test\n");

  if (testType === "self") {
    // Test 1: Compare golden against itself (should score ~100)
    console.log("Test 1: Golden vs Self (expect ~100)");
    console.log("─".repeat(50));

    const result = await runGoldenComparison(
      GOLDEN_PATH,
      GOLDEN_METADATA,
      GOLDEN_PATH,
      verbose
    );

    console.log(`\nResult: ${result.verdict === "pass" ? "✅ PASS" : "❌ FAIL"}`);
    console.log(`Score: ${result.overall_score}/100`);

    if (result.overall_score < 90) {
      console.warn("\n⚠️  Self-comparison scored below 90 - judge may be too strict");
    }
  } else if (testType === "minimal") {
    // Test 2: Compare against minimal output (should score low)
    console.log("Test 2: Golden vs Minimal (expect low score)");
    console.log("─".repeat(50));

    // Create minimal C3 structure
    const tempDir = await mkdtemp(join(tmpdir(), "golden-test-"));
    const c3Dir = join(tempDir, ".c3");

    await Bun.write(
      join(c3Dir, "README.md"),
      `---
id: c3-0
title: Test Project
---

# Test Project

Just a minimal README.
`
    );

    try {
      const result = await runGoldenComparison(
        GOLDEN_PATH,
        GOLDEN_METADATA,
        c3Dir,
        verbose
      );

      console.log(`\nResult: ${result.verdict === "pass" ? "✅ PASS" : "❌ FAIL"}`);
      console.log(`Score: ${result.overall_score}/100`);

      if (result.overall_score > 50) {
        console.warn("\n⚠️  Minimal comparison scored above 50 - judge may be too lenient");
      }
    } finally {
      await rm(tempDir, { recursive: true });
    }
  } else if (testType === "path") {
    // Test 3: Compare against a provided path
    const outputPath = args[1];
    if (!outputPath) {
      console.error("Usage: bun test-golden-judge.ts path <path-to-c3-dir>");
      process.exit(1);
    }

    console.log(`Test: Golden vs ${outputPath}`);
    console.log("─".repeat(50));

    const result = await runGoldenComparison(
      GOLDEN_PATH,
      GOLDEN_METADATA,
      outputPath,
      verbose
    );

    console.log(`\nResult: ${result.verdict === "pass" ? "✅ PASS" : "❌ FAIL"}`);
    console.log(`Score: ${result.overall_score}/100`);
  } else {
    console.log("Usage:");
    console.log("  bun test-golden-judge.ts self       # Compare golden to itself");
    console.log("  bun test-golden-judge.ts minimal    # Compare golden to minimal output");
    console.log("  bun test-golden-judge.ts path <dir> # Compare golden to specific output");
    console.log("");
    console.log("Add --verbose or -v for detailed output");
  }
}

main().catch(console.error);
