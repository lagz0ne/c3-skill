# Living Entity Experiment

Turn a C3 architecture graph into a team of self-aware Claude Code agents.

## Concept

After C3 onboarding produces a `.c3/` directory with components, containers, refs, and ADRs,
we generate one agent per entity. Each agent:

1. **Knows its code** - owns specific file paths, verified through inspection
2. **Enforces refs** - behavioral contracts (error handling, testing, delegation patterns)
3. **Understands relationships** - dependencies, boundaries, who calls whom
4. **Remembers decisions** - ADRs that constrain future changes

An orchestrator routes change requests to the relevant agent(s), which advise on impact.

## Usage

```bash
# Generate agents from a C3-documented project
bun experiments/living-entity/generate.ts /path/to/project/.c3/

# Output goes to experiments/living-entity/generated/agents/
```

## Architecture

```
User: "Add retry logic to payment processing"
  → Orchestrator: identifies c3-207 (Payment Flows)
  → c3-207 agent checks:
     - Code ownership: src/server/flows/payment*.ts
     - ref-query-services: must use service() pattern
     - ref-sync: must broadcast changes via NATS
     - Depends on: c3-204 (Drizzle), c3-201 (Flow Pattern)
  → Advisory: structured guidance with ALL constraints
```

## Constraint Chain

Each entity-agent enforces four layers:

| Layer | Source | Purpose |
|-------|--------|---------|
| Code ownership | Component doc → code paths | Guard what files this component owns |
| Behavioral refs | Applicable refs | How code must be written, tested, handled |
| Relationships | Container/component graph | Who depends on me, who I depend on |
| ADR history | Linked ADRs | Past decisions that constrain future changes |
