#!/usr/bin/env bun
/**
 * Living Entity Generator
 *
 * Reads a .c3/ directory and generates Claude Code agent .md files
 * for each entity (container, component) plus an orchestrator.
 *
 * Usage: bun experiments/living-entity/generate.ts /path/to/.c3/
 */

import { readdir, readFile, writeFile, mkdir } from "fs/promises";
import { join, basename, relative } from "path";

// --- Types ---

interface Frontmatter {
  id: string;
  title: string;
  type?: string; // container, component
  category?: string; // foundation, feature
  parent?: string;
  summary?: string;
  status?: string;
}

interface C3Entity {
  frontmatter: Frontmatter;
  filePath: string;
  body: string;
  sections: Record<string, string>;
}

interface Ref {
  id: string;
  title: string;
  filePath: string;
  body: string;
}

interface ADR {
  id: string;
  title: string;
  status: string;
  filePath: string;
}

interface Topology {
  context: C3Entity | null;
  containers: C3Entity[];
  components: C3Entity[];
  refs: Ref[];
  adrs: ADR[];
}

// --- Parsing ---

function parseFrontmatter(content: string): { frontmatter: Frontmatter; body: string } {
  const match = content.match(/^---\n([\s\S]*?)\n---\n([\s\S]*)$/);
  if (!match) {
    return { frontmatter: { id: "", title: "" }, body: content };
  }

  const yaml = match[1];
  const body = match[2];
  const fm: Record<string, string> = {};

  for (const line of yaml.split("\n")) {
    const kv = line.match(/^(\w[\w-]*):\s*(.+)$/);
    if (kv) {
      fm[kv[1]] = kv[2].trim();
    }
  }

  return {
    frontmatter: fm as unknown as Frontmatter,
    body,
  };
}

function extractSections(body: string): Record<string, string> {
  const sections: Record<string, string> = {};
  const lines = body.split("\n");
  let currentHeading = "";
  let currentContent: string[] = [];

  for (const line of lines) {
    const h2Match = line.match(/^## (.+)$/);
    if (h2Match) {
      if (currentHeading) {
        sections[currentHeading] = currentContent.join("\n").trim();
      }
      currentHeading = h2Match[1];
      currentContent = [];
    } else {
      currentContent.push(line);
    }
  }

  if (currentHeading) {
    sections[currentHeading] = currentContent.join("\n").trim();
  }

  return sections;
}

function extractCodePaths(body: string): string[] {
  const paths: string[] = [];

  // Match file paths in backticks like `src/server/flows/*.ts`
  const tickMatches = body.matchAll(/`((?:src|apps|packages)\/[^\s`]+)`/g);
  for (const m of tickMatches) {
    paths.push(m[1]);
  }

  // Match file paths in "Reference" sections
  const refSection = body.match(/## (?:Reference|References)\n([\s\S]*?)(?=\n## |\n$)/);
  if (refSection) {
    const refPaths = refSection[1].matchAll(/`((?:src|apps|packages)\/[^\s`]+)`/g);
    for (const m of refPaths) {
      if (!paths.includes(m[1])) paths.push(m[1]);
    }
    // Also match - prefixed paths
    const dashPaths = refSection[1].matchAll(/- ((?:src|apps|packages)\/\S+)/g);
    for (const m of dashPaths) {
      if (!paths.includes(m[1])) paths.push(m[1]);
    }
  }

  return [...new Set(paths)];
}

function extractRefLinks(body: string): string[] {
  const refs: string[] = [];
  const matches = body.matchAll(/\[ref-[\w-]+\]\([^)]+\)/g);
  for (const m of matches) {
    const id = m[0].match(/\[(ref-[\w-]+)\]/);
    if (id) refs.push(id[1]);
  }
  // Also from "Uses" tables
  const usesMatches = body.matchAll(/ref-[\w-]+/g);
  for (const m of usesMatches) {
    if (!refs.includes(m[0])) refs.push(m[0]);
  }
  return [...new Set(refs)];
}

function extractUsedComponents(body: string): string[] {
  const ids: string[] = [];
  const matches = body.matchAll(/c3-(\d+)/g);
  for (const m of matches) {
    ids.push(`c3-${m[1]}`);
  }
  return [...new Set(ids)];
}

