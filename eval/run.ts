#!/usr/bin/env bun
/**
 * Eval test runner - uses Claude Agent SDK
 *
 * Usage:
 *   bun eval/run.ts cases/onboard-simple.yaml
 *   bun eval/run.ts cases/  # run all
 *   bun eval/run.ts cases/onboard-simple.yaml --verbose
 */

import { parseArgs } from "util";
import { join, dirname } from "path";
import { parse as parseYaml } from "yaml";
import { $ } from "bun";
import { query, type HookInput, type HookJSONOutput, type PreToolUseHookInput } from "@anthropic-ai/claude-agent-sdk";
import type { TestCase, TestResult, TraceEntry } from "./lib/types";
import { callJudge } from "./lib/judge";
import { judgeAlterOutput, type AlterJudgeInput, type AlterJudgeResult, type Rating } from "./lib/alter-judge";

const EVAL_ROOT = dirname(import.meta.path);
const PLUGIN_ROOT = join(EVAL_ROOT, "..");

// Parse arguments
const { values, positionals } = parseArgs({
  args: Bun.argv.slice(2),
  options: {
    verbose: { type: "boolean", short: "v", default: false },
    keep: { type: "boolean", short: "k", default: false },
  },
  allowPositionals: true,
});

const verbose = values.verbose ?? false;
const keepTemp = values.keep ?? false;

function log(...args: unknown[]) {
  if (verbose) {
    console.log("[eval]", ...args);
  }
}

/**
 * Load test case from YAML file
 */
async function loadTestCase(path: string): Promise<TestCase> {
  const content = await Bun.file(path).text();
  return parseYaml(content) as TestCase;
}

/**
 * Setup temp directory with fixtures
 */
async function setupTempDir(testCase: TestCase): Promise<string> {
  const tempDir = await $`mktemp -d`.text();
  const trimmed = tempDir.trim();

  const fixturesPath = join(EVAL_ROOT, testCase.fixtures);
  await $`cp -r ${fixturesPath}/. ${trimmed}/`;

  log(`Created temp dir: ${trimmed}`);
  return trimmed;
}

/**
 * Collect alter output (ADR, plan, components) for alter eval
 */
async function collectAlterOutput(tempDir: string): Promise<AlterJudgeInput["output"]> {
  const c3Dir = join(tempDir, ".c3");
  const adrDir = join(c3Dir, "adr");

  let adr = "";
  let plan: string | undefined;
  const componentDocs: string[] = [];

  // Collect ADR files
  try {
    const adrFiles = await $`find ${adrDir} -name "adr-*.md" ! -name "*.plan.md" 2>/dev/null | sort`.text();
    for (const file of adrFiles.trim().split("\n").filter(Boolean)) {
      adr += await Bun.file(file).text() + "\n\n";
    }
  } catch {
    // No ADR dir
  }

  // Collect plan files
  try {
    const planFiles = await $`find ${adrDir} -name "*.plan.md" 2>/dev/null | sort`.text();
    for (const file of planFiles.trim().split("\n").filter(Boolean)) {
      plan = (plan || "") + await Bun.file(file).text() + "\n\n";
    }
  } catch {
    // No plan files
  }

  // Collect component docs (c3-NNN-*.md)
  try {
    const compFiles = await $`find ${c3Dir} -name "c3-[0-9][0-9][0-9]-*.md" 2>/dev/null | sort`.text();
    for (const file of compFiles.trim().split("\n").filter(Boolean)) {
      componentDocs.push(await Bun.file(file).text());
    }
  } catch {
    // No component files
  }

  return { adr, plan, componentDocs };
}

/**
 * Convert alter judge result to standard test result format
 */
