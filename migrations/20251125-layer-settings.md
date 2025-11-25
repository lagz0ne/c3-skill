# Migration: 20251125-layer-settings

> From: `20251124-toc-fix-links`
> To: `20251125-layer-settings`

## Changes

- Layer sections in settings.yaml now support structured configuration
- New keys per layer: `useDefaults`, `include`, `exclude`, `litmus`, `diagrams`
- Existing prose-only layer config (e.g., `context: |`) still works (backward compatible)
- Layer design skills now read defaults from co-located `defaults.md` files

## Transforms

### Upgrade settings.yaml (Optional)

**If exists:** `.c3/settings.yaml`

Users MAY upgrade from prose-only format to structured format:

**Before (still works):**
```yaml
context: |
  system boundaries, actors, external integrations
  avoid implementation details
```

**After (new capabilities):**
```yaml
context:
  useDefaults: true
  guidance: |
    system boundaries, actors, external integrations
    avoid implementation details
  include: |
    # Optional: add items to defaults
  exclude: |
    # Optional: add items to defaults
  litmus: |
    # Optional: override default litmus test
```

**No automatic transforms needed** - the change is backward compatible.

Existing prose-only format continues to work; users can opt-in to structured format when they want customization.

## Verification

```bash
# VERSION updated
cat .c3/README.md | grep -q 'c3-version: 20251125-layer-settings' && echo "VERSION: OK"

# settings.yaml still valid (if exists)
if [ -f ".c3/settings.yaml" ]; then
  grep -q '^context:' .c3/settings.yaml && echo "context section: OK"
  grep -q '^container:' .c3/settings.yaml && echo "container section: OK"
  grep -q '^component:' .c3/settings.yaml && echo "component section: OK"
fi
```

## Backward Compatibility

| Old Format | New Format | Works? |
|------------|------------|--------|
| `context: \|` (prose) | `context:` (structured) | ✅ Yes, prose still valid |
| No settings.yaml | No settings.yaml | ✅ Yes, skills use defaults.md |

Users only need to migrate if they want the new customization features.
