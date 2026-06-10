## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "Slack approve reject payment request notification sync"
c3 read c3-215 --full
c3 read adr-20260305-slack-bot-integration --full
c3 read c3-202 --full
c3 read c3-205 --full
c3 read ref-sync --full
c3 read c3-101 --full
c3 read c3-211 --full
c3 read ref-pull-dispatcher --full
c3 graph c3-215 --depth 1
c3 graph c3-205 --depth 1
c3 graph c3-211 --depth 1
```

## Answer

`c3-215` owns Slack inbound actions. `adr-20260305-slack-bot-integration` says the `approve_pr` / `reject_pr` handlers resolve a Slack user, build execution context using `c3-202` concepts, then call `c3-205` PR flows, not services. That keeps approval state mutation inside the same flow path.

Web clients stay current through `ref-sync`: NATS WebSocket deltas/acks are received by `c3-101`. Next approvers stay informed through `c3-211`; Slack is a notification channel registered through `ref-pull-dispatcher`. Emergent property: entering via flows preserves sync and notification side effects that a direct service call would bypass.

## Grounding

`ref-sync` governs because browser state consistency depends on flow acks/deltas. `ref-pull-dispatcher` governs because notification channels are registered subscribers.

## Caveats

Slack runtime config can still disable delivery, but the architectural path is flow-based.
