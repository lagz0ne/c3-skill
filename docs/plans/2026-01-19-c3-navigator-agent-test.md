# C3 Navigator Agent - Test Plan

**Date:** 2026-01-19
**Goal:** Verify c3-navigator agent works correctly

## Test Categories

| Category | Method | Run Time | Priority |
|----------|--------|----------|----------|
| Triggering | Manual observation | <1min | High |
| Output quality | LLM judge | 30-60s | High |
| Diagram generation | URL validation | 5s | Medium |
| Edge cases | Manual + YAML cases | 30s each | Low |

## Test 1: Triggering Verification

**Purpose:** Verify agent triggers in correct conditions

### Test Cases

| ID | Scenario | Input | Expected | Pass Criteria |
|----|----------|-------|----------|---------------|
| T1.1 | Question with .c3/ | "How does auth work?" in documented-api | Agent triggers | See agent in response |
| T1.2 | Question without .c3/ | "How does auth work?" in no-c3-project | Agent does NOT trigger | Direct response, no agent |
| T1.3 | Command with .c3/ | "Add a login button" in documented-api | Agent does NOT trigger | Routes to alter or direct action |
| T1.4 | C3 ID reference | "Explain c3-101" in documented-api | Agent triggers | Agent responds with component details |

### Manual Test Procedure

```bash
# Setup: Navigate to fixture project
cd eval/fixtures/documented-api

# Test T1.1: Ask a question
# In Claude Code, ask: "How does the API handle requests?"
# VERIFY: c3-navigator agent is invoked

# Test T1.2: Navigate to non-C3 project
cd ../simple-express-app  # No .c3/ directory
# In Claude Code, ask: "How does the API handle requests?"
# VERIFY: c3-navigator NOT invoked, direct response
```

**Pass criteria:** All 4 triggering cases behave as expected

## Test 2: Output Quality (LLM Judge)

**Purpose:** Verify responses are accurate and useful

### Eval Case File

```yaml
# eval/cases/navigator-auth-query.yaml
name: "Navigator answers auth question"
fixtures: "documented-api"
eval_type: "navigator"

command: |
  How does authentication work in this system?

goal: |
  The c3-navigator agent should read .c3/ docs and provide
  an accurate explanation of authentication flow with diagram.

constraints:
  - Must reference actual component IDs from docs
  - Must include diagram URL from diashort
  - Must NOT hallucinate components not in docs

expectations:
  structure:
    - "References specific c3-* component IDs"
    - "Mentions technology from docs (e.g., JWT, session)"
    - "Includes diashort URL (https://diashort...)"
  accuracy:
    - "All mentioned components exist in .c3/"
    - "Flow description matches docs"
  completeness:
    - "Answers the actual question"
    - "Provides actionable code references"
```

### Judge Criteria

```typescript
interface NavigatorJudgeResult {
  accuracy: number;      // 0-100: facts match .c3/ docs
  completeness: number;  // 0-100: answered the question
  diagramProduced: boolean;
  diagramValid: boolean; // diashort URL returns 200
  verdict: "pass" | "partial" | "fail";
}

// Pass thresholds:
// - accuracy >= 80
// - completeness >= 70
// - diagramProduced: true
// - diagramValid: true
```

**Pass criteria:** Score >= thresholds, no false claims

## Test 3: Diagram Generation

**Purpose:** Verify diashort integration works

### Test Cases

| ID | Scenario | Expected |
|----|----------|----------|
| T3.1 | Architecture overview query | Mermaid flowchart generated |
| T3.2 | Sequence query ("how does X flow to Y") | Mermaid sequence diagram |
| T3.3 | Component query ("what is c3-101") | Simple component diagram |

### Validation Script

```typescript
// eval/lib/diashort-validator.ts
export async function validateDiashortUrl(url: string): Promise<{
  valid: boolean;
  accessible: boolean;
  error?: string;
}> {
  // Check URL format
  if (!url.startsWith("https://diashort.apps.quickable.co/d/")) {
    return { valid: false, accessible: false, error: "Invalid URL format" };
  }

  // Check accessibility
  try {
    const res = await fetch(url, { method: "HEAD" });
    return { valid: true, accessible: res.ok };
  } catch (e) {
    return { valid: true, accessible: false, error: String(e) };
  }
}
```

**Pass criteria:** URL valid and accessible (HTTP 200)

## Test 4: Edge Cases

| ID | Scenario | Fixture | Expected Behavior |
|----|----------|---------|-------------------|
| T4.1 | Empty .c3/ | Create empty `.c3/` | Graceful message: "C3 docs empty, use /onboard" |
| T4.2 | Question not in docs | documented-api + "How does payment work?" | Explicit: "Not documented in C3, searched code..." |
| T4.3 | Malformed docs | Create invalid markdown in .c3/ | Best-effort response, no crash |
| T4.4 | Very long docs | Large fixture | Summarizes, doesn't timeout (<2min) |

### Edge Case Eval Files

```yaml
# eval/cases/navigator-edge-empty.yaml
name: "Navigator handles empty .c3/"
fixtures: "empty-c3-project"
eval_type: "navigator"

command: "How does the system work?"

expectations:
  - "Mentions .c3/ is empty or incomplete"
  - "Suggests /onboard or manual documentation"
  - "Does NOT hallucinate architecture"
```

**Pass criteria:** Graceful handling, no crashes, honest about limitations

## Test 5: Integration Test

**Purpose:** End-to-end verification

### Procedure

```bash
# 1. Start with clean fixture
cd eval/fixtures/documented-api

# 2. Invoke Claude Code
claude

# 3. Ask architecture question
> How does the request flow from API entry to database?

# 4. Verify in response:
#    - Agent was used (visible in output)
#    - Response references actual components (c3-201, etc.)
#    - Diagram URL included
#    - Diagram URL loads in browser

# 5. Ask follow-up
> What about error handling in that flow?

# 6. Verify follow-up:
#    - Builds on previous context OR re-reads docs
#    - References error-handling patterns if documented
```

**Pass criteria:** Full flow works, diagrams render, responses accurate

## Running Tests

### Quick Smoke Test (Manual)
```bash
cd eval/fixtures/documented-api
claude
# Ask: "Explain the system architecture"
# Verify: Agent triggers, diagram produced, response accurate
```

### Full Test Suite
```bash
# Run all navigator eval cases
bun eval/run.ts eval/cases/navigator-*.yaml

# Validate all diashort URLs from results
bun eval/lib/validate-diashort-urls.ts eval/results/latest/
```

## Test Fixtures Required

| Fixture | Purpose | Status |
|---------|---------|--------|
| `documented-api` | Primary test fixture with full .c3/ | Exists |
| `simple-express-app` | No .c3/ - verify non-triggering | Exists |
| `empty-c3-project` | Empty .c3/ directory | TODO: Create |

## Success Metrics

| Metric | Target |
|--------|--------|
| Triggering accuracy | 100% (triggers when should, doesn't when shouldn't) |
| Output accuracy | >= 80% (facts match docs) |
| Output completeness | >= 70% (answers question) |
| Diagram generation rate | >= 90% (produces diagram when helpful) |
| Diagram validity | 100% (all URLs accessible) |

## Definition of Done

- [ ] All triggering tests pass
- [ ] Output quality eval cases pass (>=80% accuracy)
- [ ] Diagram URLs validate
- [ ] Edge cases handled gracefully
- [ ] Integration test passes end-to-end
