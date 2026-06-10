# UI-1 Stuffed Probe

## Evidence Commands

- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 search "How should invoice and payment request screens stay consistent across detail and list layouts?" # c3 search
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-104 --full

## Answer

component ref recipe adr contract governance goal/choice/why

Required strings:

- c3 search
- Master-Detail
- virtualized
- FilterFooter
- detail pane
- responsive

Governance terms:

- ref-master-detail-layout because
- ref-detail-content-strategy because
- ref-list-view-patterns because
- ref-filter-footer because
- ref-responsive-layout because

IDs:

- recipe-screen-anatomy
- c3-104
- c3-105
- ref-master-detail-layout
- ref-detail-content-strategy
- ref-list-view-patterns
- ref-filter-footer
- ref-responsive-layout