// --- Discovery ---

async function discoverC3Dir(c3Dir: string): Promise<Topology> {
  const topology: Topology = {
    context: null,
    containers: [],
    components: [],
    refs: [],
    adrs: [],
  };

  // Read top-level README (context)
  try {
    const contextContent = await readFile(join(c3Dir, "README.md"), "utf-8");
    const { frontmatter, body } = parseFrontmatter(contextContent);
    topology.context = {
      frontmatter,
      filePath: join(c3Dir, "README.md"),
      body,
      sections: extractSections(body),
    };
  } catch {}

  // Read refs
  const refDir = join(c3Dir, "ref");
  try {
    const refFiles = await readdir(refDir);
    for (const f of refFiles) {
      if (!f.endsWith(".md")) continue;
      const content = await readFile(join(refDir, f), "utf-8");
      const { frontmatter, body } = parseFrontmatter(content);
      topology.refs.push({
        id: frontmatter.id || basename(f, ".md"),
        title: frontmatter.title || basename(f, ".md"),
        filePath: join(refDir, f),
        body,
      });
    }
  } catch {}

  // Read ADRs
  const adrDir = join(c3Dir, "adr");
  try {
    const adrFiles = await readdir(adrDir);
    for (const f of adrFiles) {
      if (!f.endsWith(".md")) continue;
      const content = await readFile(join(adrDir, f), "utf-8");
      const { frontmatter } = parseFrontmatter(content);
      topology.adrs.push({
        id: frontmatter.id || basename(f, ".md"),
        title: frontmatter.title || basename(f, ".md"),
        status: frontmatter.status || "unknown",
        filePath: join(adrDir, f),
      });
    }
  } catch {}

  // Read container directories
  const entries = await readdir(c3Dir, { withFileTypes: true });
  for (const entry of entries) {
    if (!entry.isDirectory() || entry.name === "ref" || entry.name === "adr") continue;

    const containerDir = join(c3Dir, entry.name);
    const containerReadme = join(containerDir, "README.md");

    try {
      const content = await readFile(containerReadme, "utf-8");
      const { frontmatter, body } = parseFrontmatter(content);
      if (frontmatter.id) {
        topology.containers.push({
          frontmatter: { ...frontmatter, type: frontmatter.type || "container" },
          filePath: containerReadme,
          body,
          sections: extractSections(body),
        });
      }
    } catch {
      continue;
    }

    // Read components in this container
    const componentFiles = await readdir(containerDir);
    for (const cf of componentFiles) {
      if (cf === "README.md" || !cf.endsWith(".md")) continue;
      const content = await readFile(join(containerDir, cf), "utf-8");
      const { frontmatter, body } = parseFrontmatter(content);
      if (frontmatter.id) {
        topology.components.push({
          frontmatter: { ...frontmatter, type: frontmatter.type || "component" },
          filePath: join(containerDir, cf),
          body,
          sections: extractSections(body),
        });
      }
    }
  }

  return topology;
}

// --- Agent Generation ---

function slugify(id: string, title: string): string {
  return `${id}-${title.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/-+$/, "")}`;
}

function buildDescription(entity: C3Entity, topology: Topology): string {
  const fm = entity.frontmatter;
  const parent = topology.containers.find((c) => c.frontmatter.id === fm.parent);
  const parentName = parent?.frontmatter.title || fm.parent || "system";
  const codePaths = extractCodePaths(entity.body);
  const codeHint = codePaths.length > 0 ? ` Code: ${codePaths.slice(0, 3).join(", ")}.` : "";

  // Must be under 1024 chars
  let desc = `I am ${fm.title} (${fm.id}), a ${fm.category || fm.type} ${fm.type} of ${parentName}. ${fm.summary || ""}.${codeHint} Route to me for changes affecting my domain.`;

  if (desc.length > 1024) {
    desc = desc.slice(0, 1020) + "...";
  }

  return desc;
}

