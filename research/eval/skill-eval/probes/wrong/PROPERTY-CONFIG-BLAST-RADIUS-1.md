# PROPERTY-CONFIG-BLAST-RADIUS-1 Wrong Probe

## Evidence Commands

- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 search "configuration scope mechanism blast radius sync subject prefix" # c3 search
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 graph ref-sync --depth 2 --direction reverse # c3 graph --direction reverse

## Answer

Changing `NATS_SUBJECT_PREFIX` away from `sync` is safe as a backend-only setting. `ref-scope-controlled-config` governs because any code can read the new env value lazily, so scope tags are only documentation and not a real dependency boundary.

`ref-sync` governs because the prefix only affects notifications, not app data sync. The reverse graph is useful mainly as a list of docs to ignore for this change: `c3-205`, `c3-206`, `c3-207`, `c3-210`, `c3-212`, `recipe-approval-workflow`, and `recipe-realtime-sync` do not need updates if the prefix changes.

The frontend follows automatically. `c3-101` subscribes to whatever `subjectPrefix` the server chooses, so there is no lockstep requirement. Keeping `sync.broadcast` and `sync.user` in the UI is fine even when `NATS_SUBJECT_PREFIX` changes, because those names are aliases.

Credential and broker permissions are unrelated. `c3-209` can keep issuing the old `nats_subject_prefix` permissions while `c3-4` can keep `{prefix}.>` at the default; NATS will route across both. `c3-211` and `recipe-backend-foundations` only need awareness if notification copy changes. `c3-202`, `c3-203`, and `c3-204` are not part of the actual impact scope.

This is a small blast radius and not a scope of impact problem: ref-scope-controlled-config, ref-sync, c3-202, c3-203, c3-204, c3-209, c3-211, recipe-backend-foundations, c3-101, c3-4, c3-205, c3-206, c3-207, c3-210, c3-212, recipe-approval-workflow, recipe-realtime-sync. Mechanism terms: reverse graph, nats_subject_prefix, subjectprefix, sync.broadcast, sync.user, lockstep. Also: c3 search, c3 graph, --direction reverse, component, ref, recipe, adr, contract, governance, goal/choice/why.