function alterResultToTestResult(
  alterResult: AlterJudgeResult,
  name: string,
  duration: number,
  trace: TraceEntry[]
): TestResult {
  const ratingToScore: Record<Rating, number> = {
    reject: 0,
    rework: 25,
    conditional: 50,
    approve: 75,
    exemplary: 100,
  };

  const score = ratingToScore[alterResult.verdict];
  const pass = alterResult.verdict === "approve" || alterResult.verdict === "exemplary";

  // Convert purposes to expectations format for compatibility
  const expectations = alterResult.purposes.flatMap((p) =>
    p.concerns.map((c) => ({
      text: `[${p.purpose}] ${c.concern}`,
      pass: c.status === "addressed",
      reasoning: c.evidence + (c.gap ? ` Gap: ${c.gap}` : ""),
    }))
  );

  return {
    name,
    pass,
    score,
    duration,
    trace,
    judgment: {
      pass,
      score,
      expectations,
      summary: alterResult.summary,
    },
  };
}

/**
 * Collect .c3/ output as string
 */
async function collectOutput(tempDir: string): Promise<string> {
  const c3Dir = join(tempDir, ".c3");

  try {
    const stat = await $`test -d ${c3Dir} && echo "yes" || echo "no"`.text();
    if (stat.trim() !== "yes") {
      return "No .c3/ directory created.";
    }
  } catch {
    return "No .c3/ directory created.";
  }

  try {
    const files = await $`find ${c3Dir} -type f -name "*.md" | sort`.text();
    const fileList = files.trim().split("\n").filter(Boolean);

    if (fileList.length === 0) {
      return ".c3/ directory exists but contains no markdown files.";
    }

    const contents: string[] = [];
    for (const file of fileList) {
      const relativePath = file.replace(tempDir + "/", "");
      const content = await Bun.file(file).text();
      contents.push(`=== ${relativePath} ===\n${content}`);
    }

    return contents.join("\n\n");
  } catch (error) {
    return `Error reading .c3/ directory: ${error}`;
  }
}

/**
 * Answer questions using a separate Claude call (Human LLM)
 */
async function answerQuestions(
  questions: Array<{ question: string; options: Array<{ label: string; description: string }>; multiSelect?: boolean }>,
  goal: string,
  constraints: string[]
): Promise<Record<string, string>> {
  const questionsText = questions
    .map((q, i) => {
      const optionsText = q.options
        .map((o, j) => `  ${j + 1}. ${o.label} - ${o.description}`)
        .join("\n");
      const selectNote = q.multiSelect ? " (select all that apply)" : "";
      return `Question ${i + 1}${selectNote}: ${q.question}\nOptions:\n${optionsText}`;
    })
    .join("\n\n");

  const constraintsText = constraints.length > 0
    ? `Constraints:\n${constraints.map((c) => `- ${c}`).join("\n")}`
    : "";

  const prompt = `You are a developer planning a project. Answer questions to help achieve the goal.

Goal: ${goal}

${constraintsText}

Instructions:
- Answer based on your understanding of the project and goal
- Pick options that best advance toward the goal
- For multi-select questions, pick all relevant options separated by comma
- Be concise - just the answer
- Return JSON only: {"0": "option_label", "1": "option_label", ...}

Questions:
${questionsText}

Respond with JSON mapping question index (0-based) to your chosen option label(s).`;

  try {
    let result = "";
    const response = query({
      prompt,
      options: {
        tools: [], // No tools needed
      },
    });

    for await (const message of response) {
      if (message.type === "assistant") {
        const content = message.message.content;
        if (typeof content === "string") {
          result += content;
        } else if (Array.isArray(content)) {
          for (const block of content) {
            if (block.type === "text") {
              result += block.text;
            }
          }
        }
      }
    }

    const jsonMatch = result.match(/\{[^}]+\}/);
    if (!jsonMatch) {
      const answers: Record<string, string> = {};
      questions.forEach((q) => {
        answers[q.question] = q.options[0]?.label || "yes";
      });
      return answers;
    }

    const parsed = JSON.parse(jsonMatch[0]);
    const answers: Record<string, string> = {};
    questions.forEach((q, i) => {
      const answer = parsed[String(i)] || parsed[i] || q.options[0]?.label || "yes";
      answers[q.question] = answer;
    });
    return answers;
  } catch {
    const answers: Record<string, string> = {};
    questions.forEach((q) => {
      answers[q.question] = q.options[0]?.label || "yes";
    });
    return answers;
  }
}

