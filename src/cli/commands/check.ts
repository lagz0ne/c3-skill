import * as path from "node:path";
import type { CliOptions } from "../context";
import { graphAtom } from "../context";
import type { Lite } from "@pumped-fn/lite";
import { showHelp } from "../help";
import { renderJson } from "../output";

interface Issue {
  severity: "error" | "warning";
  entity?: string;
  message: string;
}

interface CheckResult {
  total: number;
  issues: Issue[];
}

export async function checkCommand(options: CliOptions, c3Dir: string, scope: Lite.Scope): Promise<void> {
  if (options.help) {
    showHelp("check");
    return;
  }

  const graph = await scope.resolve(graphAtom);

  const result: CheckResult = { total: graph.entities.size, issues: [] };

  const seenIds = new Set<string>();

  for (const entity of graph.entities.values()) {
    // Duplicate ID check
    if (seenIds.has(entity.id)) {
      result.issues.push({
        severity: "error",
        entity: entity.id,
        message: "duplicate ID",
      });
    }
    seenIds.add(entity.id);

    // Broken relationships: references to non-existent entities
    for (const relId of entity.relationships) {
      if (!graph.entities.has(relId)) {
        result.issues.push({
          severity: "error",
          entity: entity.id,
          message: `broken link to '${relId}'`,
        });
      }
    }

    // Missing parent container for components
    if (entity.type === "component") {
      const parentId = entity.frontmatter.parent;
      if (!parentId) {
        result.issues.push({
          severity: "error",
          entity: entity.id,
          message: "missing parent container",
        });
      } else if (!graph.entities.has(parentId)) {
        result.issues.push({
          severity: "error",
          entity: entity.id,
          message: `parent container '${parentId}' not found`,
        });
      }
    }

    // Empty content check
    if (!entity.body.trim()) {
      result.issues.push({
        severity: "warning",
        entity: entity.id,
        message: "empty content body",
      });
    }

    // ID/filename mismatch check
    const basename = path.basename(entity.path, ".md");
    if (basename !== "README") {
      if (!basename.startsWith(entity.id)) {
        result.issues.push({
          severity: "warning",
          entity: entity.id,
          message: `ID/filename mismatch: id='${entity.id}', file='${basename}.md'`,
        });
      }
    }
  }

  // Orphan check: entities with no incoming relationships (except context and containers)
  for (const entity of graph.entities.values()) {
    if (entity.type === "context" || entity.type === "container") continue;
    const incoming = graph.reverse(entity.id);
    if (incoming.length === 0 && !entity.frontmatter.parent) {
      result.issues.push({
        severity: "warning",
        entity: entity.id,
        message: "orphan: no incoming relationships",
      });
    }
  }

  // --- Output ---
  if (options.json) {
    console.log(renderJson(result));
    return;
  }

  const errors = result.issues.filter(i => i.severity === "error");
  const warnings = result.issues.filter(i => i.severity === "warning");

  if (errors.length === 0 && warnings.length === 0) {
    console.log(`✓ ${result.total} entities, 0 issues`);
  } else {
    console.log(`${result.total} entities, ${errors.length} errors, ${warnings.length} warnings`);
    for (const issue of result.issues) {
      const icon = issue.severity === "error" ? "✗" : "!";
      console.log(`  ${icon} ${issue.entity ?? "global"}: ${issue.message}`);
    }
  }
}
