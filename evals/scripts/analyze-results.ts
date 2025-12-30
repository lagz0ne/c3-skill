#!/usr/bin/env bun
/**
 * Analyze eval results and generate improvement suggestions
 * Usage: bun analyze-results.ts <results-dir>
 */

import { readFile } from "fs/promises";
import { existsSync } from "fs";
import { join } from "path";

interface CheckResult {
  id: string;
  status: "pass" | "fail";
  reason?: string;
}

interface ChecklistResult {
  passed: number;
  failed: number;
  total: number;
  checks: CheckResult[];
}

interface AnalysisResult {
  score: number;
  passed: number;
  total: number;
  failures: {
    id: string;
    reason: string;
    category: string;
    suggestion: string;
  }[];
  summary: string;
  next_actions: string[];
}

const resultsDir = Bun.argv[2];

if (!resultsDir) {
  console.error("Usage: bun analyze-results.ts <results-dir>");
  process.exit(1);
}

// Categorize failures and suggest fixes
function categorizeFailure(check: CheckResult): { category: string; suggestion: string } {
  const id = check.id;
  const reason = check.reason || "";

  // Frontmatter issues
  if (id.includes("frontmatter")) {
    if (reason.includes("c3-version")) {
      return {
        category: "frontmatter",
        suggestion: "Update skill to use c3-version: 3 (not 1.0.0 or other)"
      };
    }
    if (reason.includes("Missing fields")) {
      const fields = reason.match(/Missing fields: (.+)/)?.[1] || "";
      return {
        category: "frontmatter",
        suggestion: `Add missing frontmatter fields: ${fields}`
      };
    }
    return {
      category: "frontmatter",
      suggestion: "Check frontmatter format in references/structure-guide.md"
    };
  }

  // Section issues
  if (id.includes("sections") || id.includes("section")) {
    return {
      category: "sections",
      suggestion: `Add missing sections. Check exact names in references/structure-guide.md`
    };
  }

  // Table format issues
  if (id.includes("table") || id.includes("inventory")) {
    if (id.includes("containers")) {
      return {
        category: "table-format",
        suggestion: "Use 3-column table: | ID | Name | Responsibility |"
      };
    }
    if (id.includes("container-inventory")) {
      return {
        category: "table-format",
        suggestion: "Use 5-column table: | ID | Name | Type | Responsibility | Status |"
      };
    }
    return {
      category: "table-format",
      suggestion: "Check table format in references/structure-guide.md"
    };
  }

  // Diagram issues
  if (id.includes("mermaid")) {
    return {
      category: "diagrams",
      suggestion: "Use Mermaid syntax for diagrams, not ASCII art"
    };
  }

  // Structure issues
  if (id.includes("exists")) {
    return {
      category: "structure",
      suggestion: `Create required file/directory: ${reason}`
    };
  }

  return {
    category: "other",
    suggestion: `Fix: ${reason}`
  };
}

async function main() {
  const checklistPath = join(resultsDir, "checklist-result.json");

  if (!existsSync(checklistPath)) {
    console.error(`Checklist results not found: ${checklistPath}`);
    process.exit(1);
  }

  const checklist: ChecklistResult = JSON.parse(
    await readFile(checklistPath, "utf-8")
  );

  const score = Math.round((checklist.passed / checklist.total) * 100);

  const failures = checklist.checks
    .filter((c) => c.status === "fail")
    .map((c) => {
      const { category, suggestion } = categorizeFailure(c);
      return {
        id: c.id,
        reason: c.reason || "Unknown",
        category,
        suggestion
      };
    });

  // Group by category
  const byCategory: Record<string, typeof failures> = {};
  for (const f of failures) {
    if (!byCategory[f.category]) byCategory[f.category] = [];
    byCategory[f.category].push(f);
  }

  // Generate next actions
  const nextActions: string[] = [];

  if (byCategory["frontmatter"]) {
    nextActions.push("Fix frontmatter format in skill (c3-version: 3, add type/parent)");
  }
  if (byCategory["table-format"]) {
    nextActions.push("Update table formats to match spec (ID|Name|Type|Responsibility|Status)");
  }
  if (byCategory["sections"]) {
    nextActions.push("Ensure exact section names (## Overview, ## Containers, etc.)");
  }
  if (byCategory["diagrams"]) {
    nextActions.push("Use Mermaid for all diagrams");
  }
  if (byCategory["structure"]) {
    nextActions.push("Create missing files/directories");
  }

  // Generate summary
  let summary: string;
  if (score === 100) {
    summary = "All checks passed. Ready for rubric evaluation.";
  } else if (score >= 80) {
    summary = `Good progress (${score}%). Minor fixes needed.`;
  } else if (score >= 50) {
    summary = `Partial success (${score}%). Several format issues to fix.`;
  } else {
    summary = `Significant issues (${score}%). Review skill output structure.`;
  }

  const result: AnalysisResult = {
    score,
    passed: checklist.passed,
    total: checklist.total,
    failures,
    summary,
    next_actions: nextActions
  };

  console.log(JSON.stringify(result, null, 2));
}

main().catch((e) => {
  console.error(`Analysis error: ${e}`);
  process.exit(1);
});
