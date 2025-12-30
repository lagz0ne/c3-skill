# Test: Discovery-Based Adopt

## Setup

Fixture: `fixtures/06-discovery-adopt/`

```
apps/
  backend/
    src/
      main.ts
      auth/
        index.ts
      api/
        routes.ts
    package.json
    Dockerfile
  frontend/
    src/
      index.tsx
      components/
    package.json
docker-compose.yml
```

## Query

```
Set up C3 documentation for this project.
```

## Expect

### PASS: Discovery Should Find

| Element | Check |
|---------|-------|
| 2 containers detected | backend, frontend |
| Components in backend | auth, api |
| External actor | detected from docker-compose |

### PASS: User Interaction

| Element | Check |
|---------|-------|
| AskUserQuestion called | For container confirmation |
| AskUserQuestion called | For component confirmation |

### PASS: Output Structure

| Element | Check |
|---------|-------|
| .c3/README.md | Has container inventory |
| .c3/c3-1-backend/README.md | Has component inventory |
| .c3/c3-2-frontend/README.md | Has component inventory |
| No component docs | Inventory-first respected |

### FAIL: Should NOT Include

| Element | Failure Reason |
|---------|----------------|
| Component .md files | Should be inventory-first |
| Code blocks in docs | NO CODE rule |
