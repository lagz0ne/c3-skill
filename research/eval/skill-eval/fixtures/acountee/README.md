# Acountee fixture

Portable snapshot of `~/dev/acountee/.c3` captured on 2026-06-09 for the C3
skill-effectiveness eval.

This is a real but imperfect base snapshot. It intentionally preserves current
documentation drift so the eval measures how agents answer against real C3
material, not a cleaned-up demo.

Included:

- canonical `.c3` markdown entities
- `.c3/code-map.yaml`

Excluded:

- `.c3/c3.db` cache

Before running `search`, `read`, `lookup`, or `graph` against this fixture, run:

```bash
C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 check || true
```

Remove any regenerated `fixtures/acountee/.c3/c3.db` before committing.
