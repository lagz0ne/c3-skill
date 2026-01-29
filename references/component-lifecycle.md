# Component Lifecycle

Components have a lifecycle status indicating their implementation state.

## Status Values

| Status | Meaning | Code References | Location |
|--------|---------|-----------------|----------|
| `provisioned` | Documented, not implemented | None | `.c3/provisioned/` |
| `active` | Implemented, in use | Required | `.c3/c3-X-*/` |
| `deprecated` | Being phased out | May have | `.c3/c3-X-*/` |

## File Organization

### Active Components

Live in standard container directories:

```
.c3/
  c3-1-frontend/
    README.md
    c3-101-component.md    # status: active
  c3-2-api/
    README.md
    c3-201-component.md    # status: active
```

### Provisioned Components

Live in parallel `.c3/provisioned/` directory:

```
.c3/
  provisioned/
    c3-1-frontend/
      c3-105-new-feature.md    # status: provisioned (new)
    c3-2-api/
      c3-201-component.md      # status: provisioned (planned update)
```

**Key insight:** Provisioned directory mirrors container structure. Same ID can exist in both locations - active version vs provisioned version.

## Frontmatter

### Provisioned Component (New)

```yaml
---
id: c3-205-rate-limiter
status: provisioned
adr: adr-20260129-rate-limiter
---
```

### Provisioned Component (Update to Existing)

```yaml
---
id: c3-201-auth
status: provisioned
supersedes: ../c3-2-api/c3-201-auth.md
adr: adr-20260129-oauth-support
---
```

### Active Component

```yaml
---
id: c3-201-auth
status: active
---
```

## Transitions

### Provision → Active (Implementation)

1. Create implementation ADR with `implements: <provisioned-adr>`
2. Execute implementation
3. Move file: `.c3/provisioned/c3-X/c3-XXX.md` → `.c3/c3-X/c3-XXX.md`
4. Update component status: `provisioned` → `active`
5. Add `## Code References` section
6. Update provisioned ADR: `superseded-by: <implementation-adr>`

### Active → Deprecated

1. Create deprecation ADR
2. Update component status: `active` → `deprecated`
3. Document replacement/migration path

## Audit Rules

| Rule | Check |
|------|-------|
| Provisioned has no Code References | `.c3/provisioned/**` files must NOT have `## Code References` |
| Active has Code References | `.c3/c3-*/**` files (except README) MUST have `## Code References` |
| Stale provisioned | Warn if active component changed after provisioned version created |
| Orphaned provisioned | Error if ADR is superseded but provisioned file not moved |
