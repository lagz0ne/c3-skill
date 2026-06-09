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

`recipe-screen-anatomy` maps the shared screen structure. `c3-104` owns InvoiceScreen and `c3-105` owns PaymentRequestsScreen. Both should keep Master-Detail structure, virtualized list rows, `FilterFooter` filtering, detail pane sections, and responsive behavior aligned.

## Grounding

`ref-master-detail-layout` governs because both screens use list/detail navigation. `ref-detail-content-strategy` governs because detail panes need consistent sections. `ref-list-view-patterns` governs because rows and virtualized list behavior should stay common. `ref-filter-footer` governs because filters sit in the shared footer pattern. `ref-responsive-layout` governs because mobile/desktop adaptation belongs to the shared layout pattern.

## Caveats

Fixture has no rule entities.
