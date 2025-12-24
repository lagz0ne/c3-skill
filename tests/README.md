# C3 Skill Test Cases

Test cases for verifying c3 agent behavior in **isolation** with **full audit trail**.

## Key Principles

1. **Natural behavior**: Agent discovers and uses skills on its own (not force-fed)
2. **Audit**: Every tool call traced (READ, WRITE, SKILL, BASH, etc.)
3. **Verification**: Explicit PASS/FAIL criteria per layer
4. **Isolation**: Workspace-only access verified

## Structure

```
tests/
├── run-test.sh       # Test runner (isolation + audit)
├── verify-output.sh  # Content verification
├── audit/
│   ├── hooks.json    # Audit hooks config
│   ├── log-access.sh # Logs all tool calls
│   └── analyze.sh    # Analyze audit log
├── cases/            # Test case definitions
│   ├── 01-new-project.md
│   ├── 02-explore-architecture.md
│   ├── 03-add-container.md
│   ├── 04-add-component.md
│   └── 05-create-adr.md
└── output/           # Test outputs + audit logs (gitignored)
```

## Running Tests

### Run a test

```bash
# Run with audit (default)
./tests/run-test.sh 01-new-project

# Run with specific workspace
./tests/run-test.sh 03-add-container --workspace /tmp/my-project

# Run without audit
./tests/run-test.sh 01-new-project --no-audit
```

### Verify results

```bash
# Check workspace content against expectations
./tests/verify-output.sh 01-new-project /tmp/workspace-path
```

### Analyze audit log

```bash
# View audit log
cat tests/output/01-new-project-TIMESTAMP.audit.log

# Analyze access patterns
./tests/audit/analyze.sh tests/output/01-new-project-TIMESTAMP.audit.log

# Strict mode - fail if any access outside allowed paths
./tests/audit/analyze.sh tests/output/*.audit.log --strict
```

### Manual run (for debugging)

```bash
cd /tmp/test-workspace
claude --no-plugins --plugin /path/to/c3-design
```

## Audit System

Every test run generates an audit log tracking:

| Operation | Logged |
|-----------|--------|
| `READ` | File path read |
| `WRITE` | File path written |
| `EDIT` | File path edited |
| `GLOB` | Pattern + path searched |
| `GREP` | Pattern + path searched |
| `BASH` | Command executed |

### Audit Categories

| Category | Allowed |
|----------|---------|
| Workspace files | Yes - test subject |
| Plugin skill files | Yes - skill being tested |
| Plugin references | Yes - skill dependencies |
| Home directory | No - isolation violation |
| System files | No - isolation violation |
| Other plugins | No - isolation violation |

### Example Audit Output

```
Available skill files:
  - agents/c3.md (322 lines)
  - skills/c3-context-design/SKILL.md (238 lines)
  ...

[16:06:01] SKILL: c3-skill:c3-adopt
[16:06:06] BASH: ls -la .c3/ 2>/dev/null
[16:06:19] SKILL: c3-skill:c3-context-design
[16:06:49] WRITE: /tmp/workspace/.c3/README.md
[16:07:00] SKILL: c3-skill:c3-container-design
[16:07:05] READ: /tmp/workspace/.c3/README.md
[16:07:41] WRITE: /tmp/workspace/.c3/c3-1-api/README.md
```

**Key insight:** Agent discovers skills via Skill tool (plugin registration), not by reading SKILL.md files directly.

## Test Case Format

Each test case has:

```markdown
# Test: Name

## Setup (optional)
What needs to exist before test runs.

## Query
```
The exact prompt to send to the agent.
```

## Expect

### PASS: Should Include
| Element | Check |
|---------|-------|
| ... | What must be present |

### FAIL: Should NOT Include
| Element | Failure Reason |
|---------|----------------|
| ... | What indicates wrong layer/violation |
```

## Layer Criteria Summary

### Context (c3-0)
| PASS | FAIL |
|------|------|
| Container inventory | Component details |
| WHY containers exist | Tech stack specifics |
| Container relationships | HOW things work |
| External actors | Code of any kind |

### Container (c3-N)
| PASS | FAIL |
|------|------|
| Component inventory | WHY container exists |
| WHAT components do | HOW components work |
| Tech stack table | Implementation logic |
| Component relationships | Code of any kind |

### Component (c3-NNN)
| PASS | FAIL |
|------|------|
| HOW it implements | WHAT it does (Container's job) |
| Interface diagram | Component relationships |
| Hand-offs table | System context diagrams |
| Edge cases/errors | Code blocks (except Mermaid) |

## Test Categories

| Category | Test | Validates |
|----------|------|-----------|
| Adoption | `01-new-project` | Socratic discovery, structure creation, layer placement |
| Exploration | `02-explore-architecture` | Read-only discovery, gap identification |
| Container | `03-add-container` | Context update, foundation-first, container content |
| Component | `04-add-component` | Parent check, NO CODE rule, component content |
| ADR | `05-create-adr` | Layer impact, cross-references, verification |
