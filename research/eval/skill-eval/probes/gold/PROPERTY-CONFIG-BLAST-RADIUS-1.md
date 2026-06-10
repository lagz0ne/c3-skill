# PROPERTY-CONFIG-BLAST-RADIUS-1 Gold Probe

## Evidence Commands

- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 search "configuration scope mechanism blast radius sync subject prefix" # c3 search
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read ref-scope-controlled-config --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 graph ref-scope-controlled-config --depth 2 --direction reverse # c3 graph
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read ref-sync --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 graph ref-sync --depth 2 --direction reverse
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-202 --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-203 --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-204 --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-209 --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-101 --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-4 --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-211 --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read recipe-backend-foundations --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read recipe-realtime-sync --full

## Answer

Changing `NATS_SUBJECT_PREFIX` away from `sync` is a shared contract change, not a local env rename. The action owner is the runtime configuration boundary described by `ref-scope-controlled-config`: config must be injected through scope tags such as `natsConfig.subjectPrefix`, and env vars are read only at scope creation. `c3-202` confirms `natsConfig` includes `subjectPrefix`; `c3-203` confirms middleware/bootstrap rewrites infrastructure env into standard runtime config before downstream code sees it.

Config state mutation: the prefix value changes the scope tag that downstream atoms and resources resolve. `ref-scope-controlled-config` governs because its Choice and Why reject closure-captured `process.env` and require config through scope tags. The direct reverse-graph dependents of this config ref are `c3-202`, `c3-203`, `c3-204`, `c3-209`, `c3-211`, and `recipe-backend-foundations`. Those are direct config-scope dependents, not all necessarily subject publishers.

Subject mechanism: `ref-sync` governs because it defines `{prefix}.broadcast` and `{prefix}.user.{escaped_email}`, states the current default prefix is `sync` from `NATS_SUBJECT_PREFIX`, and explicitly says frontend subscription wiring must change in lockstep if the prefix changes. `c3-101` confirms the frontend `natsSync` atom currently subscribes directly to `sync.broadcast` and `sync.user.<email>`, so the browser will not automatically follow a new prefix.

Permission mechanism: `c3-209` uses `natsConfig.subjectPrefix` to grant browser subscribe permissions for `{prefix}.broadcast` and `{prefix}.user.{escaped_email}`. `c3-4` is the external broker boundary: its required permissions use `{prefix}.>` with default prefix `sync`. Therefore credential generation, broker permissions, and frontend subscriptions all have to agree on the same prefix.

Direct sync dependents from the reverse graph for `ref-sync`: `c3-101`, `c3-205`, `c3-206`, `c3-207`, `c3-210`, `c3-212`, `recipe-approval-workflow`, and `recipe-realtime-sync`. These are affected because they rely on the real-time sync contract, even when they do not own the config tag itself. `c3-205` covers PR flow sync, `c3-206` invoice flow sync, `c3-207` payment flow sync, `c3-210` admin flow sync, and `c3-212` workbench flow sync by graph evidence.

Indirect/supporting boundaries:

- `c3-4` is not shown as a direct `ref-sync` dependent in the reverse graph, but its read output proves the broker permission boundary for `{prefix}.>`.
- `c3-211` is a direct config-scope dependent and owns notifications; `recipe-realtime-sync` adds the caveat that broadcast sync and notifications share NATS but remain architecturally separate.
- Historical ADRs in the reverse graphs, such as NATS websocket/auth and notification ADRs, are historical context. The current change should be driven by the active refs and component contracts above.

Emergent property: this is a blast radius across config, permissions, broker routing, frontend subscriptions, sync flows, and notification subjects. A prefix change can be syntactically small but behaviorally wide because `subjectPrefix`, `sync.broadcast`, `sync.user`, JWT permissions, broker permissions, and browser subscriptions must stay in lockstep.

Failure boundary:

- If only backend env changes, `c3-101` can stay subscribed to old `sync.*` subjects and miss deltas/notifications.
- If `c3-209` and `c3-4` disagree, browser JWT credentials can be validly signed but lack matching subscribe permissions at the broker.
- If `ref-sync` flows publish under a new prefix while the client waits under the old prefix, `executionTracker` will only resolve by timeout fallback or not receive the intended update promptly.

Concrete change checks:

- Update scope tag injection for `NATS_SUBJECT_PREFIX` and test `ref-scope-controlled-config` behavior with a non-`sync` value.
- Update and test `c3-101` subscriptions for both `sync.broadcast` and `sync.user` equivalent subjects under the new prefix.
- Test `c3-209` JWT permissions and `c3-4` broker permissions with the same prefix.
- Run a PR mutation (`c3-205`), invoice mutation (`c3-206`), payment mutation (`c3-207`), admin mutation (`c3-210`), and workbench mutation (`c3-212`) to prove ref-sync dependents still broadcast and ack.
