import { query } from "@anthropic-ai/claude-agent-sdk";

export interface CodebaseAnalysis {
  purpose: string;
  containers: string[];
  components: {
    name: string;
    type: "foundation" | "feature";
    responsibility: string;
    file: string;
  }[];
  interactions: {
    from: string;
    to: string;
    type: string;
  }[];
  patterns: string[];
}

export interface Question {
  id: string;
  category: "purpose" | "structure" | "behavior" | "patterns" | "decisions";
  question: string;
  expectedAnswer?: string; // From codebase analysis
}

export interface AnswerResult {
  questionId: string;
  answerable: boolean;
  answer: string;
  confidence: "high" | "medium" | "low";
  sources: string[];
}

export interface VerificationResult {
  questionId: string;
  correct: boolean;
  partial: boolean;
  reasoning: string;
}

export interface QuestionEvalResult {
  questions: Question[];
  answers: AnswerResult[];
  verifications: VerificationResult[];
  score: {
    totalQuestions: number;
    answerable: number;
    correct: number;
    partial: number;
    answerableRate: number;
    accuracyRate: number;
    overallScore: number;
  };
  summary: string;
}

/**
 * Analyze codebase to understand its structure
 */
export async function analyzeCodebase(
  codebasePath: string
): Promise<CodebaseAnalysis> {
  // Read key files to understand structure
  const files = await listCodeFiles(codebasePath);
  const fileContents = await readSampleFiles(codebasePath, files);

  const prompt = `Analyze this codebase and extract its architectural structure.

## Files Found
${files.join("\n")}

## Sample Contents
${fileContents}

---

Return a JSON object with:
{
  "purpose": "One sentence describing what this system does",
  "containers": ["List of deployable units"],
  "components": [
    {
      "name": "ComponentName",
      "type": "foundation or feature",
      "responsibility": "What it does",
      "file": "path/to/file.ts"
    }
  ],
  "interactions": [
    { "from": "ComponentA", "to": "ComponentB", "type": "depends/calls/imports" }
  ],
  "patterns": ["Middleware pattern", "Router pattern", etc]
}

Be specific about THIS codebase, not generic.`;

  let result = "";
  const response = query({
    prompt,
    options: { allowedTools: [], model: "haiku" },
  });

  for await (const message of response) {
    if (message.type === "assistant") {
      const content = message.message.content;
      if (typeof content === "string") result += content;
      else if (Array.isArray(content)) {
        for (const block of content) {
          if (block.type === "text") result += block.text;
        }
      }
    }
  }

  const jsonMatch = result.match(/\{[\s\S]*\}/);
  if (!jsonMatch) throw new Error("Failed to analyze codebase");
  return JSON.parse(jsonMatch[0]) as CodebaseAnalysis;
}

/**
 * Generate questions based on codebase analysis
 */
export function generateQuestions(analysis: CodebaseAnalysis): Question[] {
  const questions: Question[] = [];
  let id = 0;

  // Purpose questions
  questions.push({
    id: `q${++id}`,
    category: "purpose",
    question: "What is this system's primary purpose?",
    expectedAnswer: analysis.purpose,
  });

  // Structure questions - one per component
  for (const comp of analysis.components) {
    questions.push({
      id: `q${++id}`,
      category: "structure",
      question: `What is the responsibility of the ${comp.name} component?`,
      expectedAnswer: comp.responsibility,
    });
  }

  // Behavior questions - based on interactions
  for (const interaction of analysis.interactions) {
    questions.push({
      id: `q${++id}`,
      category: "behavior",
      question: `How does ${interaction.from} interact with ${interaction.to}?`,
      expectedAnswer: `${interaction.from} ${interaction.type} ${interaction.to}`,
    });
  }

  // Pattern questions
  for (const pattern of analysis.patterns) {
    questions.push({
      id: `q${++id}`,
      category: "patterns",
      question: `How is the ${pattern} pattern implemented in this codebase?`,
    });
  }

  // Universal questions
  questions.push(
    {
      id: `q${++id}`,
      category: "structure",
      question: "How would I add a new feature to this system?",
    },
    {
      id: `q${++id}`,
      category: "decisions",
      question: "What architectural decisions have been documented and why?",
    }
  );

  return questions;
}

