#!/usr/bin/env bun
/**
 * Simple YAML parser CLI
 * Usage: bun parse-yaml.ts <file> <path>
 * Example: bun parse-yaml.ts scenario.yaml .prompt
 */

import { readFile } from "fs/promises";
import { parse as parseYaml } from "yaml";

const file = Bun.argv[2];
const path = Bun.argv[3];

if (!file) {
  console.error("Usage: bun parse-yaml.ts <file> <path>");
  process.exit(1);
}

async function main() {
  const content = await readFile(file, "utf-8");
  const data = parseYaml(content);

  if (!path) {
    console.log(JSON.stringify(data, null, 2));
    return;
  }

  // Navigate path like ".prompt" or ".setup[]"
  const parts = path.replace(/^\./, "").split(".");
  let value: any = data;

  for (const part of parts) {
    if (part.endsWith("[]")) {
      // Array access - return each element on a line
      const key = part.slice(0, -2);
      value = value?.[key];
      if (Array.isArray(value)) {
        for (const item of value) {
          console.log(item);
        }
        return;
      }
    } else {
      value = value?.[part];
    }
  }

  if (value !== undefined && value !== null) {
    console.log(typeof value === "string" ? value : JSON.stringify(value));
  }
}

main().catch((e) => {
  console.error(`Error: ${e}`);
  process.exit(1);
});
