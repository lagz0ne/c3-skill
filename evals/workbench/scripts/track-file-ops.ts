#!/usr/bin/env bun
/**
 * Track file operations to JSONL log
 * Usage: bun track-file-ops.ts [pre|post]
 */

import { appendFile, mkdir } from "fs/promises";
import { dirname } from "path";

interface HookInput {
  session_id: string;
  tool_name: string;
  tool_input: Record<string, unknown>;
  tool_response?: Record<string, unknown>;
  cwd: string;
}

const phase = Bun.argv[2] ?? "unknown";

async function main() {
  const input: HookInput = await Bun.stdin.json();

  const entry = {
    timestamp: new Date().toISOString(),
    phase,
    tool: input.tool_name,
    input: input.tool_input,
    cwd: input.cwd,
    ...(phase === "post" && { response: input.tool_response }),
  };

  const outputDir = Bun.env.EVAL_OUTPUT_DIR ?? "/tmp/eval-run";
  const logPath = `${outputDir}/file-operations.jsonl`;

  await mkdir(dirname(logPath), { recursive: true });
  await appendFile(logPath, JSON.stringify(entry) + "\n");

  process.exit(0);
}

main().catch((e) => {
  console.error(`track-file-ops error: ${e}`);
  process.exit(1); // Non-blocking error
});
