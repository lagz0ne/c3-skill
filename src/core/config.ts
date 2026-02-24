// src/core/config.ts
import * as fs from "node:fs";
import * as path from "node:path";
import { parse as parseYaml } from "yaml";
import { z } from "zod";

const configSchema = z.object({
  // Minimal config — embedding removed, add future CLI config here
}).default({});

export type C3Config = z.infer<typeof configSchema>;

const DEFAULT_CONFIG: C3Config = configSchema.parse({});

export function loadConfig(c3Dir: string): C3Config {
  const configPath = path.join(c3Dir, "config.yaml");
  if (!fs.existsSync(configPath)) return DEFAULT_CONFIG;
  const raw = fs.readFileSync(configPath, "utf-8");
  const parsed = parseYaml(raw);
  const result = configSchema.safeParse(parsed);
  return result.success ? result.data : DEFAULT_CONFIG;
}

export function findC3Dir(startDir: string): string | null {
  let dir = path.resolve(startDir);
  while (true) {
    const candidate = path.join(dir, ".c3");
    if (fs.existsSync(candidate) && fs.statSync(candidate).isDirectory()) {
      return candidate;
    }
    const parent = path.dirname(dir);
    if (parent === dir) return null;
    dir = parent;
  }
}
