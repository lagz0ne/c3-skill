# UI-1 Gold Probe

## Evidence Commands

- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 search "How should invoice and payment request screens stay consistent across detail and list layouts?" # c3 search
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read recipe-screen-anatomy --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-104 --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-105 --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read ref-master-detail-layout --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read ref-detail-content-strategy --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read ref-list-view-patterns --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read ref-filter-footer --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read ref-responsive-layout --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 graph ref-master-detail-layout --depth 1 --direction reverse
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 graph ref-detail-content-strategy --depth 1 --direction reverse
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 graph c3-104 --depth 1
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 graph c3-105 --depth 1

## Answer

The consistency owner is shared pattern governance, not local per-screen invention. `recipe-screen-anatomy` classifies both Invoices and Payment Requests as `Master-Detail` screens: Invoices route to `c3-104`, Payment Requests route to `c3-105`, and both use list plus detail pane composition.

Action owner: when an engineer changes invoice layout, `c3-104` owns the screen behavior; when changing PR or approval layout, `c3-105` owns it. Those components are direct owners, and their graphs show the same UI pattern refs: `ref-master-detail-layout`, `ref-detail-content-strategy`, `ref-list-view-patterns`, `ref-filter-footer`, and `ref-responsive-layout`.

State and layout mechanism: `ref-master-detail-layout` governs because it defines the `MasterDetailLayout` slot contract, desktop/tablet side-by-side behavior, mobile stacked slide behavior, and internal mobile back navigation. `ref-responsive-layout` governs because it defines the mobile, tablet, and desktop breakpoints and repeats the MasterDetailLayout behavior for each tier. Together they mean screen code should let the shared layout manage responsive transitions instead of creating one-off mobile views.

List mechanism: `ref-list-view-patterns` governs because it says rich entity screens use Master-Detail and feature screens use virtualized lists with sticky group headers. `c3-104` confirms InvoiceScreen uses a virtualized list with month group headers, and `c3-105` confirms PaymentRequestsScreen supports list/grouped filtering over PR data.

Filter mechanism: `ref-filter-footer` governs because filters live in a `FilterFooter` slide-up footer panel, with active-filter count and compact count feedback. `c3-104` names FilterFooter and invoice filter content; `c3-105` names FilterFooter and PR filter content. The consistent action is to keep filters in the list footer for both screens, not in one screen's header and the other's sidebar.

Detail mechanism: `ref-detail-content-strategy` governs because it defines facet meta grids, BIG grids, section ordering, tabs, empty states, and the distinction between feature `detail-*` classes and admin `admin-detail-*` classes. The dependent observer is the user moving between invoice and PR detail panes: values, badges, lists, audit tabs, and empty states should scan the same way even though the domain data differs.

Direct dependents: `c3-104` and `c3-105` are the direct screen owners that cite the refs. Reverse graphs show `recipe-screen-anatomy` also cites the same layout refs as an orientation recipe. Indirect dependents include navigation and responsive-design recipes returned by the reverse graph for `ref-master-detail-layout`; they are supporting documentation, not owners of invoice or PR screen behavior.

Historical ADR context: the search surfaced responsive and UI-pattern ADRs, including `adr-20260211-responsive-admin-layout` and `adr-20260226-ui-pattern-review-gap-closure`. They are historical context for UI evolution; the current answer should be governed by the active refs and the current `c3-104`/`c3-105` component contracts.

Emergent property: shared layout refs create cross-screen muscle memory. The invoice screen and PR screen can have different domain actions, but users keep the same list/detail rhythm, virtualized scanning, footer filtering, responsive breakpoints, and detail-section hierarchy.

Failure boundary and change checks:

- Do not replace one screen with a custom full-page table if the workflow still needs rich per-entity inspection.
- Check desktop, tablet, and mobile behavior against `ref-responsive-layout`.
- Check both `c3-104` and `c3-105` whenever a shared ref changes, because both directly cite the same layout refs.
- Verify filters still use `FilterFooter`, detail panes still follow `ref-detail-content-strategy`, and virtualized list behavior remains intact for large invoice and PR lists.
