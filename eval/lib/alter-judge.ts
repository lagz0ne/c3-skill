import { query } from "@anthropic-ai/claude-agent-sdk";

// Types
export type Rating = "reject" | "rework" | "conditional" | "approve" | "exemplary";
export type ConcernStatus = "addressed" | "partial" | "missing";

export interface ConcernResult {
  concern: string;
  status: ConcernStatus;
  evidence: string;
  gap?: string;
}

export interface PurposeResult {
  purpose: string;
  rating: Rating;
  concerns: ConcernResult[];
}

export interface AlterJudgeResult {
  purposes: PurposeResult[];
  verdict: Rating;
  summary: string;
}

export interface AlterJudgeInput {
  changeRequest: string;
  output: {
    adr: string;
    plan?: string;
    componentDocs: string[];
  };
  // Templates as ground truth for what's required
  templates?: {
    adrTemplate?: string;
    planTemplate?: string;
    componentTemplate?: string;
  };
}

// Purpose definitions with concerns
const PURPOSES = [
  {
    name: "Approve the Change",
    description: "Engineer reviewing ADR before accepting",
    concerns: [
      { id: "problem_validity", question: "Does the stated problem match the change request? Is it real, not invented?" },
      { id: "decision_clarity", question: "Is the decision specific and actionable?" },
      { id: "rationale_soundness", question: "Is there reasoning for the approach? (Alternatives table optional - any explanation of why this approach counts)" },
      { id: "scope_appropriateness", question: "Is the scope appropriate? Not over-scoped or under-scoped?" },
      { id: "risk_acknowledgment", question: "Are breaking changes noted? Dependencies identified?" },
    ],
  },
  {
    name: "Execute the Change",
    description: "Engineer implementing after approval",
    concerns: [
      { id: "file_identification", question: "Which files to create/modify? Are paths clear?" },
      { id: "order_clarity", question: "Is the sequence of changes explicit? Dependencies respected?" },
      { id: "content_guidance", question: "What goes in each file? Enough detail to write?" },
      { id: "integration_points", question: "How does new code connect to existing? Hand-offs clear?" },
    ],
  },
  {
    name: "Verify Correctness",
    description: "Engineer checking after implementation",
    concerns: [
      { id: "verification_steps", question: "Are concrete verification checks listed?" },
      { id: "independence", question: "Can verification happen without knowing implementation details?" },
      { id: "completeness", question: "Do all affected areas have verification?" },
    ],
  },
  {
    name: "Maintain Later",
    description: "Future engineer understanding the change",
    concerns: [
      { id: "context", question: "Is it clear why this change was made? Traceable to ADR?" },
      { id: "dependencies", question: "What does this affect? What affects this?" },
      { id: "evolution_path", question: "How might this change in future?" },
    ],
  },
];

