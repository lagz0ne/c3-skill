# UI-1 Wrong Probe

## Evidence Commands

- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 search "How should invoice and payment request screens stay consistent across detail and list layouts?" # c3 search
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-105 --full

## Answer

Invoice and payment request screens should stay consistent by each owning a separate custom layout. `c3-104` should use a simple invoice table, while `c3-105` should keep a separate approval queue layout. This avoids sharing too much `Master-Detail` behavior between unrelated domains.

`ref-master-detail-layout` governs because it mainly proves `Master-Detail` is optional and can be bypassed whenever a screen has a virtualized list. `ref-list-view-patterns` governs because any list can become a full-page table if it has enough rows. The detail pane can then be opened in a modal instead of through the shared layout.

`ref-filter-footer` governs because filters should move into the header or sidebar for more visibility. `FilterFooter` is useful only as a fallback. `ref-detail-content-strategy` governs because each screen can pick its own grid and section ordering as long as it has labels and values. `ref-responsive-layout` governs because responsive behavior should be custom per screen instead of centralized in MasterDetailLayout.

The direct owners are `recipe-screen-anatomy`, `c3-104`, and `c3-105`; the relevant refs are `ref-master-detail-layout`, `ref-detail-content-strategy`, `ref-list-view-patterns`, `ref-filter-footer`, and `ref-responsive-layout`. Required terms: c3 search, Master-Detail, virtualized, FilterFooter, detail pane, responsive, component, ref, recipe, adr, contract, governance, goal/choice/why.
