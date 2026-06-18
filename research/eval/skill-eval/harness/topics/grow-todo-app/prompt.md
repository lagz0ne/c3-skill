# Topic Prompt: Grow a Collaborative TODO Web App

Start with a small C3 project for a single-user TODO web app, then grow it as the
product adds adjacent complexity. Each growth stage introduces one primary pressure;
keep the first rung lean and raise the documentation contract only when a real
pressure demands it.

The product grows through these stages:

1. Single-user TODO web app with task CRUD.
2. Same-user realtime sync across two devices.
3. Reconnect/replay behavior for duplicate and out-of-order sync events.
4. Authenticated user ownership of tasks.
5. Shared list membership with invites and simple roles.
6. Concurrent edits on shared lists.
7. Per-user theming instead of system-wide theming.
8. Migration/refactor of the old single-user, private-list, single-owner, and
   system-theme assumptions when they break.
9. Observability readiness for sync, sharing, migration, and theme behavior.
10. Rollout and recovery with support evidence.

Pressures the docs must make reviewable as they grow: which system owns truth when
two devices disagree (sync reconciliation); who may invite, edit, or only view a
shared list (authority); how reconnect replay avoids duplicating or overwriting a
task; how theme precedence resolves per user; and how a private task is never exposed
or overwritten across a sharing or migration boundary.

Your task:

1. Initialize or grow C3 docs in the isolated project.
2. Keep the first rung lean and complete (do not pre-build later-stage sections).
3. When complexity requires richer structure, grow the docs by raising the
   documentation contract (canvas/rung) and migrating affected facts **completely**.
   Make each assumption break visible as its own growth step, not one late cleanup.
4. End with a project that clearly has multiple containers — at least a web frontend,
   an API/backend, and a realtime-sync surface; persistence and an auth/identity
   boundary as the product warrants.
5. Add components and cross-container work for meaningful feature growth (realtime
   sync, shared-list membership, concurrent-edit resolution, per-user theming).
6. Include migration and document-gardening work to keep the C3 docs coherent as
   single-user/private-list/system-theme assumptions break — and retire stale facts.

Constraints:

- Use local C3 only: `C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh`.
- Avoid codemap work unless absolutely required; it is not the focus of this eval.
- Prefer compact, concrete C3 artifacts over long prose.
- Run verification and report the exact result.
