# UI-1

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
c3 graph ref-master-detail-layout --depth 1
c3 graph ref-detail-content-strategy --depth 1
```

## Answer

Keep invoice and payment request screens consistent by using the shared
Master-Detail pattern and its governing refs. `recipe-screen-anatomy` identifies
Invoices and Payment Requests as Master-Detail screens. `c3-104` owns
InvoiceScreen; `c3-105` owns PaymentRequestsScreen.

`ref-master-detail-layout` governs the outer layout: named slots, list plus
detail pane, desktop/tablet side-by-side sizing, mobile stacked slide behavior,
and layout-owned mobile back navigation. Screens should let the shared layout
own responsive behavior instead of adding custom per-screen mobile views.

`ref-detail-content-strategy` governs detail pane content: facet grids for
summary fields, BIG grids for inline key-value pairs, section order, tabbed
detail content, and empty-state treatment.

`ref-list-view-patterns` governs the list side: feature screens use virtualized
lists with sticky group headers, and PaymentRequestsScreen can use grouped or
flat list mode. `ref-filter-footer` governs filters through `FilterFooter` at
the list footer with active counts and compact stats. `ref-responsive-layout`
defines the mobile, tablet, and desktop breakpoints used by those patterns.

## Grounding

Search returned `recipe-screen-anatomy`, `c3-104`, `c3-105`,
`ref-master-detail-layout`, `ref-detail-content-strategy`, and
`ref-list-view-patterns`. `c3-104` names Master-detail, FilterFooter,
virtualized list, and tabbed detail pane. `c3-105` names MasterDetailLayout,
FilterFooter, PRFilterContent, and audit/detail wiring. The layout graph shows
both components cite `ref-master-detail-layout`.

## Caveats

No `rule-*` entities exist in this fixture. Use the component Governance,
Contract, and Change Safety rows with the refs' Goal/Choice/Why as the
constraint source.
