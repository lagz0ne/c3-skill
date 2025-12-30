#!/usr/bin/env bun
/**
 * Rubric evaluator - uses LLM to assess quality of C3 output
 * Usage: bun evaluate.ts <output-dir> [rubric.yaml]
 */

import { readFile, readdir } from "fs/promises";
import { existsSync } from "fs";
import { join } from "path";
import { parse as parseYaml } from "yaml";

interface RubricCriterion {
  name: string;
  weight: number;
  description: string;
  scoring: string;
}

interface Rubric {
  name: string;
  description?: string;
  criteria: RubricCriterion[];
}

interface EvalResult {
  overall_score: number;
  max_score: number;
  percentage: number;
  criteria: {
    name: string;
    score: number;
    max_score: number;
    feedback: string;
  }[];
  summary: string;
  improvements: string[];
}

const outputDir = Bun.argv[2];
const rubricPath = Bun.argv[3] ?? join(import.meta.dir, "../rubric.yaml");

if (!outputDir) {
  console.error("Usage: bun evaluate.ts <output-dir> [rubric.yaml]");
  process.exit(1);
}

async function collectC3Output(dir: string): Promise<string> {
  const c3Dir = join(dir, "c3-output");
  if (!existsSync(c3Dir)) {
    return "No .c3/ output found";
  }

  const files: string[] = [];

  async function walk(currentDir: string, prefix = "") {
    const entries = await readdir(currentDir, { withFileTypes: true });
    for (const entry of entries) {
      const path = join(currentDir, entry.name);
      const relativePath = prefix ? `${prefix}/${entry.name}` : entry.name;

      if (entry.isDirectory()) {
        await walk(path, relativePath);
      } else if (entry.name.endsWith(".md")) {
        const content = await readFile(path, "utf-8");
        files.push(`=== ${relativePath} ===\n${content}\n`);
      }
    }
  }

  await walk(c3Dir);
  return files.join("\n");
}

async function collectSessionInfo(dir: string): Promise<string> {
  const sessionPath = join(dir, "session.json");
  if (!existsSync(sessionPath)) {
    return "No session info";
  }
  const session = JSON.parse(await readFile(sessionPath, "utf-8"));
  return JSON.stringify(session, null, 2);
}

async function collectFileOps(dir: string): Promise<string> {
  const opsPath = join(dir, "file-operations.jsonl");
  if (!existsSync(opsPath)) {
    return "No file operations logged";
  }
  return await readFile(opsPath, "utf-8");
}

function buildPrompt(rubric: Rubric, c3Output: string, sessionInfo: string): string {
  const criteriaDesc = rubric.criteria
    .map((c, i) => `${i + 1}. **${c.name}** (weight: ${c.weight})\n   ${c.description}\n   Scoring: ${c.scoring}`)
    .join("\n\n");

  return `You are an expert evaluator for C3 architecture documentation.

## Task

Evaluate the following C3 documentation output against the rubric criteria.

## Rubric: ${rubric.name}

${rubric.description ?? ""}

### Criteria

${criteriaDesc}

## Session Info

\`\`\`json
${sessionInfo}
\`\`\`

## C3 Output to Evaluate

${c3Output}

## Instructions

For each criterion:
1. Assess the output against the scoring guide
2. Assign a score from 0-10
3. Provide specific feedback

Then provide:
- Overall summary
- Top 3 specific improvements

## Response Format

Respond with valid JSON only:

\`\`\`json
{
  "criteria": [
    {
      "name": "criterion name",
      "score": 8,
      "max_score": 10,
      "feedback": "specific feedback"
    }
  ],
  "summary": "overall assessment",
  "improvements": ["improvement 1", "improvement 2", "improvement 3"]
}
\`\`\``;
}

async function callClaude(prompt: string): Promise<string> {
  // Use claude CLI in print mode for evaluation
  const proc = Bun.spawn(
    [
      "claude",
      "-p",
      "--output-format",
      "text",
      "--no-session-persistence",
      "--model",
      "sonnet",
      "--max-budget-usd",
      "0.10",
      prompt,
    ],
    {
      stdout: "pipe",
      stderr: "pipe",
    }
  );

  const output = await new Response(proc.stdout).text();
  const exitCode = await proc.exited;

  if (exitCode !== 0) {
    const stderr = await new Response(proc.stderr).text();
    throw new Error(`Claude exited with ${exitCode}: ${stderr}`);
  }

  return output;
}

function parseResponse(response: string, rubric: Rubric): EvalResult {
  // Extract JSON from response (might be wrapped in markdown)
  const jsonMatch = response.match(/```json\n?([\s\S]*?)\n?```/) ||
                    response.match(/\{[\s\S]*\}/);

  if (!jsonMatch) {
    throw new Error("Could not parse JSON from response");
  }

  const jsonStr = jsonMatch[1] ?? jsonMatch[0];
  const parsed = JSON.parse(jsonStr);

  // Calculate weighted score
  let totalWeight = 0;
  let weightedScore = 0;

  const criteriaResults = parsed.criteria.map((c: any) => {
    const rubricCriterion = rubric.criteria.find((rc) => rc.name === c.name);
    const weight = rubricCriterion?.weight ?? 1;
    totalWeight += weight;
    weightedScore += (c.score / c.max_score) * weight;

    return {
      name: c.name,
      score: c.score,
      max_score: c.max_score,
      feedback: c.feedback,
    };
  });

  const percentage = totalWeight > 0 ? (weightedScore / totalWeight) * 100 : 0;

  return {
    overall_score: weightedScore,
    max_score: totalWeight,
    percentage: Math.round(percentage * 10) / 10,
    criteria: criteriaResults,
    summary: parsed.summary,
    improvements: parsed.improvements ?? [],
  };
}

async function main() {
  // Load rubric
  if (!existsSync(rubricPath)) {
    console.error(`Rubric not found: ${rubricPath}`);
    process.exit(1);
  }

  const rubricContent = await readFile(rubricPath, "utf-8");
  const rubric: Rubric = parseYaml(rubricContent);

  // Collect output
  const c3Output = await collectC3Output(outputDir);
  const sessionInfo = await collectSessionInfo(outputDir);

  // Build and send prompt
  const prompt = buildPrompt(rubric, c3Output, sessionInfo);

  console.error("Calling evaluator...");
  const response = await callClaude(prompt);

  // Parse and output results
  const result = parseResponse(response, rubric);
  console.log(JSON.stringify(result, null, 2));
}

main().catch((e) => {
  console.error(`Evaluate error: ${e}`);
  process.exit(1);
});
