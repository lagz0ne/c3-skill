# Component Template (c3-NNN)

```markdown
---
id: c3-{N}{NN}
c3-version: 3
title: [Component Name]
type: component
parent: c3-{N}
summary: >
  [One-line description]
---

# [Component Name]

## Contract

From c3-{N} ([Container Name]): "[responsibility from Container inventory]"

## Interface

` ` `mermaid
flowchart LR
    subgraph IN["Receives"]
        I1[Input 1]
        I2[Input 2]
    end

    subgraph PROCESS["Owns"]
        P1[Step 1]
        P2[Step 2]
    end

    subgraph OUT["Provides"]
        O1[Output 1]
    end

    IN --> PROCESS --> OUT
` ` `

## Hand-offs

| Direction | What | To/From |
|-----------|------|---------|
| IN | [data/request] | [source component] |
| OUT | [result/response] | [target component] |

## Conventions

| Rule | Why |
|------|-----|
| [pattern consumers must follow] | [reason] |

## Edge Cases & Errors

| Scenario | Behavior |
|----------|----------|
| [what can go wrong] | [how it's handled] |
```

---

## Notes

- **Contract first:** Reference what Container says about this component
- **Diagram required:** Interface diagram showing IN/PROCESS/OUT
- **No code:** Use tables instead of snippets
- **Conventions focus:** Why this doc exists is to document rules for consumers
