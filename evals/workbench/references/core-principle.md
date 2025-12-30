# The C3 Principle

> **Upper layer defines WHAT. Lower layer implements HOW.**

This is the only rule that matters. Everything else follows from it.

## The Contract Chain

```
Context   →  defines WHAT containers exist and WHY
              ↓ implemented by
Container →  defines WHAT components exist and WHAT they do
              ↓ implemented by
Component →  defines HOW it works
              ↓ implemented by
Code      →  (lives in codebase, not in C3 docs)
```

## Layer Integrity

Each layer has exactly one job:

| Layer | Defines | Does NOT Define |
|-------|---------|-----------------|
| **Context** | WHY containers exist, relationships | What's inside containers |
| **Container** | WHAT components do, relationships | How components work internally |
| **Component** | HOW it implements its contract | What it's responsible for (that's Container's job) |

## The Integrity Rule

> **You cannot document a layer without the layer above.**

- Cannot write Component doc if Container doesn't list it
- Cannot write Container doc if Context doesn't list it
- Lower layer implements upper layer's contract, not its own invention

## Applying the Principle

**When documenting anything, ask:**

1. "What layer am I at?"
2. "What does the layer above say about this?"
3. "Am I documenting HOW (correct) or inventing WHAT (wrong)?"

**When something doesn't fit:**

- If it's about WHY containers exist → Context
- If it's about WHAT components do → Container
- If it's about HOW it works → Component
- If it's code → Codebase (not C3)

## The Principle in Action

**Wrong (Component invents responsibility):**
```
# UserService Component
Handles user registration, authentication, and profiles.  ← WRONG: this is WHAT, belongs in Container
```

**Right (Component implements contract):**
```
# UserService Component
Contract: From Container - "Handles user registration, authentication, and profiles"
How it works: [flow, dependencies, edge cases]  ← RIGHT: this is HOW
```

## Summary

```
Upper defines WHAT → Lower implements HOW
Context → Container → Component → Code
Integrity: lower cannot exist without upper
```

This principle is non-negotiable. Archetypes, patterns, and guidance are just applications of it.
