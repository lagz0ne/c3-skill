# C3 Blindbox Growth Rubric

This rubric scores real Codex, Claude, and Kilo outputs from
`harness/bin/run-blindbox.sh`. The candidate sees only the prompt packet and an
isolated project.

| ID | Dimension | Score |
| --- | --- | --- |
| G1 | Local C3 isolation: uses `C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh`; no bare/global `c3x`, no host skill paths. | 0-2 |
| G2 | Multi-container system: models at least frontend, backend, integration, and database containers with explicit responsibilities. | 0-3 |
| G3 | Component boundaries: creates or updates components under those containers and explains ownership boundaries. | 0-3 |
| G4 | Concepts and cross-container work: documents domain concepts and cross-container flows rather than only listing screens/services. | 0-3 |
| G5 | Feature growth: adds meaningful complexity such as reservations, transfers, quarantine, returns, cycle counts, reporting, carrier/procurement/finance integration, or role-specific operations. | 0-3 |
| G6 | Rung behavior: treats current facts as complete to their rung and grows by raising canvas/contracts plus migrating all affected facts. | 0-4 |
| G7 | Migration and document gardening: includes concrete migration, repair, and doc cleanup steps, with verification after each growth step. | 0-4 |
| G8 | Change discipline: uses ADR/change-unit style work for architecture changes and avoids direct frozen-fact mutation once facts exist. | 0-3 |
| G9 | Verification: runs or plans concrete C3 checks and reports failures honestly. | 0-3 |
| G10 | Codemap restraint: does not make code-map/codemap work central to passing the task. | 0-2 |

Suggested pass bar: at least 24/30, with `G1 >= 1`, `G2 >= 2`, `G6 >= 3`,
`G7 >= 3`, and no invented claim that verification passed when it did not run.

## Review Guidance

Strong answers show a sequence: initialize lean C3 docs, identify pressure for a
higher rung, raise the documentation contract, migrate every affected container
or component to that contract, add feature/cross-container work, then verify.

Weak answers usually fail by creating a plausible warehouse architecture but
not showing C3 growth mechanics, by keeping everything in one container, by
using codemap as a shortcut, or by saying "migration" without naming affected
docs and checks.
