# Test: Explore Existing Architecture

## Setup

**Fixture:** `fixtures/02-explore-architecture/`

Pre-existing state (MESSY but valid v3):
- `.c3/README.md` - Context with 4 containers listed, but:
  - Diagram incomplete (missing c3-3, c3-4)
  - Brief overview (single sentence)
- `.c3/c3-1-api/README.md` - Container with:
  - Mixed component IDs (c3-101, c3-102, c3-103 + "OrderController")
  - TODOs in notes
- `.c3/c3-2-database/README.md` - Stub only, missing required sections
- c3-3 (Search) - Listed in Context but NO folder exists
- c3-4 (Payment) - Listed in Context but NO folder exists

## Query

```
I want to understand the current architecture.
Walk me through what's documented and identify any gaps.
```

## Expect

### Discovery Process
- [ ] Loaded `.c3/README.md` first
- [ ] Attempted to read each container listed in Context
- [ ] Identified which containers exist vs missing
- [ ] Read existing container docs

---

## Output Criteria

### PASS: Identified Issues
| Issue Type | Should Detect |
|------------|---------------|
| Missing containers | c3-3 (Search) and c3-4 (Payment) folders don't exist |
| Inconsistent IDs | "OrderController" not following c3-xxx pattern |
| Stub container | c3-2 is stub with TODOs, missing required sections |
| Diagram drift | Diagram doesn't show all containers |
| Missing sections | c3-2 lacks Technology Stack, Components tables |

### PASS: Correct Summary Structure
| Element | Check |
|---------|-------|
| Container inventory | Listed all 4 containers from Context |
| Actual vs documented | Clearly distinguish what exists vs what's listed |
| Gaps prioritized | Most critical issues highlighted |
| Relationships | What's shown in diagram vs what's missing |

### FAIL: Missed Critical Issues
| Element | Failure Reason |
|---------|----------------|
| Didn't mention c3-3/c3-4 missing | Major gap - containers in Context but no docs |
| Didn't notice OrderController | Inconsistent with C3 ID patterns |
| Didn't flag c3-2 as incomplete | Missing required sections per v3-structure |

---

## Behavior Criteria

### PASS: Read-Only Exploration
- [ ] Did NOT create missing containers without asking
- [ ] Did NOT fix issues without permission
- [ ] Presented findings with recommendations
- [ ] Asked what to fix first

### FAIL: Unsolicited Changes
- [ ] Created c3-3 or c3-4 without confirmation
- [ ] "Fixed" OrderController ID without permission
- [ ] Added sections to c3-2 without asking