/**
 * Run the agent using SDK with plugins and hooks
 */
async function runAgent(
  testCase: TestCase,
  tempDir: string
): Promise<{ trace: TraceEntry[] }> {
  const trace: TraceEntry[] = [];

  log(`Running command: ${testCase.command}`);
  log(`Working dir: ${tempDir}`);
  log(`Plugin dir: ${PLUGIN_ROOT}`);

  const response = query({
    prompt: testCase.command,
    options: {
      cwd: tempDir,
      permissionMode: "bypassPermissions",
      allowDangerouslySkipPermissions: true,
      plugins: [
        { type: "local", path: PLUGIN_ROOT },
      ],
      hooks: {
        PreToolUse: [
          {
            hooks: [
              async (input: HookInput, toolUseID: string | undefined): Promise<HookJSONOutput> => {
                const preInput = input as PreToolUseHookInput;
                const toolName = preInput.tool_name;
                const toolInput = preInput.tool_input as Record<string, unknown>;

                // Log to trace
                trace.push({
                  timestamp: new Date().toISOString(),
                  event: "PreToolUse",
                  tool: toolName,
                  input: toolInput,
                });

                // Sandbox check for file operations
                const FILE_TOOLS = ["Read", "Write", "Edit", "Glob", "Grep"];
                if (FILE_TOOLS.includes(toolName)) {
                  const filePath = toolInput.file_path || toolInput.path || tempDir;
                  const resolved = String(filePath).startsWith("/")
                    ? String(filePath)
                    : `${tempDir}/${filePath}`;

                  if (!resolved.startsWith(tempDir)) {
                    trace.push({
                      timestamp: new Date().toISOString(),
                      event: "Blocked",
                      tool: toolName,
                      input: toolInput,
                      blocked: true,
                      reason: `Path outside sandbox: ${filePath}`,
                    });

                    return {
                      hookSpecificOutput: {
                        hookEventName: "PreToolUse",
                        permissionDecision: "deny",
                        permissionDecisionReason: `Path "${filePath}" is outside the allowed test directory.`,
                      },
                    };
                  }
                }

                // Intercept AskUserQuestion and inject answers
                if (toolName === "AskUserQuestion") {
                  const questions = toolInput.questions as Array<{
                    question: string;
                    options: Array<{ label: string; description: string }>;
                    multiSelect?: boolean;
                  }>;

                  if (questions && questions.length > 0) {
                    log(`[human] Answering ${questions.length} questions`);
                    const answers = await answerQuestions(questions, testCase.goal, testCase.constraints);

                    trace.push({
                      timestamp: new Date().toISOString(),
                      event: "HumanAnswer",
                      tool: toolName,
                      input: { questions },
                      result: answers,
                    });

                    // Format answers with descriptions for clearer feedback
                    const formattedAnswers: Record<string, string> = {};
                    for (const q of questions) {
                      const answer = answers[q.question];
                      const option = q.options.find(o => o.label === answer);
                      formattedAnswers[q.question] = option
                        ? `${option.label} - ${option.description}`
                        : answer;
                    }

                    // Instead of allowing the tool with answers, deny it with a message that includes the answers
                    // This forces the agent to see the answer and continue without "completing" AskUserQuestion
                    return {
                      hookSpecificOutput: {
                        hookEventName: "PreToolUse",
                        permissionDecision: "deny",
                        permissionDecisionReason: `[EVAL SYSTEM - User Response Received]\n\nUser answered: ${JSON.stringify(formattedAnswers)}\n\nThe user has approved. You MUST now continue executing the next steps of the workflow. Do NOT ask the same question again. Proceed immediately with creating the files specified in the ADR (component doc, README updates).`,
                      },
                    };
                  }
                }

                return {
                  hookSpecificOutput: {
                    hookEventName: "PreToolUse",
                    permissionDecision: "allow",
                  },
                };
              },
            ],
          },
        ],
        PostToolUse: [
          {
            hooks: [
              async (input: HookInput): Promise<HookJSONOutput> => {
                const postInput = input as {
                  hook_event_name: "PostToolUse";
                  tool_name: string;
                  tool_input: unknown;
                  tool_response: unknown;
                };

                trace.push({
                  timestamp: new Date().toISOString(),
                  event: "PostToolUse",
                  tool: postInput.tool_name,
                  input: postInput.tool_input as Record<string, unknown>,
                  result: postInput.tool_response,
                });

                return {};
              },
            ],
          },
        ],
      },
    },
  });

  // Process the stream
  for await (const message of response) {
    if (verbose) {
      if (message.type === "assistant") {
        const content = message.message.content;
        if (typeof content === "string") {
          log("[assistant]", content.slice(0, 100));
        } else if (Array.isArray(content)) {
          const textBlock = content.find((b) => b.type === "text");
          log("[assistant]", textBlock && "text" in textBlock ? String(textBlock.text).slice(0, 100) : "structured content");
        }
      } else if (message.type === "result") {
        log("[result]", message.subtype);
      }
    }
  }

  return { trace };
}

