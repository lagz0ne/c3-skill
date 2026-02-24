// src/core/wiring.ts
import * as fs from "node:fs";

export function addComponentToContainerTable(containerReadmePath: string, componentId: string, name: string, category: string, goal: string): void {
  let content = fs.readFileSync(containerReadmePath, "utf-8");
  // Find the Components table and append a row
  const tablePattern = /(\| ID \| Name \| Category \| Status \| Goal Contribution \|[\s\S]*?)(\n\n|\n##|\n---|\Z)/;
  const match = content.match(tablePattern);
  if (match) {
    const newRow = `| ${componentId} | ${name} | ${category} | active | ${goal} |\n`;
    content = content.replace(match[0], match[1] + newRow + match[2]);
    fs.writeFileSync(containerReadmePath, content, "utf-8");
  }
}

