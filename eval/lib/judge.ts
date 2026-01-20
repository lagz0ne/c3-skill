import { query } from "@anthropic-ai/claude-agent-sdk";
import type { TestCase, JudgeResponse } from "./types";

const JUDGE_PROMPT = `You evaluate if an AI agent's output serves a stated goal.

Be strict but fair. Focus on whether the output achieves the intent, not whether it matches exact wording.

Evaluate each expectation independently. An expectation passes if the output reasonably satisfies it.

Return your evaluation as valid JSON only, no other text.`;

/**
 * Call Judge via Agent SDK
 */
export async function callJudge(
  testCase: TestCase,
  output: string,
  traceSummary: string
): Promise<JudgeResponse> {
  const constraintsText =
    testCase.constraints.length > 0
      ? testCase.constraints.map((c) => `- ${c}`).join("\n")
      : "None specified";

  const expectationsText = testCase.expectations
    .map((e, i) => `${i + 1}. ${e}`)
    .join("\n");

  const prompt = `${JUDGE_PROMPT}

## Goal
${testCase.goal}

## Constraints
${constraintsText}

## Expectations
${expectationsText}

## Output Produced
${output}

## Trace (tool calls made)
${traceSummary}

---

Evaluate each expectation. Return JSON only:
{
  "pass": boolean (true if ALL expectations pass),
  "score": number (0-100, percentage of expectations met),
  "expectations": [
    { "text": "expectation text", "pass": boolean, "reasoning": "why pass/fail" }
  ],
  "summary": "Overall assessment in 1-2 sentences"
}`;

  try {
    let result = "";

    // Use Agent SDK query - no tools needed for judging
    const response = query({
      prompt,
      options: {
        allowedTools: [], // Judge doesn't need any tools
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
        pass: false,
        score: 0,
        expectations: testCase.expectations.map((e) => ({
          text: e,
          pass: false,
          reasoning: "Failed to parse judge response",
        })),
        summary: "Judge response was not valid JSON",
      };
    }

    return JSON.parse(jsonMatch[0]) as JudgeResponse;
  } catch (error) {
    return {
      pass: false,
      score: 0,
      expectations: testCase.expectations.map((e) => ({
        text: e,
        pass: false,
        reasoning: `Judge error: ${error}`,
      })),
      summary: `Judge failed: ${error}`,
    };
  }
}