/**
 * Summarize trace for judge
 */
function summarizeTrace(trace: TraceEntry[]): string {
  if (trace.length === 0) {
    return "No tool calls recorded.";
  }

  const summary: string[] = [];
  for (const entry of trace) {
    if (entry.event === "PreToolUse") {
      const inputSummary = JSON.stringify(entry.input).slice(0, 200);
      summary.push(`- ${entry.tool}: ${inputSummary}`);
    } else if (entry.event === "HumanAnswer") {
      summary.push(`- Human answered: ${JSON.stringify(entry.result)}`);
    } else if (entry.event === "Blocked") {
      summary.push(`- BLOCKED ${entry.tool}: ${entry.reason}`);
    }
  }

  return summary.join("\n");
}

/**
 * Run a single test case
 */
async function runTest(casePath: string): Promise<TestResult> {
  const testCase = await loadTestCase(casePath);
  console.log(`\nRunning: ${testCase.name}`);

  const start = Date.now();
  let tempDir = "";
  let trace: TraceEntry[] = [];

  try {
    tempDir = await setupTempDir(testCase);

    // Run agent with SDK
    const result = await runAgent(testCase, tempDir);
    trace = result.trace;

    // Collect results
    const output = await collectOutput(tempDir);
    const traceSummary = summarizeTrace(trace);

    log(`Trace entries: ${trace.length}`);
    log(`Output length: ${output.length} chars`);

    const duration = Date.now() - start;

    // Use alter judge for alter eval type
    if (testCase.eval_type === "alter") {
      const alterOutput = await collectAlterOutput(tempDir);
      const changeRequest = testCase.change_request || testCase.command;

      log("Using alter judge");
      log(`ADR length: ${alterOutput.adr.length}`);
      log(`Plan: ${alterOutput.plan ? "yes" : "no"}`);
      log(`Components: ${alterOutput.componentDocs.length}`);

      // Load templates as ground truth
      const templates: AlterJudgeInput["templates"] = {};
      try {
        templates.adrTemplate = await Bun.file(join(PLUGIN_ROOT, "references/adr-template.md")).text();
        templates.planTemplate = await Bun.file(join(PLUGIN_ROOT, "references/plan-template.md")).text();
        templates.componentTemplate = await Bun.file(join(PLUGIN_ROOT, "templates/component.md")).text();
        log("Loaded templates as ground truth");
      } catch (e) {
        log(`Warning: Could not load all templates: ${e}`);
      }

      const alterResult = await judgeAlterOutput({
        changeRequest,
        output: alterOutput,
        templates,
      });

      return alterResultToTestResult(alterResult, testCase.name, duration, trace);
    }

    // Default: expectations-based judge
    const judgment = await callJudge(testCase, output, traceSummary);

    return {
      name: testCase.name,
      pass: judgment.pass,
      score: judgment.score,
      duration,
      trace,
      judgment,
    };
  } catch (error) {
    const duration = Date.now() - start;
    return {
      name: testCase.name,
      pass: false,
      score: 0,
      duration,
      trace,
      judgment: {
        pass: false,
        score: 0,
        expectations: [],
        summary: `Error: ${error}`,
      },
      error: String(error),
    };
  } finally {
    if (tempDir && !keepTemp) {
      await $`rm -rf ${tempDir}`.quiet();
    } else if (tempDir && keepTemp) {
      console.log(`  Temp dir kept: ${tempDir}`);
    }
  }
}

