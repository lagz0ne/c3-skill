# Test: New Project Adoption

## Query

```
I have a new project - a REST API backend for a task management app.
It uses Node.js with Express, PostgreSQL for storage, and Redis for caching.
Help me set up C3 documentation.
```

## Expect

### Structure Created
- [ ] `.c3/README.md` exists with valid frontmatter
- [ ] At least one container folder `.c3/c3-1-*/` created
- [ ] Container has `README.md` with valid frontmatter

---

## Context Level (c3-0) Criteria

### PASS: Should Include
| Element | Check |
|---------|-------|
| Container inventory table | Lists containers with IDs and purposes |
| WHY containers exist | Explains the problem each container solves |
| Container relationships | How containers connect/depend on each other |
| External actors | Users, external services interacting with system |
| System boundary | Clear inside vs outside distinction |
| Connecting points | APIs, events, data flows between containers |

### FAIL: Should NOT Include
| Element | Failure Reason |
|---------|----------------|
| Component list | Push to Container - internal structure |
| Express/Redis/PostgreSQL details | Push to Container - tech stack |
| How request handling works | Push to Container/Component - internal |
| Code snippets | Never in C3 docs |
| File paths or folder structure | Push to auxiliary docs |

---

## Container Level (c3-1) Criteria

### PASS: Should Include
| Element | Check |
|---------|-------|
| Component inventory table | Lists components with IDs and responsibilities |
| WHAT each component does | Role/responsibility (not HOW) |
| Tech stack table | Express, PostgreSQL, Redis with purposes |
| Component relationships | How components call/depend on each other |
| Data flows | How data moves across components |
| Inner patterns | Shared logging, config, error handling approaches |
| Foundational aspects identified | Entry point, request pipeline, config, etc. |

### FAIL: Should NOT Include
| Element | Failure Reason |
|---------|----------------|
| WHY this container exists | Push to Context - already there |
| Container-to-container communication details | Push to Context |
| HOW components work internally | Push to Component |
| Implementation logic | Push to Component |
| Code snippets | Never in C3 docs |
| Middleware implementation details | Push to Component |

---

## Socratic Process Criteria

### PASS: Discovery Happened
- [ ] Asked about purpose/scope before documenting
- [ ] Explored actors and system boundaries
- [ ] Identified container responsibilities through questions
- [ ] Asked about foundational aspects (entry point, config, logging)

### FAIL: Skipped Discovery
- [ ] Jumped straight to creating files
- [ ] Made assumptions without asking
- [ ] Used generic template without tailoring
