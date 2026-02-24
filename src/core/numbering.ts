// src/core/numbering.ts
import type { C3Graph } from "./walker";

export function nextContainerId(graph: C3Graph): number {
  const containers = graph.byType.get("container") || [];
  const nums = containers
    .map(c => parseInt(c.id.replace("c3-", ""), 10))
    .filter(n => !isNaN(n));
  return nums.length === 0 ? 1 : Math.max(...nums) + 1;
}

export function nextComponentId(graph: C3Graph, containerNum: number, feature: boolean): string {
  const prefix = `c3-${containerNum}`;
  const components = (graph.byType.get("component") || [])
    .filter(c => c.id.startsWith(prefix))
    .map(c => parseInt(c.id.replace(prefix, ""), 10))
    .filter(n => !isNaN(n));

  if (feature) {
    // Feature: 10+
    const featureNums = components.filter(n => n >= 10);
    const next = featureNums.length === 0 ? 10 : Math.max(...featureNums) + 1;
    return `c3-${containerNum}${String(next).padStart(2, "0")}`;
  } else {
    // Foundation: 01-09
    const foundationNums = components.filter(n => n >= 1 && n <= 9);
    const next = foundationNums.length === 0 ? 1 : Math.max(...foundationNums) + 1;
    if (next > 9) throw new Error(`Container c3-${containerNum} has no more foundation slots (01-09 full)`);
    return `c3-${containerNum}${String(next).padStart(2, "0")}`;
  }
}

export function nextAdrId(slug: string): string {
  const date = new Date().toISOString().slice(0, 10).replace(/-/g, "");
  return `adr-${date}-${slug}`;
}
