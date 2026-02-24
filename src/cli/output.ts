// src/cli/output.ts
import type { C3Graph } from "../core/walker";

function truncate(s: string, max: number): string {
  return s.length > max ? s.slice(0, max - 3) + "..." : s;
}

export function renderTopology(graph: C3Graph): string {
  const lines: string[] = [];

  // Containers with their components
  const containers = graph.byType.get("container") || [];
  for (const container of containers.sort((a, b) => a.id.localeCompare(b.id))) {
    const containerGoal = container.frontmatter.goal ? ` — ${truncate(container.frontmatter.goal, 60)}` : "";
    lines.push(`${container.id}-${container.slug} (container)${containerGoal}`);
    const components = graph.children(container.id).sort((a, b) => a.id.localeCompare(b.id));
    for (let i = 0; i < components.length; i++) {
      const comp = components[i];
      const isLast = i === components.length - 1;
      const prefix = isLast ? "└── " : "├── ";
      const category = comp.frontmatter.category || (parseInt(comp.id.replace(/c3-\d+/, ""), 10) <= 9 ? "foundation" : "feature");
      const refs = graph.refsFor(comp.id).map(r => r.id).join(", ");
      const suffix = refs ? ` → ref: ${refs}` : "";
      const compGoal = comp.frontmatter.goal ? ` — ${truncate(comp.frontmatter.goal, 60)}` : "";
      lines.push(`${prefix}${comp.id}-${comp.slug} (${category})${compGoal}${suffix}`);
    }
    lines.push("");
  }

  // Cross-cutting refs
  const refs = graph.byType.get("ref") || [];
  if (refs.length > 0) {
    lines.push("Cross-cutting:");
    for (const ref of refs.sort((a, b) => a.id.localeCompare(b.id))) {
      const refGoal = ref.frontmatter.goal ? ` — ${truncate(ref.frontmatter.goal, 60)}` : "";
      const citers = graph.citedBy(ref.id).map(c => c.id).join(", ");
      lines.push(`  ${ref.id}${refGoal}${citers ? ` → used by: ${citers}` : ""}`);
    }
    lines.push("");
  }

  // ADRs
  const adrs = graph.byType.get("adr") || [];
  if (adrs.length > 0) {
    lines.push("ADRs:");
    for (const adr of adrs.sort((a, b) => a.id.localeCompare(b.id))) {
      const status = adr.frontmatter.status || "unknown";
      lines.push(`  ${adr.id}: ${adr.title} → status: ${status}`);
    }
    lines.push("");
  }

  return lines.join("\n").trim();
}

export function renderFlatList(graph: C3Graph): string {
  return [...graph.entities.values()]
    .sort((a, b) => a.path.localeCompare(b.path))
    .map(e => `${e.id}\t${e.type}\t${e.path}`)
    .join("\n");
}

export function renderJson(data: unknown): string {
  return JSON.stringify(data, null, 2);
}
