#!/usr/bin/env bun
/**
 * Test the question-based evaluation on the latest onboard output
 */

import { runQuestionEval } from "./lib/question-eval";

async function main() {
  const args = process.argv.slice(2);

  if (args.length < 2) {
    console.log("Usage: bun test-question-eval.ts <codebase-path> <docs-path>");
    console.log("");
    console.log("Example:");
    console.log("  bun test-question-eval.ts eval/fixtures/simple-express-app /tmp/tmp.xyz/.c3");
    process.exit(1);
  }

  const [codebasePath, docsPath] = args;

  console.log("🧪 Question-Based Evaluation\n");
  console.log(`Codebase: ${codebasePath}`);
  console.log(`Docs: ${docsPath}`);
  console.log("─".repeat(50));

  try {
    const result = await runQuestionEval(codebasePath, docsPath);

    console.log("\n" + "═".repeat(50));
    console.log("RESULTS");
    console.log("═".repeat(50));

    // Score breakdown
    console.log(`\n📊 Score: ${result.score.overallScore}/100\n`);
    console.log(`   Questions generated: ${result.score.totalQuestions}`);
    console.log(`   Answerable from docs: ${result.score.answerable} (${result.score.answerableRate}%)`);
    console.log(`   Correct answers: ${result.score.correct} (${result.score.accuracyRate}% of answerable)`);
    console.log(`   Partial credit: ${result.score.partial}`);

    // Question-by-question breakdown
    console.log("\n📋 Question Breakdown:\n");
    for (const q of result.questions) {
      const answer = result.answers.find((a) => a.questionId === q.id);
      const verification = result.verifications.find((v) => v.questionId === q.id);

      let status = "❓";
      if (!answer?.answerable) {
        status = "⚪"; // Not answerable
      } else if (verification?.correct) {
        status = "✅"; // Correct
      } else if (verification?.partial) {
        status = "🟡"; // Partial
      } else {
        status = "❌"; // Incorrect
      }

      console.log(`${status} [${q.category}] ${q.question}`);
      if (answer?.answerable && !verification?.correct) {
        console.log(`   → ${verification?.reasoning || "No verification"}`);
      }
    }

    // Summary
    console.log("\n" + "─".repeat(50));
    console.log("Summary:\n");
    console.log(result.summary);

    // Verdict
    const threshold = 70;
    if (result.score.overallScore >= threshold) {
      console.log(`\n✅ PASS (${result.score.overallScore} >= ${threshold})`);
    } else {
      console.log(`\n❌ FAIL (${result.score.overallScore} < ${threshold})`);
    }
  } catch (error) {
    console.error("Error:", error);
    process.exit(1);
  }
}

main();