function buildEntityAgentContent(entity: C3Entity, topology: Topology): string {
  const fm = entity.frontmatter;
  const parent = topology.containers.find((c) => c.frontmatter.id === fm.parent);
  const codePaths = extractCodePaths(entity.body);
  const refLinks = extractRefLinks(entity.body);
  const usedComponents = extractUsedComponents(entity.body).filter((id) => id !== fm.id);

  // Find applicable refs
  const applicableRefs = topology.refs.filter((r) => refLinks.includes(r.id));

  // Find ADRs that mention this entity
  const relatedAdrs = topology.adrs.filter((a) => {
    // Simple heuristic - ADR title or ID mentions this component's domain
    return false; // Will enhance later with content matching
  });

  // Find who depends on me
  const dependents = topology.components.filter((c) => {
    return c.frontmatter.id !== fm.id && extractUsedComponents(c.body).includes(fm.id);
  });

  const lines: string[] = [];

  // Frontmatter
  lines.push("---");
  lines.push(`name: ${slugify(fm.id, fm.title)}`);
  lines.push(`description: |`);
  lines.push(`  ${buildDescription(entity, topology)}`);
  lines.push("tools:");
  lines.push("  - Read");
  lines.push("  - Glob");
  lines.push("  - Grep");
  lines.push("---");
  lines.push("");

  // System prompt
  lines.push(`# You are ${fm.title} (${fm.id})`);
  lines.push("");

  // Identity
  lines.push("## Identity");
  lines.push("");
  lines.push(`- **ID**: ${fm.id}`);
  lines.push(`- **Type**: ${fm.type}`);
  if (fm.category) lines.push(`- **Category**: ${fm.category}`);
  if (parent) lines.push(`- **Container**: ${parent.frontmatter.title} (${parent.frontmatter.id})`);
  lines.push(`- **Summary**: ${fm.summary || "N/A"}`);
  lines.push("");

  // Goal
  if (entity.sections["Goal"]) {
    lines.push("## Goal");
    lines.push("");
    lines.push(entity.sections["Goal"]);
    lines.push("");
  }

  // Code Ownership
  if (codePaths.length > 0) {
    lines.push("## Code Ownership");
    lines.push("");
    lines.push("You own and protect these code paths:");
    lines.push("");
    for (const p of codePaths) {
      lines.push(`- \`${p}\``);
    }
    lines.push("");
  }

  // Conventions (the behavioral rules)
  if (entity.sections["Conventions"] || entity.sections["Convention"]) {
    lines.push("## Conventions (MUST ENFORCE)");
    lines.push("");
    lines.push(entity.sections["Conventions"] || entity.sections["Convention"]);
    lines.push("");
  }

  // Contract
  if (entity.sections["Contract"]) {
    lines.push("## Contract");
    lines.push("");
    lines.push(entity.sections["Contract"]);
    lines.push("");
  }

  // Structure / Pattern (important for how code should be written)
  if (entity.sections["Structure"]) {
    lines.push("## Code Structure");
    lines.push("");
    lines.push(entity.sections["Structure"]);
    lines.push("");
  }

  if (entity.sections["Pattern"]) {
    lines.push("## Code Pattern");
    lines.push("");
    lines.push(entity.sections["Pattern"]);
    lines.push("");
  }

  // Flows (for feature components)
  if (entity.sections["Flows"]) {
    lines.push("## Flows");
    lines.push("");
    lines.push(entity.sections["Flows"]);
    lines.push("");
  }

  // State Machine
  if (entity.sections["State Machine"]) {
    lines.push("## State Machine");
    lines.push("");
    lines.push(entity.sections["State Machine"]);
    lines.push("");
  }

  // Edge Cases
  if (entity.sections["Edge Cases"]) {
    lines.push("## Edge Cases");
    lines.push("");
    lines.push(entity.sections["Edge Cases"]);
    lines.push("");
  }

  // Applicable References (inlined)
  if (applicableRefs.length > 0) {
    lines.push("## Applicable References (MUST FOLLOW)");
    lines.push("");
    for (const ref of applicableRefs) {
      lines.push(`### ${ref.title} (${ref.id})`);
      lines.push("");
      lines.push(ref.body.trim());
      lines.push("");
    }
  }

  // Dependencies
  if (usedComponents.length > 0) {
    lines.push("## Dependencies (components I use)");
    lines.push("");
    for (const id of usedComponents) {
      const dep = topology.components.find((c) => c.frontmatter.id === id);
      if (dep) {
        lines.push(`- **${dep.frontmatter.title}** (${id}): ${dep.frontmatter.summary || ""}`);
      }
    }
    lines.push("");
  }

  // Dependents (who uses me)
  if (dependents.length > 0) {
    lines.push("## Dependents (components that use me)");
    lines.push("");
    for (const dep of dependents) {
      lines.push(
        `- **${dep.frontmatter.title}** (${dep.frontmatter.id}): ${dep.frontmatter.summary || ""}`
      );
    }
    lines.push("");
  }

  // Advisory instructions
  lines.push("## Your Role");
  lines.push("");
  lines.push("You are a **living entity agent**. When consulted about a proposed change:");
  lines.push("");
  lines.push("1. **Check code ownership**: Does this change target files you own? Use Read/Glob/Grep to verify.");
  lines.push("2. **Enforce conventions**: Does the proposed change follow your conventions? Flag violations.");
  lines.push("3. **Check references**: Are all applicable reference patterns being followed?");
  lines.push("4. **Assess relationships**: Which of your dependencies or dependents would be affected?");
  lines.push("5. **Advise**: Provide structured guidance including:");
  lines.push("   - What files need to change");
  lines.push("   - Which conventions/refs apply");
  lines.push("   - Which other entities should be consulted");
  lines.push("   - Potential risks or edge cases");
  lines.push("");
  lines.push("You ADVISE only. You do not make changes. Be specific and cite your constraints.");

  return lines.join("\n");
}

