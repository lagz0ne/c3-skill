# Migration: 20251124-settings-yaml

> From: `20251124-adr-date-naming`
> To: `20251124-settings-yaml`

## Changes

- Add `.c3/settings.yaml` for project preferences
- New sections: diagrams, context, container, component, guard, handoff, audit
- c3-adopt now calls c3-config if settings missing
- c3-design reads settings at start

## Transforms

### Create settings.yaml

**If not exists:** `.c3/settings.yaml`

Create with defaults:
```yaml
diagrams: |
  mermaid
  sequence: API interactions, request flows
  flowchart: decision logic, error handling

context: |
  system boundaries, actors, external integrations
  avoid implementation details

container: |
  service responsibilities, API contracts
  backend: endpoints, data flows
  frontend: component hierarchy, state

component: |
  technical specifics, configs, algorithms

guard: |
  discovered incrementally via c3-config

handoff: |
  after ADR accepted:
  1. create implementation tasks
  2. notify team

audit: |
  handoff: tasks
```

No transforms needed for existing files - settings.yaml is additive.

## Verification

```bash
# settings.yaml exists
ls .c3/settings.yaml

# Has expected sections
grep -q '^diagrams:' .c3/settings.yaml
grep -q '^context:' .c3/settings.yaml
grep -q '^container:' .c3/settings.yaml
grep -q '^component:' .c3/settings.yaml
grep -q '^guard:' .c3/settings.yaml
grep -q '^handoff:' .c3/settings.yaml
grep -q '^audit:' .c3/settings.yaml
```
