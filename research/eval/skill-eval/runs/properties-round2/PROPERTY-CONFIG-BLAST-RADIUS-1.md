## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "configuration scope mechanism blast radius sync subject prefix"
c3 read ref-scope-controlled-config --full
c3 graph ref-scope-controlled-config --depth 2 --direction reverse
c3 read ref-sync --full
c3 graph ref-sync --depth 2 --direction reverse
c3 read c3-202 --full
c3 read c3-203 --full
c3 read c3-209 --full
c3 read c3-101 --full
c3 read c3-4 --full
c3 read c3-211 --full
c3 read recipe-realtime-sync --full
```

## Answer

Changing `NATS_SUBJECT_PREFIX` has a broad **blast radius** because it is a scope-controlled config value that feeds the real-time sync subject contract.

Config blast radius from the reverse graph:

| Entity | Impact |
| --- | --- |
| `ref-scope-controlled-config` | Governs because config must come from runtime scope tags, not hidden env closure captures. |
| `c3-202` | Defines infrastructure config tags, including NATS config. |
| `c3-203` | Bootstraps env-derived NATS URLs before downstream config validation. |
| `c3-204` | Depends on the same scope config pattern for backend infrastructure config. |
| `c3-209` | Uses `natsConfig.subjectPrefix` when generating JWT subscribe permissions. |
| `c3-211` | Uses scope-controlled config for notification delivery paths. |
| `recipe-backend-foundations` | Captures the backend scope-config contract. |

Sync blast radius from `ref-sync` reverse graph:

| Entity | Impact |
| --- | --- |
| `ref-sync` | Governs because it defines `{prefix}.broadcast`, `{prefix}.user.{escaped_email}`, and default `sync` from `NATS_SUBJECT_PREFIX`. |
| `c3-101` | Subscribes directly to `sync.broadcast` and `sync.user.<email>`; frontend wiring must change in lockstep. |
| `c3-4` | Broker permissions use `{prefix}.>`; NATS pub/sub permissions must match the new prefix. |
| `c3-205`, `c3-206`, `c3-207`, `c3-210`, `c3-212` | Backend flows depend on `ref-sync` to publish mutation deltas/acks. |
| `recipe-approval-workflow`, `recipe-realtime-sync` | Recipes depend on the sync subject contract and execution path. |

Mechanism: the subjectPrefix / `NATS_SUBJECT_PREFIX` value has to match across backend config tags, JWT credential permissions, `c3-4` broker permissions, server publish subjects, and `c3-101` subscriptions to `sync.broadcast` and `sync.user.<email>`. Scope of impact: a mismatch can make flows mutate data but clients never receive NATS sync or user-subject messages.

## Grounding

`ref-scope-controlled-config` governs runtime config access. `ref-sync` governs because it defines the subject prefix contract and lockstep frontend/backend behavior.

## Caveats

The fixture has known `c3 check` drift. No `rule-*` entities found in the fixture.