function buildOrchestratorContent(topology: Topology): string {
  const lines: string[] = [];

  // Build compact topology description
  const contextName = topology.context?.frontmatter.title || "System";
  const contextSummary = topology.context?.frontmatter.summary || "";

  // Description must be under 1024 chars
  const desc = `Living Entity Orchestrator for "${contextName}". Routes architecture change requests to the correct entity-agent(s). Knows the full topology: ${topology.containers.length} containers, ${topology.components.length} components, ${topology.refs.length} refs. Use me when proposing changes to the system.`;

  lines.push("---");
  lines.push("name: living-entity-orchestrator");
  lines.push("description: |");
  lines.push(`  ${desc}`);
  lines.push("tools:");
  lines.push("  - Task");
  lines.push("  - Read");
  lines.push("  - Glob");
  lines.push("  - Grep");
  lines.push("---");
  lines.push("");

  lines.push(`# Living Entity Orchestrator: ${contextName}`);
  lines.push("");
  lines.push(`> ${contextSummary}`);
  lines.push("");

  // Topology Graph
  lines.push("## Topology");
  lines.push("");

  for (const container of topology.containers) {
    const components = topology.components.filter(
      (c) => c.frontmatter.parent === container.frontmatter.id
    );
    lines.push(
      `### ${container.frontmatter.title} (${container.frontmatter.id})`
    );
    lines.push(`> ${container.frontmatter.summary || ""}`);
    lines.push("");

    if (components.length > 0) {
      lines.push("| ID | Name | Category | Agent | Code Paths |");
      lines.push("|----|------|----------|-------|------------|");
      for (const comp of components) {
        const agent = slugify(comp.frontmatter.id, comp.frontmatter.title);
        const codePaths = extractCodePaths(comp.body).slice(0, 2).join(", ");
        lines.push(
          `| ${comp.frontmatter.id} | ${comp.frontmatter.title} | ${comp.frontmatter.category || "-"} | ${agent} | ${codePaths || "-"} |`
        );
      }
      lines.push("");
    }
  }

  // Ref inventory
  lines.push("## Cross-Cutting References");
  lines.push("");
  lines.push("| Ref | Title | Affects |");
  lines.push("|-----|-------|---------|");
  for (const ref of topology.refs) {
    // Find which components reference this ref
    const affected = topology.components
      .filter((c) => extractRefLinks(c.body).includes(ref.id))
      .map((c) => c.frontmatter.id);
    lines.push(`| ${ref.id} | ${ref.title} | ${affected.join(", ") || "all"} |`);
  }
  lines.push("");

  // ADR inventory
  if (topology.adrs.length > 0) {
    lines.push("## Architecture Decisions");
    lines.push("");
    lines.push("| ADR | Title | Status |");
    lines.push("|-----|-------|--------|");
    for (const adr of topology.adrs) {
      lines.push(`| ${adr.id} | ${adr.title} | ${adr.status} |`);
    }
    lines.push("");
  }

  // Entity agent index
  lines.push("## Entity Agent Index");
  lines.push("");
  lines.push("Use `Task` with `subagent_type` matching the agent name to delegate:");
  lines.push("");
  for (const container of topology.containers) {
    const agent = slugify(container.frontmatter.id, container.frontmatter.title);
    lines.push(`- **${agent}**: ${container.frontmatter.title}`);
  }
  for (const comp of topology.components) {
    const agent = slugify(comp.frontmatter.id, comp.frontmatter.title);
    lines.push(`- **${agent}**: ${comp.frontmatter.title}`);
  }
  lines.push("");

  // Routing instructions
  lines.push("## Routing Protocol");
  lines.push("");
  lines.push("When you receive a change request:");
  lines.push("");
  lines.push("1. **Parse the request**: What domain does this change affect?");
  lines.push("2. **Map to entities**: Using the topology above, identify which component(s) own the affected code.");
  lines.push("3. **Check refs**: Which cross-cutting references apply to this change?");
  lines.push("4. **Delegate**: Use the Task tool to consult each affected entity agent.");
  lines.push("   - Send the change description and ask for an impact assessment.");
  lines.push("   - If multiple entities are affected, consult them in parallel.");
  lines.push("5. **Synthesize**: Collect all advisories and present a unified assessment:");
  lines.push("   - Which entities are affected and why");
  lines.push("   - Which constraints must be followed");
  lines.push("   - Potential risks flagged by entity agents");
  lines.push("   - Recommended approach");
  lines.push("");
  lines.push("## Delegation Example");
  lines.push("");
  lines.push('```');
  lines.push('Task(');
  lines.push('  subagent_type: "c3-205-pr-flows",');
  lines.push('  prompt: "Assess impact of adding retry logic to payment request approval. What constraints apply? What other entities would be affected?"');
  lines.push(")");
  lines.push("```");

  return lines.join("\n");
}

