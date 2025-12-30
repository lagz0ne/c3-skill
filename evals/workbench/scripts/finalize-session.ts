#!/usr/bin/env bun
/**
 * Finalize eval session - generates summary from operation log
 */

import { readFile, writeFile } from "fs/promises";
import { existsSync } from "fs";

interface LogEntry {
  phase: string;
  tool: string;
  input: { file_path?: string };
}

async function main() {
  const outputDir = Bun.env.EVAL_OUTPUT_DIR ?? "/tmp/eval-run";
  const logPath = `${outputDir}/file-operations.jsonl`;
  const sessionPath = `${outputDir}/session.json`;

  // Parse operation log
  const stats: Record<string, number> = {};
  const filesRead = new Set<string>();
  const filesWritten = new Set<string>();

  if (existsSync(logPath)) {
    const content = await readFile(logPath, "utf-8");
    for (const line of content.trim().split("\n")) {
      if (!line) continue;
      const entry: LogEntry = JSON.parse(line);

      if (entry.phase === "post") {
        stats[entry.tool] = (stats[entry.tool] ?? 0) + 1;

        const filePath = entry.input.file_path;
        if (filePath) {
          if (entry.tool === "Read") filesRead.add(filePath);
          if (entry.tool === "Write" || entry.tool === "Edit") {
            filesWritten.add(filePath);
          }
        }
      }
    }
  }

  // Update session metadata
  let metadata: Record<string, unknown> = {};
  if (existsSync(sessionPath)) {
    metadata = JSON.parse(await readFile(sessionPath, "utf-8"));
  }

  metadata.ended_at = new Date().toISOString();
  metadata.stats = stats;
  metadata.files_read = [...filesRead].sort();
  metadata.files_written = [...filesWritten].sort();

  await writeFile(sessionPath, JSON.stringify(metadata, null, 2));

  console.error(`Eval complete: ${filesWritten.size} files written`);
  process.exit(0);
}

main().catch((e) => {
  console.error(`finalize-session error: ${e}`);
  process.exit(1);
});
