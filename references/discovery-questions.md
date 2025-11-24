# Discovery Questions

Questions to ask during Socratic discovery. Use AskUserQuestion tool when presenting choices.

## Philosophy

1. **Ask to understand**, not to enforce
2. **Accept "we don't"** as valid answer (document as TBD)
3. **Explore before asking** - use Task Explore first
4. **Present choices** when options are clear (use AskUserQuestion)
5. **Open-ended** when context is needed

## Container-Level Questions

### Identity
- "What is this container's primary responsibility?"
- "Who/what calls this container?"
- "What would break if this container was removed?"

### Technology (use AskUserQuestion for choices)
- "What runtime/framework? [discovered options from exploration]"
- "What language version?"
- "Any significant libraries?"

### Structure
- "How is the code organized?"
- "What are the main entry points?"
- "Any distinct layers or modules?"

### Dependencies
- "What databases/caches does this use?"
- "What other services does this call?"
- "Any external APIs?"

### Testing
- "How is this container tested?"
- "What runs in CI?"
- "Any manual testing steps?"

## Component-Level Questions

### Purpose
- "What does this component do?"
- "Why does it exist separately from [adjacent component]?"
- "What would you lose if this was merged into its caller?"

### Implementation
- "What's the main pattern here? [service, repository, handler, etc.]"
- "Any complex algorithms or business rules?"
- "How are errors handled?"

### Configuration
- "What configuration does this need?"
- "Any environment-specific behavior?"
- "Secrets or sensitive values?"

### Dependencies
- "What does this component depend on?"
- "What depends on this component?"
- "Any circular dependencies?"

## Context-Level Questions

### System Boundary
- "What is this system responsible for?"
- "What is explicitly NOT in scope?"
- "Who are the users/actors?"

### External Systems
- "What external systems does this integrate with?"
- "Who owns those systems?"
- "What happens when they're unavailable?"

### Cross-Cutting Concerns
- "How is authentication handled?"
- "How is logging/monitoring done?"
- "Any shared infrastructure?"

## Platform Questions

### Deployment
- "How are containers deployed?"
- "What orchestrator (K8s, ECS, etc.)?"
- "Rollback strategy?"

### Networking
- "Network topology?"
- "Service discovery mechanism?"
- "Any service mesh?"

### Secrets
- "Where are secrets stored?"
- "How are they injected?"
- "Rotation policy?"

### CI/CD
- "What triggers deployments?"
- "Pipeline stages?"
- "Approval gates?"

## When User Doesn't Know

If the user doesn't know the answer:
1. Mark as `TBD` in documentation
2. Move on - don't block on unknowns
3. Note it as something to discover later
4. Don't guess or assume

Example:
```markdown
## Testing
TBD - testing strategy not yet established
```