/**
 * Answer questions using ONLY the documentation (fresh context)
 */
export async function answerFromDocs(
  docsPath: string,
  questions: Question[]
): Promise<AnswerResult[]> {
  // Read all docs
  const docsContent = await readAllDocs(docsPath);

  const prompt = `You are a developer who has ONLY read the architecture documentation below.
You have NO access to the actual codebase.

## Documentation
${docsContent}

---

Answer each question based ONLY on what's in the documentation.
If the documentation doesn't answer a question, say so.

## Questions
${questions.map((q) => `${q.id}: ${q.question}`).join("\n")}

---

For each question, return JSON:
[
  {
    "questionId": "q1",
    "answerable": true/false,
    "answer": "Your answer based on docs",
    "confidence": "high/medium/low",
    "sources": ["file.md:section"]
  }
]`;

  let result = "";
  const response = query({
    prompt,
    options: { allowedTools: [], model: "sonnet" },
  });

  for await (const message of response) {
    if (message.type === "assistant") {
      const content = message.message.content;
      if (typeof content === "string") result += content;
      else if (Array.isArray(content)) {
        for (const block of content) {
          if (block.type === "text") result += block.text;
        }
      }
    }
  }

  const jsonMatch = result.match(/\[[\s\S]*\]/);
  if (!jsonMatch) throw new Error("Failed to get answers");
  return JSON.parse(jsonMatch[0]) as AnswerResult[];
}

/**
 * Verify answers against codebase reality
 */
export async function verifyAnswers(
  codebasePath: string,
  questions: Question[],
  answers: AnswerResult[]
): Promise<VerificationResult[]> {
  const files = await listCodeFiles(codebasePath);
  const fileContents = await readSampleFiles(codebasePath, files);

  const qaPairs = questions.map((q) => {
    const a = answers.find((a) => a.questionId === q.id);
    return {
      id: q.id,
      question: q.question,
      expectedAnswer: q.expectedAnswer,
      docAnswer: a?.answer || "Not answered",
      answerable: a?.answerable || false,
    };
  });

  const prompt = `Verify if these documentation-based answers are correct by checking against the actual codebase.

## Codebase
${fileContents}

## Question-Answer Pairs
${JSON.stringify(qaPairs, null, 2)}

---

For each Q&A pair, verify if the doc-based answer matches codebase reality.
Return JSON:
[
  {
    "questionId": "q1",
    "correct": true/false,
    "partial": true/false,
    "reasoning": "Why correct or incorrect"
  }
]`;

  let result = "";
  const response = query({
    prompt,
    options: { allowedTools: [], model: "sonnet" },
  });

  for await (const message of response) {
    if (message.type === "assistant") {
      const content = message.message.content;
      if (typeof content === "string") result += content;
      else if (Array.isArray(content)) {
        for (const block of content) {
          if (block.type === "text") result += block.text;
        }
      }
    }
  }

  const jsonMatch = result.match(/\[[\s\S]*\]/);
  if (!jsonMatch) throw new Error("Failed to verify answers");
  return JSON.parse(jsonMatch[0]) as VerificationResult[];
}

/**
 * Run full question-based evaluation
 */
