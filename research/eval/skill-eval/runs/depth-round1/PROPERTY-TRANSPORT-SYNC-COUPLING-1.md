# How does changing NATS JWT auth affect the sync layer?

## Evidence Commands

All run as `c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }`:

```bash
c3 search "How does changing NATS JWT auth affect the sync layer?"
c3 read recipe-auth-and-access --full
c3 read recipe-realtime-sync --full
c3 read ref-nats-jwt-auth --full
c3 read c3-209 --full                  # NATS Credential Generator
c3 read ref-sync --full
c3 read c3-101 --full                  # State Management (natsSync atom)
c3 read c3-4                           # NATS Server (External)
c3 read c3-211                         # Notification System
c3 read adr-20260113-nats-jwt-resolver --full
c3 read adr-20260112-nats-auth-callout
c3 read adr-20260112-nats-websocket-sync
c3 graph c3-209 --depth 1 --direction reverse --format mermaid
c3 graph ref-nats-jwt-auth --direction reverse --format mermaid
c3 graph ref-sync --direction reverse --format mermaid
c3 lookup 'lib/pumped/atoms/natsSync.ts'   # no codemap match
c3 lookup '**/_authed*'                    # no codemap match
c3 lookup 'infra/nats.conf'                # no codemap match
```

## Answer

**Layer:** spans c3-2 (API Backend: c3-209), c3-4 (NATS Server, external), c3-1 (Web Frontend: c3-101), governed by ref-nats-jwt-auth and ref-sync.

### The coupling, named

**Transport-auth-feeds-sync**: the sync layer's entire client-side receive surface is *embedded inside the JWT* that the auth layer mints. The shared tokens are:

- the **subjects** `{prefix}.broadcast` and `{prefix}.user.{escaped_email}` (default prefix `sync`, from `natsConfig.subjectPrefix` / `NATS_SUBJECT_PREFIX`) — they appear simultaneously in the JWT's subscribe-permission list (c3-209, Permission Model) and in the frontend `natsSync` atom's subscriptions (c3-101, NATS Sync Wiring) and in the server publisher contract (ref-sync, NATS Subjects);
- the **account keypair**: `NATS_ACCOUNT_SEED` signs user JWTs (c3-209 Dependencies); the matching **account public key** must sit in the NATS `resolver_preload` (`infra/nats.conf`) for validation (ref-nats-jwt-auth, Configuration);
- the **email-escaping rule** (`@` and `.` → `_`) shared between the JWT's per-user subscribe permission (c3-209) and the publisher's `publishToUser()` subject (ref-sync);
- the **1-hour TTL** on credentials minted at page load (ref-nats-jwt-auth, Credential Flow).

There is **no graph edge** from the sync layer to the JWT auth ref: reverse graph of `ref-nats-jwt-auth` shows only `c3-209`, `adr-20260113-nats-jwt-resolver`, and `recipe-auth-and-access`; reverse graph of `c3-209` shows only its parent `c3-2`. The coupling lives in runtime values (subjects, keys), documented in body text of c3-101/ref-sync, not in citation topology — so a citation-graph-only impact sweep on JWT auth **misses the sync layer entirely**.

### Causal chain (generation → enforcement → dependent path)

