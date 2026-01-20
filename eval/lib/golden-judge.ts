import { query } from "@anthropic-ai/claude-agent-sdk";
import { parse } from "yaml";
import type { GoldenMetadata, GoldenJudgeResponse } from "./types";

const PASS_THRESHOLD = 80;

/**
 * Load golden metadata from YAML file
 */
export async function loadGoldenMetadata(
  goldenYamlPath: string
): Promise<GoldenMetadata> {
  const content = await Bun.file(goldenYamlPath).text();
  return parse(content) as GoldenMetadata;
}

/**
 * Load all files from a directory recursively as a string summary
 */
async function loadDirectoryContent(dirPath: string): Promise<string> {
  const result: string[] = [];

  async function walk(dir: string, prefix: string = "") {
    const entries = await Array.fromAsync(
      new Bun.Glob("**/*.{md,yaml,yml}").scan({ cwd: dir })
    );

    for (const entry of entries.sort()) {
      const fullPath = `${dir}/${entry}`;
      const content = await Bun.file(fullPath).text();
      result.push(`\n### ${prefix}${entry}\n\`\`\`\n${content}\n\`\`\``);
    }
  }

  await walk(dirPath);
  return result.join("\n");
}

/**
 * Compare eval output against golden example using LLM judge
 */
export async function compareToGolden(
  goldenPath: string,
  goldenMetadataPath: string,
  outputPath: string
): Promise<GoldenJudgeResponse> {
  const metadata = await loadGoldenMetadata(goldenMetadataPath);
  const goldenContent = await loadDirectoryContent(goldenPath);
  const outputContent = await loadDirectoryContent(outputPath);

  const dimensionsList = Object.entries(metadata.evaluation_dimensions)
    .map(([dim, checks]) => {
      return `### ${dim}\n${checks.map((c) => `- ${c}`).join("\n")}`;
    })
    .join("\n\n");

  const prompt = `You are evaluating C3 architecture documentation quality by comparing an agent's output against a golden example.

## Golden Example Context
**Project:** ${metadata.input_context.project_name}
**Domain:** ${metadata.input_context.domain}

The golden example represents ideal quality for this type of onboarding.

## Evaluation Dimensions
${dimensionsList}

## Golden Example (The Standard)
${goldenContent}

## Output to Evaluate
${outputContent}

---

Compare the output to the golden example across each evaluation dimension.
For each dimension, evaluate all checks and provide:
- Whether each check passes
- Specific reasoning for the assessment
- Overall dimension score (0-100)

Focus on **quality parity** - the output doesn't need to match the golden exactly, but should demonstrate the same level of architectural thinking, completeness, and C3 compliance.

Return your evaluation as valid JSON only:
{
  "overall_score": number (0-100),
  "dimensions": [
    {
      "dimension": "structure",
      "score": number (0-100),
      "checks": [
        { "check": "check text", "pass": boolean, "reasoning": "why" }
      ]
    }
  ],
  "gaps": ["specific gap 1", "specific gap 2"],
  "verdict": "pass" | "fail",
  "summary": "1-2 sentence overall assessment"
}`;

  try {
    let result = "";

    const response = query({
      prompt,
      options: {
        allowedTools: [],
        model: "sonnet", // Use Sonnet for judging - good balance of speed and quality
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

    // Extract JSON from response
    const jsonMatch = result.match(/\{[\s\S]*\}/);
    if (!jsonMatch) {
      return {
        overall_score: 0,
        dimensions: [],
        gaps: ["Failed to parse judge response"],
        verdict: "fail",
        summary: "Judge response was not valid JSON",
      };
    }

    const judgeResult = JSON.parse(jsonMatch[0]) as GoldenJudgeResponse;

    // Apply threshold for verdict
    if (judgeResult.overall_score >= PASS_THRESHOLD) {
      judgeResult.verdict = "pass";
    } else {
      judgeResult.verdict = "fail";
    }

    return judgeResult;
  } catch (error) {
    return {
      overall_score: 0,
      dimensions: [],
      gaps: [`Judge error: ${error}`],
      verdict: "fail",
      summary: `Judge failed: ${error}`,
    };
  }
}

/**
 * Run golden comparison and format results for console
 */
export async function runGoldenComparison(
  goldenPath: string,
  goldenMetadataPath: string,
  outputPath: string,
  verbose: boolean = false
): Promise<GoldenJudgeResponse> {
  console.log(`\n📊 Comparing against golden: ${goldenMetadataPath}`);

  const result = await compareToGolden(
    goldenPath,
    goldenMetadataPath,
    outputPath
  );

  // Print results
  const icon = result.verdict === "pass" ? "✅" : "❌";
  console.log(`\n${icon} ${result.verdict.toUpperCase()} (${result.overall_score}/100)`);
  console.log(`   ${result.summary}`);

  if (verbose || result.verdict === "fail") {
    console.log("\n   Dimension Scores:");
    for (const dim of result.dimensions) {
      const dimIcon = dim.score >= PASS_THRESHOLD ? "✓" : "✗";
      console.log(`   ${dimIcon} ${dim.dimension}: ${dim.score}/100`);

      for (const check of dim.checks) {
        const checkIcon = check.pass ? "  ✓" : "  ✗";
        console.log(`     ${checkIcon} ${check.check}`);
        if (!check.pass || verbose) {
          console.log(`       → ${check.reasoning}`);
        }
      }
    }

    if (result.gaps.length > 0) {
      console.log("\n   Gaps:");
      for (const gap of result.gaps) {
        console.log(`   • ${gap}`);
      }
    }
  }

  return result;
}