export async function runQuestionEval(
  codebasePath: string,
  docsPath: string
): Promise<QuestionEvalResult> {
  console.log("📊 Running question-based evaluation\n");

  // Step 1: Analyze codebase
  console.log("1. Analyzing codebase...");
  const analysis = await analyzeCodebase(codebasePath);
  console.log(`   Found: ${analysis.components.length} components, ${analysis.patterns.length} patterns`);

  // Step 2: Generate questions
  console.log("2. Generating questions...");
  const questions = generateQuestions(analysis);
  console.log(`   Generated: ${questions.length} questions`);

  // Step 3: Answer from docs only
  console.log("3. Answering questions from docs only...");
  const answers = await answerFromDocs(docsPath, questions);
  const answerable = answers.filter((a) => a.answerable).length;
  console.log(`   Answerable: ${answerable}/${questions.length}`);

  // Step 4: Verify answers
  console.log("4. Verifying answers against codebase...");
  const verifications = await verifyAnswers(codebasePath, questions, answers);
  const correct = verifications.filter((v) => v.correct).length;
  const partial = verifications.filter((v) => v.partial && !v.correct).length;
  console.log(`   Correct: ${correct}, Partial: ${partial}`);

  // Calculate scores
  const answerableRate = answerable / questions.length;
  const accuracyRate = answerable > 0 ? correct / answerable : 0;
  const overallScore = Math.round(answerableRate * accuracyRate * 100);

  const result: QuestionEvalResult = {
    questions,
    answers,
    verifications,
    score: {
      totalQuestions: questions.length,
      answerable,
      correct,
      partial,
      answerableRate: Math.round(answerableRate * 100),
      accuracyRate: Math.round(accuracyRate * 100),
      overallScore,
    },
    summary: generateSummary(questions, answers, verifications, overallScore),
  };

  return result;
}

function generateSummary(
  questions: Question[],
  answers: AnswerResult[],
  verifications: VerificationResult[],
  score: number
): string {
  const unanswered = questions.filter((q) => {
    const a = answers.find((a) => a.questionId === q.id);
    return !a?.answerable;
  });

  const incorrect = verifications.filter((v) => !v.correct && !v.partial);

  let summary = `Score: ${score}/100\n\n`;

  if (unanswered.length > 0) {
    summary += `Questions docs couldn't answer:\n`;
    for (const q of unanswered.slice(0, 3)) {
      summary += `- ${q.question}\n`;
    }
    if (unanswered.length > 3) {
      summary += `  ...and ${unanswered.length - 3} more\n`;
    }
    summary += "\n";
  }

  if (incorrect.length > 0) {
    summary += `Incorrect answers:\n`;
    for (const v of incorrect.slice(0, 3)) {
      const q = questions.find((q) => q.id === v.questionId);
      summary += `- ${q?.question}: ${v.reasoning}\n`;
    }
    if (incorrect.length > 3) {
      summary += `  ...and ${incorrect.length - 3} more\n`;
    }
  }

  return summary;
}

// Helper functions
async function listCodeFiles(path: string): Promise<string[]> {
  const files: string[] = [];
  for await (const entry of new Bun.Glob("**/*.{ts,js,tsx,jsx}").scan({
    cwd: path,
  })) {
    if (!entry.includes("node_modules")) {
      files.push(entry);
    }
  }
  return files.slice(0, 20); // Limit for context
}

async function readSampleFiles(
  basePath: string,
  files: string[]
): Promise<string> {
  const contents: string[] = [];
  for (const file of files.slice(0, 10)) {
    try {
      const content = await Bun.file(`${basePath}/${file}`).text();
      contents.push(`### ${file}\n\`\`\`\n${content.slice(0, 2000)}\n\`\`\``);
    } catch {
      // Skip unreadable files
    }
  }
  return contents.join("\n\n");
}

async function readAllDocs(docsPath: string): Promise<string> {
  const contents: string[] = [];
  for await (const entry of new Bun.Glob("**/*.md").scan({ cwd: docsPath })) {
    try {
      const content = await Bun.file(`${docsPath}/${entry}`).text();
      contents.push(`### ${entry}\n${content}`);
    } catch {
      // Skip unreadable files
    }
  }
  return contents.join("\n\n---\n\n");
}