/**
 * Get all test case files from a path
 */
async function getTestCases(path: string): Promise<string[]> {
  const stat = await $`test -d ${path} && echo "dir" || echo "file"`.text();

  if (stat.trim() === "dir") {
    const files = await $`find ${path} -name "*.yaml" -o -name "*.yml"`.text();
    return files.trim().split("\n").filter(Boolean);
  }

  return [path];
}

/**
 * Print test result
 */
function printResult(result: TestResult): void {
  const status = result.pass ? "PASS" : "FAIL";
  const icon = result.pass ? "✓" : "✗";

  console.log(`  ${icon} ${status} (${result.score}%, ${result.duration}ms)`);

  if (!result.pass || verbose) {
    console.log(`  Summary: ${result.judgment.summary}`);

    for (const exp of result.judgment.expectations) {
      const expIcon = exp.pass ? "  ✓" : "  ✗";
      console.log(`  ${expIcon} ${exp.text}`);
      if (!exp.pass || verbose) {
        console.log(`      ${exp.reasoning}`);
      }
    }
  }

  if (result.error) {
    console.log(`  Error: ${result.error}`);
  }
}

/**
 * Save results to file
 */
async function saveResults(results: TestResult[]): Promise<void> {
  const timestamp = new Date().toISOString().replace(/[:.]/g, "-");
  const resultsDir = join(EVAL_ROOT, "results", timestamp);

  await $`mkdir -p ${resultsDir}`;

  const summary = {
    timestamp,
    total: results.length,
    passed: results.filter((r) => r.pass).length,
    failed: results.filter((r) => !r.pass).length,
    avgScore: results.reduce((sum, r) => sum + r.score, 0) / results.length,
    results: results.map((r) => ({
      name: r.name,
      pass: r.pass,
      score: r.score,
      duration: r.duration,
      summary: r.judgment.summary,
    })),
  };

  await Bun.write(
    join(resultsDir, "summary.json"),
    JSON.stringify(summary, null, 2)
  );

  for (const result of results) {
    const safeName = result.name.replace(/[^a-z0-9]/gi, "-").toLowerCase();
    await Bun.write(
      join(resultsDir, `${safeName}.json`),
      JSON.stringify(result, null, 2)
    );
  }

  console.log(`\nResults saved to: ${resultsDir}`);
}

// Main
async function main() {
  if (positionals.length === 0) {
    console.log("Usage: bun eval/run.ts <case.yaml|cases-dir> [--verbose] [--keep]");
    process.exit(1);
  }

  const testCases = await getTestCases(positionals[0]);
  console.log(`Found ${testCases.length} test case(s)`);

  const results: TestResult[] = [];

  for (const casePath of testCases) {
    const result = await runTest(casePath);
    results.push(result);
    printResult(result);
  }

  const passed = results.filter((r) => r.pass).length;
  const total = results.length;
  const avgScore = results.reduce((sum, r) => sum + r.score, 0) / total;

  console.log(`\n${"=".repeat(50)}`);
  console.log(`Results: ${passed}/${total} passed (${avgScore.toFixed(1)}% avg score)`);

  await saveResults(results);

  process.exit(passed === total ? 0 : 1);
}

main().catch((error) => {
  console.error("Fatal error:", error);
  process.exit(1);
});
