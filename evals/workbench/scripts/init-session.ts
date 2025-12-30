#!/usr/bin/env bun
/**
 * Initialize eval session - creates session metadata
 */

import { mkdir, writeFile } from "fs/promises";

interface SessionInput {
  session_id: string;
  cwd: string;
}

async function main() {
  const input: SessionInput = await Bun.stdin.json();

  const outputDir = Bun.env.EVAL_OUTPUT_DIR ?? "/tmp/eval-run";
  await mkdir(outputDir, { recursive: true });

  const metadata = {
    session_id: input.session_id,
    started_at: new Date().toISOString(),
    cwd: input.cwd,
    scenario: Bun.env.EVAL_SCENARIO ?? "unknown",
    codebase: Bun.env.EVAL_CODEBASE ?? "unknown",
  };

  await writeFile(
    `${outputDir}/session.json`,
    JSON.stringify(metadata, null, 2)
  );

  process.exit(0);
}

main().catch((e) => {
  console.error(`init-session error: ${e}`);
  process.exit(1);
});
