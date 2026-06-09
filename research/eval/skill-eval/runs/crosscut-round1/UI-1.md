## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 search "How should invoice and payment request screens stay consistent across detail and list layouts?"
c3 read recipe-screen-anatomy --full
c3 read c3-104 --full
c3 read c3-105 --full
c3 read ref-master-detail-layout --full
c3 read ref-detail-content-strategy --full
c3 read ref-list-view-patterns --full
c3 read ref-filter-footer --full
c3 read ref-responsive-layout --full
c3 graph c3-105 --depth 1
```

## Answer

`recipe-screen-anatomy` orients both screens. `c3-104` owns InvoiceScreen and `c3-105` owns PaymentRequestsScreen. They should stay consistent by using Master-Detail layout, virtualized/filterable list behavior, detail panes with facet/BIG-grid style content, and shared responsive behavior.

## Grounding

`ref-master-detail-layout` governs because both screens use the same list/detail shell. `ref-detail-content-strategy` governs detail panes because invoice and PR details compose the same content strategy. `ref-list-view-patterns` governs row/list behavior, `ref-filter-footer` governs filtering controls, and `ref-responsive-layout` governs mobile/desktop layout rather than per-screen custom behavior.

## Caveats

No `rule-layout` entity exists.
