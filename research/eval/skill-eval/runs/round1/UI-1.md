# UI-1

## Evidence commands

```bash
c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }
c3 list
c3 check
c3 search "How should invoice and payment request screens stay consistent across detail and list layouts?"
c3 read recipe-screen-anatomy --full
c3 read c3-104 --full
c3 read c3-105 --full
c3 read ref-master-detail-layout --full
c3 read ref-detail-content-strategy --full
c3 read ref-list-view-patterns --full
c3 read ref-filter-footer --full
c3 read ref-responsive-layout --full
c3 graph ref-master-detail-layout --depth 1
c3 graph ref-detail-content-strategy --depth 1
```

## Answer

Use the shared Master-Detail screen pattern, not per-screen layout invention.
`recipe-screen-anatomy` identifies both Invoices and Payment Requests as
Master-Detail screens. `c3-104` owns InvoiceScreen and `c3-105` owns
PaymentRequestsScreen.

`ref-master-detail-layout` governs the outer structure: list plus detail pane,
named slots, desktop and tablet side-by-side behavior, mobile stacked slide
behavior, and layout-owned mobile back navigation. This keeps invoice and PR
screens consistent across viewport sizes.

`ref-detail-content-strategy` governs detail panes. Use facet grids for summary
fields, BIG grids for inline key-value pairs, ordered sections, tabbed detail
content, and plain empty states. This preserves consistent detail hierarchy
across invoice and payment request detail views.

`ref-list-view-patterns` governs collection display. Feature screens like
InvoiceScreen and PaymentRequestsScreen use virtualized lists with sticky group
headers; PR screens may switch between flat list and grouped mode.

`ref-filter-footer` governs list filtering. Filters live in `FilterFooter` at
the list footer, with active filter counts and compact stats. `ref-responsive-layout`
governs the mobile, tablet, and desktop breakpoints used by the shared layout.

## Grounding

`c3-104` says InvoiceScreen uses Master-detail with FilterFooter, a virtualized
list on the left, and a tabbed detail pane on the right. `c3-105` names
MasterDetailLayout, FilterFooter, PRFilterContent, and the approval/detail
screen wiring. The `ref-master-detail-layout` graph shows both `c3-104` and
`c3-105` cite it.

## Caveats

No `rule-*` entities exist in this fixture. The governing source is the refs'
Goal/Choice/Why plus the component Governance, Contract, and Change Safety rows.
