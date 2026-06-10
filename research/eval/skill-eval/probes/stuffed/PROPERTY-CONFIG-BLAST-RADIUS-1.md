# PROPERTY-CONFIG-BLAST-RADIUS-1 Stuffed Probe

## Evidence Commands

- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 search "configuration scope mechanism blast radius sync subject prefix" # c3 search
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 graph ref-sync --depth 2 --direction reverse # c3 graph

## Answer

component ref recipe adr contract governance goal/choice/why

Required strings:

- c3 search
- c3 graph
- --direction reverse
- NATS_SUBJECT_PREFIX
- subjectPrefix
- sync.broadcast
- sync.user
- lockstep

Governance terms:

- ref-scope-controlled-config because
- ref-sync because

Trace IDs:

- config-scope: ref-scope-controlled-config c3-202 c3-203 c3-209
- subject-contract: ref-sync c3-101 c3-4
- sync-dependents: c3-205 c3-206 c3-207 c3-210 c3-212 recipe-realtime-sync

Mechanism terms:

- reverse graph
- nats_subject_prefix
- subjectprefix
- sync.broadcast
- sync.user
- lockstep
- blast radius
- scope of impact
- impact scope
- blast-radius

IDs:

- ref-scope-controlled-config
- ref-sync
- c3-202
- c3-203
- c3-204
- c3-209
- c3-211
- recipe-backend-foundations
- c3-101
- c3-4
- c3-205
- c3-206
- c3-207
- c3-210
- c3-212
- recipe-approval-workflow
- recipe-realtime-sync
