## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "configuration scope mechanism blast radius sync subject prefix"
c3 read ref-scope-controlled-config --full
c3 graph ref-scope-controlled-config --depth 2 --direction reverse
c3 read ref-sync --full
c3 graph ref-sync --depth 2 --direction reverse
c3 read c3-101 --full
c3 read c3-209 --full
c3 read c3-4 --full
```

## Answer

Changing `NATS_SUBJECT_PREFIX` affects the sync subject contract. `ref-scope-controlled-config` says config should come from runtime scope tags, and `ref-sync` says the current default prefix is `sync`.

The immediate affected areas are:

| Entity | Impact |
| --- | --- |
| `ref-sync` | Defines `{prefix}.broadcast` and `{prefix}.user.{escaped_email}`. |
| `c3-101` | Frontend currently subscribes directly to `sync.broadcast` and `sync.user.<email>`, so this must change in lockstep. |
| `c3-209` | Credential generation uses `natsConfig.subjectPrefix` for allowed subscribe subjects. |
| `c3-4` | Broker permissions use `{prefix}.>` and must match the new prefix. |

Reverse graph is the right proof path here because subject prefix changes affect dependents rather than just the config ref itself. The main failure mode is mismatch: backend publishes one prefix while browser subscriptions or JWT permissions still use another.

## Grounding

`ref-scope-controlled-config` governs how `natsConfig.subjectPrefix` is injected. `ref-sync` governs NATS subject naming and client/server lockstep.

## Caveats

The fixture has known `c3 check` drift. I did not inspect source code beyond C3 docs.
