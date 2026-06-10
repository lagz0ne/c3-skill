## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "Slack approve reject payment request notification sync"
c3 read c3-215 --full
c3 read adr-20260305-slack-bot-integration --full
c3 read c3-205 --full
c3 read c3-211 --full
c3 read ref-pull-dispatcher --full
c3 read ref-sync --full
c3 graph c3-215 --depth 1
c3 graph c3-205 --depth 1
```

## Answer

Slack approval enters through `c3-215`. The Slack action resolves the Slack user, maps it to an Acountee user, then calls `c3-205` PR flows. `adr-20260305-slack-bot-integration` says Slack actions call flows, not services. That keeps the same PR mutation path as web server functions.

Web clients stay consistent through `ref-sync`: PR flows send sync acks/deltas over NATS WebSocket, and `c3-101` receives them. Next approvers stay consistent through `c3-211`, with channels registered through `ref-pull-dispatcher`; Slack's outbound notification path is `notificationDispatcher -> slackChannel -> DM card`.

## Grounding

`ref-sync` governs because the browser state update depends on NATS sync messages. `ref-pull-dispatcher` governs because Slack is a channel that subscribes to notification dispatch.

## Caveats

This answer does not inspect Slack runtime config or actual webhook logs.
