import * as fs from "node:fs";
import * as path from "node:path";
import type { CliOptions } from "../context";
import { graphAtom } from "../context";
import type { Lite } from "@pumped-fn/lite";
import { showHelp } from "../help";
import type { C3Entity } from "../../core/walker";

export async function syncCommand(options: CliOptions, c3Dir: string, scope: Lite.Scope): Promise<void> {
  if (options.help) {
    showHelp("sync");
    return;
  }

  const graph = await scope.resolve(graphAtom);
  const components = graph.byType.get("component") || [];

  // Find project root (parent of .c3/)
  const projectRoot = path.dirname(c3Dir);
  const skillsDir = path.join(projectRoot, ".claude", "skills");

  // Remove existing guard skills
  if (fs.existsSync(skillsDir)) {
    const existing = fs.readdirSync(skillsDir);
    for (const file of existing) {
      if (file.startsWith("c3-guard-")) {
        fs.unlinkSync(path.join(skillsDir, file));
      }
    }
  }

  let generated = 0;
  const generatedFiles: string[] = [];

  for (const comp of components) {
    const codeRefs = parseCodeReferences(comp.body);
    if (codeRefs.length === 0) continue;

    const skillContent = buildGuardSkill(comp, codeRefs, graph);

    fs.mkdirSync(skillsDir, { recursive: true });
    const fileName = `c3-guard-${comp.slug}.md`;
    fs.writeFileSync(path.join(skillsDir, fileName), skillContent, "utf-8");
    generatedFiles.push(fileName);
    generated++;
  }

  console.log(`Generated ${generated} guard skills in .claude/skills/`);
  for (const file of generatedFiles) {
    const comp = components.find(c => `c3-guard-${c.slug}.md` === file);
    const patterns = comp ? parseCodeReferences(comp.body).map(r => r.file).join(", ") : "";
    console.log(`  ${file}${patterns ? `\t(${patterns})` : ""}`);
  }
}

interface CodeRef {
  file: string;
  purpose: string;
}

function parseCodeReferences(body: string): CodeRef[] {
  const refs: CodeRef[] = [];

  // Find ## Code References section
  const sectionMatch = body.match(/## Code References\s*\n([\s\S]*?)(?=\n## |\n---|\Z|$)/);
  if (!sectionMatch) return refs;

  const section = sectionMatch[1];

  // Parse table rows: | File | Purpose |
  const rows = section.split("\n").filter(line => {
    const trimmed = line.trim();
    // Skip header row, separator row, and empty lines
    return trimmed.startsWith("|") &&
           !trimmed.includes("---") &&
           !trimmed.toLowerCase().includes("| file");
  });

  for (const row of rows) {
    const cells = row.split("|").map(c => c.trim()).filter(c => c.length > 0);
    if (cells.length >= 2) {
      const file = cells[0].replace(/`/g, "");
      const purpose = cells[1];
      if (file && file !== "File") {
        refs.push({ file, purpose });
      }
    }
  }

  return refs;
}

function buildGuardSkill(
  comp: C3Entity,
  codeRefs: CodeRef[],
  graph: ReturnType<typeof import("../../core/walker").buildRelationshipGraph>,
): string {
  const filePatterns = codeRefs.map(r => r.file).join(", ");

  // Find parent container
  const parentId = comp.frontmatter.parent;
  const parent = parentId ? graph.entities.get(parentId) : undefined;
  const containerName = parent ? `${parent.title}` : "unknown";
  const containerSlug = parent ? `${parent.id}-${parent.slug}` : "unknown";
  const category = comp.frontmatter.category || "component";

  // Build rules from refs
  const compRefs = graph.refsFor(comp.id);
  const rulesLines = compRefs.map(ref => {
    const goal = ref.frontmatter.goal || ref.title;
    return `- **${ref.id}**: ${goal}`;
  });

  // Build blast radius
  const affected = graph.forward(comp.id);
  const blastLines = affected
    .filter(e => e.id !== comp.id)
    .map(e => `- ${e.id}-${e.slug} (${e.type})`);

  const lines: string[] = [];
  lines.push("---");
  lines.push(`name: c3-guard-${comp.slug}`);
  lines.push(`description: >`);
  lines.push(`  Use when modifying files in ${filePatterns}.`);
  lines.push(`  Guards ${comp.title} (${comp.id}) in the ${containerName} container.`);
  lines.push("---");
  lines.push("");
  lines.push(`## Component: ${comp.id}-${comp.slug} (${category})`);
  lines.push(`Container: ${containerSlug}`);
  lines.push("");
  lines.push("## Rules");
  if (rulesLines.length > 0) {
    lines.push(...rulesLines);
  } else {
    lines.push("No refs cited.");
  }
  lines.push("");
  lines.push("## Blast radius");
  if (blastLines.length > 0) {
    lines.push(...blastLines);
  } else {
    lines.push("No downstream dependencies.");
  }
  lines.push("");
  lines.push("## After changes");
  lines.push("Invoke: c3-audit skill");
  lines.push("");

  return lines.join("\n");
}