1. **Credential generation** — `c3-209 NATS Credential Generator` (backend, c3-2): on authenticated page load, the `_authed.tsx` loader calls `natsCredentialGenerator.generate(currentUser.email, 3600)`; it creates an ephemeral user keypair, builds claims with **subscribe-only** permissions on `{prefix}.broadcast` + `{prefix}.user.{escaped_email}`, **empty publish allow**, **WEBSOCKET-only**, signs with the account seed, returns `{ jwt, seed }` to the browser via `loaderData.natsCredentials` (c3-209 How It Works; ref-nats-jwt-auth Credential Flow). *Why next hop follows:* the JWT carries the permissions; NATS is the only thing that reads them.
2. **Enforcement point** — `c3-4 NATS Server`: MEMORY resolver with `resolver_preload` (account public key → account JWT) validates the JWT signature chain and expiry **inline, with no auth callout service**, then **enforces the embedded pub/sub permissions** on every subscribe/publish (ref-nats-jwt-auth NATS Validation + NATS Server Config; c3-4 Responsibilities: "Permission Enforcement — Enforce pub/sub permissions from JWT"). *Why next hop follows:* the frontend's subscriptions only succeed if they fall inside the JWT's permission set.
3. **Dependent runtime path** — `c3-101 State Management`: the `natsSync` atom connects via WebSocket using `jwtAuthenticator(jwt, nkeySeed)` and subscribes to exactly two subjects: `sync.broadcast` (deltas + acks → update prs/invoices/payments stores and `executionTracker.notify(executionId)`) and `sync.user.<email>` (notification messages → notifications store) (c3-101 NATS Sync Wiring; ref-sync Golden Example: Client-Side Subscription). Server-side, services `sync.emit()` deltas and flows `sync.ack(executionId)` over **TCP with full permissions — not user JWTs** (ref-sync Architecture; ref-nats-jwt-auth Permissions Model: server = full TCP).
4. **Observers/dependents of that path** — all flow components citing ref-sync (`c3-205` PR Flows, `c3-206` Invoice Flows, `c3-207` Payment Flows, `c3-210` Admin Flows, `c3-212` Workbench Flows, per reverse graph of ref-sync) emit the deltas/acks; screens `c3-104`/`c3-105` and `c3-211` Notification System are listed in ref-sync Cited By. These are **transitive** dependents of JWT auth — they reach it only through the c3-101 subscription leg. **Direct** dependents of JWT auth are `c3-209` (mints), `c3-4` (enforces), and `c3-101` (consumes credentials at connect).
5. **Emergent property**: the sync layer is *receive-gated by auth but send-independent of it*. Mutations, DB writes, and publishes never touch user JWTs; only delivery to browsers does.
6. **Failure boundary**: below.

```mermaid
graph LR
  subgraph c3-2 API Backend
    L["_authed.tsx loader"] --> G["c3-209 Credential Generator<br/>JWT: sub-only sync.broadcast + sync.user.*, 1h TTL"]
    F["c3-205/206/207/210/212 Flows + Services<br/>sync.emit / sync.ack"]
  end
  G -- "loaderData.natsCredentials" --> C["c3-101 natsSync atom"]
  C -- "WSS + jwtAuthenticator" --> N["c3-4 NATS Server<br/>MEMORY resolver_preload<br/>enforces JWT permissions"]
  F -- "TCP 4222, full perms (no user JWT)" --> N
  N -- "sync.broadcast (deltas+acks)<br/>sync.user.{escaped_email} (notifications)" --> C
  C --> S["stores: prs/invoices/payments<br/>executionTracker.wait()<br/>notifications"]
```
*(Mermaid rendered from the `c3 graph` node outputs above — agent mode emitted TOON node lists rather than mermaid text.)*

### How the sync layer fails while upstream looks healthy

The asymmetry is the point: HTTP auth (cookie session, RBAC) is a **separate identity layer** ("Separate from HTTP auth — NATS has its own identity layer", recipe-auth-and-access), and the server's emit side uses full TCP access, not user JWTs. So after a JWT-auth change, every upstream signal stays green — login works, mutations succeed, `sync.emit()`/`sync.ack()` publish without error — while the browser receive leg is dead:

- **Key rotation without resolver update** (rotate `NATS_ACCOUNT_SEED` but not `resolver_preload` public key, or vice versa): NATS rejects every browser connection ("Invalid JWT signature" — ref-nats-jwt-auth Troubleshooting). No client receives deltas/acks/notifications.
- **Permission-set narrowing**: the connection itself can still be *established* (auth "healthy" by the connection-up signal) while a subscribe is denied. Concretely documented hazard: ref-nats-jwt-auth's Permissions Model table grants browsers subscribe on **`sync.broadcast` only**, while c3-209's Permission Model and c3-101's wiring require **both** `{prefix}.broadcast` *and* `{prefix}.user.{escaped_email}`. A change that implements the ref's narrower table keeps broadcast sync working but silently kills per-user notification delivery.
- **Prefix or escaping drift**: ref-sync's Subject Prefix Contract states frontend subscriptions hardcode `sync.*`; changing `NATS_SUBJECT_PREFIX` (or the `@`/`.` → `_` escaping in c3-209) without lockstep frontend/JWT changes means publishes land on subjects no client is permitted on/subscribed to. Nothing errors server-side.
- **TTL expiry mid-session**: JWTs expire after 1h and "client must reconnect" (c3-209 Security); a long-lived page loses sync while its HTTP session remains valid.

