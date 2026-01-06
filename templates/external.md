---
id: E${N}
c3-version: 3
title: ${EXTERNAL_NAME}
type: external
category: ${TYPE}
summary: ${SUMMARY}
---
<!-- TYPE: database | queue | cache | api | auth | storage -->

# ${EXTERNAL_NAME}

<!-- what this external provides to our system -->

## Contract Summary

**We depend on:** <!-- what we consume -->

**They expect:** <!-- what we must provide -->

**Failure mode:** <!-- what happens when unavailable -->

## Aspects (if complex)

<!-- SKIP IF: simple external, well-known service, thin wrapper -->

<!--
For complex externals, create aspect files:
- schema.md: data model, constraints, relationships
- access-patterns.md: how we query, indexes we rely on
- lifecycle.md: entity states, retention, archival
- workarounds.md: quirks, known issues, compensations
- versioning.md: API versions, migration notes
-->

| Aspect | File | Why Documented |
|--------|------|----------------|

## Linkages

```mermaid
graph LR
    %% Which containers connect to this external
    %% Edge labels: "purpose"
```

## Testing (if warranted)

<!-- SKIP IF: mocked in all tests, no integration tests needed -->

<!--
trivial: skip
simple: "Mock with in-memory store"
moderate: integration test approach, test data strategy
complex: + contract tests, chaos scenarios
-->