function buildJudgePrompt(input: AlterJudgeInput): string {
  const purposesSection = PURPOSES.map((p) => {
    const concernsList = p.concerns.map((c) => `    - ${c.question}`).join("\n");
    return `### ${p.name}\n*${p.description}*\n\nConcerns:\n${concernsList}`;
  }).join("\n\n");

  const componentDocsSection = input.output.componentDocs.length > 0
    ? input.output.componentDocs.map((doc, i) => `### Component Doc ${i + 1}\n${doc}`).join("\n\n")
    : "(none)";

  // Build templates section if provided
  let templatesSection = "";
  if (input.templates) {
    templatesSection = `
## Templates (Ground Truth for Requirements)

IMPORTANT: Use these templates to determine what is REQUIRED vs OPTIONAL.
Only flag something as missing if the template explicitly requires it.
Do NOT apply generic "good documentation" standards - only what templates require.

${input.templates.adrTemplate ? `### ADR Template (What ADR must contain)\n${input.templates.adrTemplate}` : ""}

${input.templates.planTemplate ? `### Plan Template (What Plan must contain)\n${input.templates.planTemplate}` : ""}

${input.templates.componentTemplate ? `### Component Template (What Component doc must contain)\n${input.templates.componentTemplate}` : ""}
`;
  }

  return `You are evaluating architectural change documentation.

## Change Request (Ground Truth)
${input.changeRequest}
${templatesSection}
## Output Documents

### ADR
${input.output.adr}

### Plan
${input.output.plan || "(none)"}

### Component Docs
${componentDocsSection}

---

## Evaluation Task

For each purpose below, evaluate whether the concerns are addressed by the output documents.

CRITICAL: If templates are provided above, use them to determine what is actually required.
- Only mark something as "missing" if the template explicitly requires it
- Do NOT penalize for missing "future evolution" or "extensibility" if not in template
- Do NOT penalize for missing "runtime verification" if the change is documentation-only
- Do NOT penalize for implementation details (how to mount routes) if dependencies are documented
- Focus on whether output follows the templates, not generic best practices

${purposesSection}

---

## Instructions

For each concern:
1. Check if the template requires this (if templates provided)
2. Search the output documents for evidence that addresses it
3. Rate as: "addressed" (fully covered), "partial" (some coverage, gaps remain), or "missing" (not covered AND required by template)
4. Provide evidence (quote or reference) if addressed/partial
5. Note what's missing if partial/missing

For each purpose, determine a rating:
- "reject" = most concerns missing, fundamentally flawed
- "rework" = major gaps, can't proceed safely
- "conditional" = usable with caveats
- "approve" = solid, ready to proceed
- "exemplary" = reference quality

Overall verdict = worst purpose rating (weakest link)

Return ONLY valid JSON:
{
  "purposes": [
    {
      "purpose": "Approve the Change",
      "rating": "approve",
      "concerns": [
        {
          "concern": "problem_validity",
          "status": "addressed",
          "evidence": "quote or reference from docs"
        },
        {
          "concern": "decision_clarity",
          "status": "partial",
          "evidence": "what was found",
          "gap": "what's missing"
        }
      ]
    }
  ],
  "verdict": "conditional",
  "summary": "One sentence overall assessment"
}`;
}

function aggregateRating(concerns: ConcernResult[]): Rating {
  const statusCounts = {
    addressed: 0,
    partial: 0,
    missing: 0,
  };

  for (const c of concerns) {
    statusCounts[c.status]++;
  }

  const total = concerns.length;
  const addressedRatio = statusCounts.addressed / total;
  const missingRatio = statusCounts.missing / total;

  if (missingRatio > 0.5) return "reject";
  if (missingRatio > 0.25 || statusCounts.partial > total * 0.5) return "rework";
  if (statusCounts.missing > 0 || statusCounts.partial > 0) return "conditional";
  if (addressedRatio === 1) return "exemplary";
  return "approve";
}

function worstRating(ratings: Rating[]): Rating {
  const order: Rating[] = ["reject", "rework", "conditional", "approve", "exemplary"];
  let worst = 4; // exemplary
  for (const r of ratings) {
    const idx = order.indexOf(r);
    if (idx < worst) worst = idx;
  }
  return order[worst];
}

export async function judgeAlterOutput(input: AlterJudgeInput): Promise<AlterJudgeResult> {
  const prompt = buildJudgePrompt(input);

  let result = "";

  const response = query({
    prompt,
    options: {
      allowedTools: [],
      model: "sonnet",
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
      purposes: [],
      verdict: "reject",
      summary: "Failed to parse judge response",
    };
  }

  try {
    const parsed = JSON.parse(jsonMatch[0]) as AlterJudgeResult;

    // Recompute verdict from actual concern statuses (don't trust LLM's verdict)
    const allConcerns = parsed.purposes.flatMap((p) => p.concerns);
    const recomputedVerdict = aggregateRating(allConcerns);

    // Also recompute purpose-level ratings
    for (const purpose of parsed.purposes) {
      purpose.rating = aggregateRating(purpose.concerns);
    }

    return {
      ...parsed,
      verdict: recomputedVerdict,
    };
  } catch {
    return {
      purposes: [],
      verdict: "reject",
      summary: "Failed to parse judge JSON",
    };
  }
}

/**
 * Run alter judge and print formatted results
 */
export async function runAlterJudge(
  input: AlterJudgeInput,
  verbose: boolean = false
): Promise<AlterJudgeResult> {
  console.log("\n📋 Evaluating alter output...\n");

  const result = await judgeAlterOutput(input);

  // Print results
  const verdictIcon = {
    reject: "❌",
    rework: "🔄",
    conditional: "⚠️",
    approve: "✅",
    exemplary: "🌟",
  };

  console.log(`${verdictIcon[result.verdict]} Verdict: ${result.verdict.toUpperCase()}`);
  console.log(`   ${result.summary}\n`);

  if (verbose || result.verdict === "reject" || result.verdict === "rework") {
    for (const purpose of result.purposes) {
      console.log(`\n   ${verdictIcon[purpose.rating]} ${purpose.purpose}: ${purpose.rating}`);

      for (const concern of purpose.concerns) {
        const statusIcon = {
          addressed: "✓",
          partial: "~",
          missing: "✗",
        };
        console.log(`      ${statusIcon[concern.status]} ${concern.concern}`);

        if (concern.evidence && (verbose || concern.status !== "addressed")) {
          console.log(`         → ${concern.evidence.substring(0, 100)}...`);
        }
        if (concern.gap) {
          console.log(`         ⚠ Gap: ${concern.gap}`);
        }
      }
    }
  }

  return result;
}