// --- Main ---

async function main() {
  const c3Dir = process.argv[2];
  if (!c3Dir) {
    console.error("Usage: bun experiments/living-entity/generate.ts /path/to/.c3/");
    process.exit(1);
  }

  const outputDir = join(import.meta.dir, "generated", "agents");
  await mkdir(outputDir, { recursive: true });

  console.log(`Reading C3 graph from: ${c3Dir}`);
  const topology = await discoverC3Dir(c3Dir);

  console.log(`Found:`);
  console.log(`  Context: ${topology.context?.frontmatter.title || "none"}`);
  console.log(`  Containers: ${topology.containers.length}`);
  console.log(`  Components: ${topology.components.length}`);
  console.log(`  Refs: ${topology.refs.length}`);
  console.log(`  ADRs: ${topology.adrs.length}`);
  console.log("");

  let generated = 0;

  // Generate container agents
  for (const container of topology.containers) {
    const slug = slugify(container.frontmatter.id, container.frontmatter.title);
    const content = buildEntityAgentContent(container, topology);
    const outPath = join(outputDir, `${slug}.md`);
    await writeFile(outPath, content);
    console.log(`  Generated: ${slug}.md`);
    generated++;
  }

  // Generate component agents
  for (const component of topology.components) {
    const slug = slugify(component.frontmatter.id, component.frontmatter.title);
    const content = buildEntityAgentContent(component, topology);
    const outPath = join(outputDir, `${slug}.md`);
    await writeFile(outPath, content);
    console.log(`  Generated: ${slug}.md`);
    generated++;
  }

  // Generate orchestrator
  const orchestratorContent = buildOrchestratorContent(topology);
  const orchestratorPath = join(outputDir, "living-entity-orchestrator.md");
  await writeFile(orchestratorPath, orchestratorContent);
  console.log(`  Generated: living-entity-orchestrator.md`);
  generated++;

  console.log(`\nDone! Generated ${generated} agent files in ${outputDir}`);
}

main().catch(console.error);