**Who observes the failure, and how it degrades:** per ref-sync's Execution ID Contract, `result.wait()` is "a UX optimization, not correctness-critical" with a **2s timeout fallback** — so the originating user sees sluggish-but-successful actions (wait resolves by timeout, not by ack); all *other* clients simply go stale until reload (SSR loaders re-fetch state, c3-101 SSR Hydration). Per-user notifications to the browser stop arriving. Note the durability split (recipe-realtime-sync Risk + c3-211): broadcast sync is **ephemeral** (missed deltas are gone), while notifications are **durable in JetStream server-side** — but their browser delivery leg still rides the JWT-gated `sync.user.*` subject, so the user doesn't see them until the delivery path is restored. No documented alerting observes the dead receive leg; the failure surfaces as UI staleness.

### ADR labels

| ADR | Status field | Label for this answer |
| --- | --- | --- |
| adr-20260113-nats-jwt-resolver | implemented | **current** — describes the live mechanism (JWT resolver, no callout); matches ref-nats-jwt-auth and c3-4's wiring row "no external auth callout" |
| adr-20260112-nats-auth-callout | implemented (frontmatter); body says "**Superseded** — 2026-01-13" | **superseded** by adr-20260113; do not treat auth-callout as the live mechanism |
| adr-20260112-nats-websocket-sync | implemented | **historical** — established the NATS transport and the `sync.broadcast` / `sync.user.{email}` subjects (still live), but its point 4 ("NATS auth callout") was superseded by the JWT resolver |

### Concrete checks when changing NATS JWT auth

1. **Key pairing**: assert the account public key in `infra/nats.conf` `resolver_preload` corresponds to the `NATS_ACCOUNT_SEED` used by c3-209 (ref-nats-jwt-auth Configuration lists both env vars).
2. **Permission completeness**: assert the minted JWT's subscribe list contains *both* `{prefix}.broadcast` and `{prefix}.user.{escaped_email}` (c3-209 Permission Model) — not just `sync.broadcast` as ref-nats-jwt-auth's table suggests.
3. **Prefix lockstep**: if `NATS_SUBJECT_PREFIX` ≠ `sync`, update frontend `natsSync` subscriptions in the same change (ref-sync Subject Prefix Contract; c3-101 NATS Sync Wiring).
4. **Runtime probe of the failure mode**: with new credentials, load an authed page, perform a mutation, and assert (a) the delta visibly updates a *second* client, and (b) `result.wait()` resolves *before* the 2s timeout — timeout-resolution is the silent-failure signature (ref-sync Execution ID Contract). Then assert a per-user notification arrives on `sync.user.<escaped_email>`.
5. **TTL behavior**: keep a page open past the 1h expiry and verify reconnect behavior (c3-209 Security: "client must reconnect").

## Grounding

