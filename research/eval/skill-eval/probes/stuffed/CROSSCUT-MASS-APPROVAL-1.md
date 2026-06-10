# CROSSCUT-MASS-APPROVAL-1 Stuffed Probe

## Evidence Commands

- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 search "bulk approve payment requests notify approvers sync" # c3 search
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 graph c3-205 --depth 1 # c3 graph

## Answer

component ref recipe adr contract governance goal/choice/why

Required strings:

- c3 search
- c3 graph
- approveAll
- non-blocking
- async
- error suppression
- logged, not thrown

Governance terms:

- ref-approval-chain because
- ref-sync because

Trace IDs:

- action: c3-105 c3-205
- sync: ref-sync c3-101
- notification: c3-211 adr-20260121-notification-system

Mechanism terms:

- nats websocket ref-sync
- c3-211 jetstream notification

IDs:

- c3-105
- c3-205
- ref-approval-chain
- ref-sync
- c3-101
- c3-211
- adr-20260121-notification-system
- adr-20260202-notification-on-step-advance
