# Meaningful Evaluation Design

**Status: Implemented and Validated**

## Results

| Test | Golden Judge (superficial) | Question Eval (meaningful) |
|------|---------------------------|---------------------------|
| Good docs | 72/100 | 94/100 |
| Minimal docs | 8/100 | 0/100 |

The question-based approach correctly:
- Scores good docs highly (94%) because they answer questions
- Scores minimal docs at 0% because they answer nothing
- Identifies specific gaps (ADR question failed - docs claim ADRs but codebase doesn't have them)

## The Problem

Current approach: "Does this look like good docs?" (structural similarity to golden)
What we actually want: "Are these docs useful for understanding THIS system?"

## The Insight

Good documentation should answer questions a developer would ask about the system.
If someone reading ONLY the docs can understand the system, the docs are useful.

## New Approach: Question-Based Evaluation

### Flow

```
1. Analyze codebase → Extract key elements
2. Generate questions → What would a new dev ask?
3. Run onboarding → Get the docs
4. Answer questions using ONLY the docs
5. Verify answers against codebase
6. Score: % of questions answered correctly
```

### Question Categories

**Universal (any C3 docs should answer):**
- What is this system's purpose?
- What are the main containers/components?
- How do components interact?
- What patterns should I follow?
- What decisions were made and why?

**Derived (from analyzing the specific codebase):**
- How does [specific feature] work?
- Where is [specific responsibility] handled?
- What happens when [specific scenario]?

### Why This Is Better

| Old Approach | New Approach |
|--------------|--------------|
| Compare to unrelated golden | Test against THIS codebase |
| Measure structural similarity | Measure actual utility |
| "Looks like good docs" | "Works as good docs" |
| One golden fits all | Each codebase has its own questions |

## Implementation

### 1. Question Generator

Analyzes codebase to generate relevant questions:

```typescript
// Input: path to codebase
// Output: questions a new developer would ask

analyzeCodebase(path) → {
  containers: ["api", "web"],
  components: ["auth", "users", "products"],
  patterns: ["middleware", "router"],
  interactions: [
    { from: "users", to: "auth", type: "depends" }
  ]
}

generateQuestions(analysis) → [
  "What is the purpose of the auth component?",
  "How does users depend on auth?",
  "Where would I add a new API endpoint?",
  "What pattern does the router follow?"
]
```

### 2. Doc-Only Answerer

Answers questions using ONLY the generated docs:

```typescript
// Fresh LLM context with only docs
// No access to actual codebase

answerFromDocs(docs, question) → {
  answer: "Auth validates JWT tokens...",
  confidence: "high" | "medium" | "low",
  sources: ["c3-101-auth.md:15-20"]
}
```

### 3. Answer Verifier

Checks if the doc-based answer is correct:

```typescript
// Compare doc-based answer to codebase reality

verifyAnswer(answer, codebase, question) → {
  correct: true | false | "partial",
  reasoning: "The docs say X, codebase shows X"
}
```

### 4. Scorer

```typescript
score = {
  answerable: 12/15,           // Questions docs could answer
  correct: 10/12,              // Correct answers
  coverage: 0.80,              // % of codebase concepts in docs
  final: (answerable * correct * coverage)
}
```

## Test Case Format

```yaml
name: "Onboard simple Express app"
fixtures: "fixtures/simple-express-app"
command: "Please onboard this project..."

evaluation:
  type: "question-based"

  # Auto-generated from codebase analysis + these universal questions
  universal_questions:
    - "What is this system for?"
    - "What are the main components?"
    - "How would I add a new feature?"

  # Minimum thresholds
  thresholds:
    answerable: 0.8      # 80% of questions should be answerable
    accuracy: 0.9        # 90% of answers should be correct
```

## Why This Makes Evaluation Meaningful

1. **Grounded**: Questions come from the actual codebase
2. **Practical**: Tests what matters - can someone use these docs?
3. **Fair**: Each system evaluated against its own requirements
4. **Measurable**: Clear pass/fail on each question
5. **Debuggable**: Know exactly which questions failed and why