| Claim | Source |
| --- | --- |
| JWT minted per session at page load in `_authed.tsx` loader, `{jwt, seed}` via `loaderData.natsCredentials`, 1h TTL | `c3 read ref-nats-jwt-auth --full` (Credential Flow, Choice); `c3 read c3-209 --full` (How It Works) |
| JWT permissions: subscribe-only `{prefix}.broadcast` + `{prefix}.user.{escaped_email}`, empty publish, WEBSOCKET-only; email escaping `@`/`.` → `_`; signed with `natsConfig.accountSeed` | `c3 read c3-209 --full` (Permission Model, Dependencies, Configuration) |
| Enforcement: MEMORY resolver + `resolver_preload` in `infra/nats.conf`, inline signature/expiry validation, permission enforcement from JWT, no callout | `c3 read ref-nats-jwt-auth --full` (NATS Validation, NATS Server Config); `c3 read c3-4` (Wiring, Responsibilities) |
| Frontend `natsSync` connects with `jwtAuthenticator(jwt, seed)`, subscribes `sync.broadcast` + `sync.user.<email>`, updates stores + executionTracker + notifications | `c3 read c3-101 --full` (Atoms table, NATS Sync Wiring); `c3 read ref-sync --full` (Client-Side Subscription) |
| Server emits over TCP with full permissions (not user JWTs); services emit deltas, flows ack with executionId | `c3 read ref-nats-jwt-auth --full` (Permissions Model table); `c3 read ref-sync --full` (Architecture); `c3 read c3-4` (Wiring) |
| `result.wait()` UX-only, 2s timeout fallback; skipping ack ⇒ "sluggish UI" | `c3 read ref-sync --full` (Execution ID Contract, Anti-patterns) |
| Prefix contract: frontend hardcodes `sync.*`; prefix change requires lockstep frontend change | `c3 read ref-sync --full` (Subject Prefix Contract); `c3 read c3-101 --full` (NATS Sync Wiring) |
| Permission-table discrepancy (ref says broadcast-only; c3-209 + c3-101 require user subject too) | `c3 read ref-nats-jwt-auth --full` (Permissions Model) vs `c3 read c3-209 --full` (Permission Model) vs `c3 read c3-101 --full` |
| NATS auth separate from HTTP auth ("own identity layer") | `c3 read recipe-auth-and-access --full` (Narrative) |
| Sync ephemeral vs notifications durable (JetStream `NOTIFICATIONS`), architecturally separate | `c3 read recipe-realtime-sync --full` (Narrative, Risk); `c3 read c3-211` (Goal) |
| Direct dependents of JWT auth ref: c3-209, adr-20260113, recipe-auth-and-access only — no edge to sync entities | `c3 graph ref-nats-jwt-auth --direction reverse`; `c3 graph c3-209 --depth 1 --direction reverse` |
| Transitive dependents via ref-sync: c3-101, c3-205, c3-206, c3-207, c3-210, c3-212 (+ Cited By: c3-104, c3-105, c3-211, c3-209) | `c3 graph ref-sync --direction reverse`; `c3 read ref-sync --full` (Cited By) |
| ADR statuses (current / superseded / historical) | `c3 read adr-20260113-nats-jwt-resolver --full` (Status: Implemented 2026-01-13); `c3 read adr-20260112-nats-auth-callout` (body: "Superseded — 2026-01-13. Replaced by JWT resolver"); `c3 read adr-20260112-nats-websocket-sync` (Status: Implemented 2026-01-12; Decision point 4 = auth callout) |
| TTL expiry requires reconnect | `c3 read c3-209 --full` (Security); `c3 read ref-nats-jwt-auth --full` (Security Considerations) |

## Caveats

- **Codemap gap**: `c3 lookup` on `lib/pumped/atoms/natsSync.ts`, `**/_authed*`, and `infra/nats.conf` returned no matches ("codemap coverage gap" in c3x help output). File-level claims (loader location, nats.conf contents) rest on doc bodies (ref-nats-jwt-auth, c3-209, c3-101), not on code-map-verified source reads.
- **Doc discrepancy is real and unresolved in the fixture**: ref-nats-jwt-auth's Permissions Model table (subscribe: `sync.broadcast` only) conflicts with c3-209's Permission Model and c3-101's wiring (both subjects). I treated c3-209/c3-101 as authoritative for runtime behavior since c3-209 is the component that mints the claims; an audit of ref-nats-jwt-auth would be warranted.
- **Dependent-list discrepancy**: `c3 graph ref-sync --direction reverse` includes c3-210 (Admin Flows), which ref-sync's own Cited By section omits; conversely Cited By lists c3-104/c3-105/c3-209/c3-211 which appear without `uses` edges in the reverse graph output. Both sources are reported above.
- **No `rule-*` entities surfaced** for NATS auth or sync in the search results or any read; the governing constraints are refs (ref-nats-jwt-auth, ref-sync) and recipes only.
- **No documented alerting/observer** for a dead browser receive leg appears in any read entity; "failure surfaces as UI staleness" is the documented degradation (ref-sync timeout fallback), and absence of monitoring is an inference from absence in the docs, not a verified property of the code.
- Agent mode (`C3X_MODE=agent`) emitted TOON node lists for `c3 graph ... --format mermaid`; the mermaid diagram in the answer is hand-rendered from those node outputs.
