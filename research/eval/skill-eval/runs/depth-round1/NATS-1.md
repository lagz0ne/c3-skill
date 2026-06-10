# What breaks if I change NATS websocket authentication?

## Evidence Commands

All run via `c3() { C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 "$@"; }` in `/var/tmp/c3-judge`:

```bash
c3 search "NATS websocket authentication"
c3 read ref-nats-jwt-auth --full
c3 read c3-209 --full
c3 graph ref-nats-jwt-auth --direction reverse --depth 2
c3 graph c3-209 --direction reverse --depth 2
c3 graph ref-nats-jwt-auth --direction reverse --format mermaid   # agent mode returned node list, not mermaid
c3 read ref-sync --full
c3 read recipe-realtime-sync --full
c3 read c3-4 --full
c3 read adr-20260113-nats-jwt-resolver
c3 read adr-20260112-nats-auth-callout
c3 search "client NATS websocket connection natsSync subscribe"
c3 read c3-101 --full
c3 read c3-211 --full
c3 read recipe-auth-and-access --full
c3 graph c3-209 --direction reverse --depth 1 --format mermaid    # agent mode returned node list
c3 read c3-202 --full
c3 lookup '**/nats*' ; c3 lookup '**/_authed*' ; c3 lookup 'infra/**'   # all returned empty (no codemap)
```

## Answer

**C3 Impact Assessment** (sweep — advisory only)

**Proposed Change:** Modify NATS WebSocket authentication — the JWT-resolver credential path owned by c3-209 (NATS Credential Generator) and governed by ref-nats-jwt-auth, enforced by c3-4 (NATS Server), consumed by c3-101 (Web Frontend State Management / `natsSync`).

**Layer:** c3-209 (NATS Credential Generator), c3-2 (API Backend) — generation; c3-4 (NATS Server, external) — enforcement; c3-101 (State Management, c3-1 Web Frontend) — consumption.

### The live mechanism (causal chain)

1. **Credential generation (action owner):** On every authenticated page load, the `_authed.tsx` loader resolves `natsCredentialGenerator` (c3-209) and calls `generate(currentUser.email, 3600)`. c3-209 parses `natsConfig.accountSeed` (tag owned by c3-202 Execution Context, `tags.ts`), creates an ephemeral user keypair (`@nats-io/nkeys createUser()`), embeds **subscribe-only** permissions on `{prefix}.broadcast` and `{prefix}.user.{escaped_email}` (`@` and `.` → `_`), empty publish allow, WEBSOCKET-only connection type, and signs a 1-hour JWT with the account key. Returns `{ jwt, seed }` to the client via `loaderData.natsCredentials`. *(ref-nats-jwt-auth "Credential Flow"; c3-209 "How It Works" + "Permission Model"; c3-202 "Configuration Tags".)*
2. **Enforcement (state gate):** c3-4 (NATS Server) validates the JWT signature against the account public key preloaded in `infra/nats.conf` (`resolver: MEMORY`, `resolver_preload`, `operator.jwt`), checks expiration, and applies the embedded pub/sub permissions per connection on WSS port 8080. No auth-callout service exists. *(ref-nats-jwt-auth "NATS Server Config"; c3-4 "Responsibilities" + "Authentication Flow".)*
3. **Consumption (dependent runtime):** The frontend `natsSync` atom (c3-101) connects with `jwtAuthenticator(jwt, nkeySeed)` and subscribes to two subjects: `sync.broadcast` (deltas + acks → prs/invoices/payments stores + `executionTracker.notify`) and `sync.user.<email>` (notification messages → notifications store). *(c3-101 "NATS Sync Wiring"; ref-nats-jwt-auth "Client Connection".)*
4. **What rides on that connection:**
   - **Real-time sync (ref-sync):** services emit deltas, flows emit acks on `{prefix}.broadcast`; the originating client's `result.wait()` resolves on the matching `executionId` arriving over this WebSocket. *(ref-sync "Architecture" + "Execution ID Contract".)*
   - **In-app notifications (c3-211):** `inAppChannel` delivers via "NATS publish (real-time) + JetStream (persistence)"; the real-time leg lands on `sync.user.{escaped_email}` (per ref-sync subject table), which the browser can only receive because c3-209 embedded that subject in the JWT's subscribe permissions. *(c3-211 "Built-in Channels"; ref-sync "NATS Subjects"; c3-209 "Permission Model".)*
5. **Emergent property:** auth is the **shared coupling between two otherwise-independent delivery paths** (broadcast sync and per-user notification). The server publishes over full-access TCP 4222 with its own APP account credentials (`APP_ACCOUNT_JWT` etc., c3-4 "Configuration"), so server-side emit/ack keeps succeeding even when every browser is locked out.
6. **Failure boundary:** if browser WS auth breaks, **HTTP mutations still succeed** — the app looks healthy. What degrades: clients stop receiving deltas (stale UI until reload), `result.wait()` resolves only via the 2s timeout fallback ("a UX optimization, not correctness-critical" — ref-sync), and in-app real-time notification delivery silently stops. JetStream-persisted notifications, email, and Slack channels are unaffected (server-side TCP, c3-211). Nobody gets a hard error; the observable is sluggish UI + stale data.

### Impact graph

Built from the reverse-graph node lists (`c3 graph ref-nats-jwt-auth --direction reverse`, `c3 graph c3-209 --direction reverse`) plus ref-sync "Cited By" — agent mode did not render mermaid, so this is hand-drawn from those outputs:

```mermaid
graph RL
  c3209["c3-209 NATS Credential Generator (change target)"] -->|uses/governed by| ref["ref-nats-jwt-auth"]
  c34["c3-4 NATS Server (enforces JWT)"] -.->|validates creds from| c3209
  c3101["c3-101 State Management / natsSync (consumes jwt+seed)"] -.->|connects with| c3209
  c3202["c3-202 Execution Context (natsConfig.accountSeed tag)"] -.->|feeds| c3209
  refsync["ref-sync runtime delivery"] -.->|rides WS connection| c3101
  c3211["c3-211 Notification System (inAppChannel real-time leg)"] -.->|delivers via sync.user.*| c3101
  screens["c3-104/c3-105 screens, c3-205/206/207/212 flows (cite ref-sync)"] -.->|observe deltas via| refsync
  recipe["recipe-auth-and-access"] -->|documents| ref
```

### Affected Entities

| Entity | Type | Impact | Reason (from read) |
|--------|------|--------|--------------------|
| c3-209 | component | direct | Owns generation: permissions shape, TTL, email escaping, signing key (c3-209 read) |
| ref-nats-jwt-auth | ref | direct | Governing ref; cited by c3-209, adr-20260113, recipe-auth-and-access (reverse graph). Must be updated in lockstep |
| c3-4 | container | direct | Enforcement point: `infra/nats.conf` resolver_preload + operator/account JWTs must match the signing seed (c3-4 read) |
| c3-101 / c3-1 | component/container | direct | `natsSync` consumes `loaderData.natsCredentials` via `jwtAuthenticator` (c3-101 read) |
| c3-202 | component | direct | Owns `natsConfig.{accountSeed, wsUrl, subjectPrefix}` tags the generator depends on (c3-202 read) |
| ref-sync (runtime) | ref | transitive | Delta/ack delivery to browsers rides the authenticated WS; server emit side unaffected (ref-sync read) |
| c3-211 | component | transitive | inAppChannel real-time leg lands on `sync.user.{escaped_email}` — only receivable if the JWT grants that subject (c3-211 + c3-209 reads) |
| c3-104, c3-105, c3-205, c3-206, c3-207, c3-212 | components | transitive | Cite ref-sync ("Cited By" list); screens observe stale data / flows' acks arrive at no one. Not read individually — candidates from ref-sync citation list, not asserted behavior |
| recipe-auth-and-access, recipe-realtime-sync | recipes | transitive (docs) | Narrate this path; need re-verification after change (both reads) |

### Specific break modes

- **Signing/verification key mismatch** (rotate `NATS_ACCOUNT_SEED` without updating `resolver_preload` account public key in `infra/nats.conf`, or vice versa): every browser connection rejected at step "Verify signature". App keeps serving HTTP. (ref-nats-jwt-auth "Troubleshooting: Connection rejected"; c3-4 "Authentication Flow".)
- **Permission shape change** (e.g., dropping the `{prefix}.user.{escaped_email}` subscribe grant): broadcast sync still works while in-app notifications silently die — partial degradation that's easy to miss. (c3-209 "Permission Model"; c3-211 "Built-in Channels".)
- **Email-escaping change**: the escape rule (`@`,`.` → `_`) appears in c3-209 (permission grant), ref-sync (publisher `publishToUser()` subject), and c3-101 (client subscription). All three must change together or per-user delivery breaks while everything else stays green.
- **Subject prefix change**: ref-sync "Subject Prefix Contract" — server subjects are prefix-driven (`natsConfig.subjectPrefix`, default `sync`) but "frontend subscriptions currently use `sync.broadcast` and `sync.user.{escaped_email}` directly. If prefix changes from `sync`, frontend subscription wiring must change in lockstep." Auth permissions embed the prefixed subjects too (c3-209).
- **TTL/expiry change**: JWT expires after 1h; "client must reconnect" (c3-209 "Security"). Shortening TTL without reconnect handling = sessions that go quiet mid-use.
- **Account-level key regeneration** (operator/account JWTs): also breaks the backend's TCP publisher leg — c3-4 requires `APP_ACCOUNT_JWT`/`SYS_ACCOUNT_JWT` env vars and APP-account pub/sub permissions on `{prefix}.>`, `$JS.API.>`, `_INBOX.>`; losing `$JS.API.>` would also break JetStream notifications (c3-4 "Required Permissions").

### Constraint Chain

| Source | Constraint | Status if you change auth |
|--------|-----------|---------------------------|
| ref-nats-jwt-auth | Subscribe-only browser permissions; account seed server-side only; no auth-callout service | Must remain or ref needs explicit update + cascade review of citers |
| ref-sync | `executionId` string end-to-end; `result.wait()` 2s fallback; broadcast-to-all, filter on client | Unchanged by auth itself, but its delivery leg depends on the WS connection succeeding |
| c3-4 | Browser read-only; backend controls all publications | Any grant of publish rights to browsers violates this boundary |
| Rules | No `rule-*` entities surfaced in any search or graph output for this area (c3-209 `uses` lists only refs) | n/a |

### ADR status labels

- **adr-20260113-nats-jwt-resolver** — `status: implemented` (terminal). This is the decision behind the **current live mechanism** (JWT resolver, no callout); historical work order, content frozen.
- **adr-20260112-nats-auth-callout** — body says "**Superseded** - 2026-01-13. Replaced by JWT resolver approach." Label: **superseded** — do not treat the callout design ($SYS.REQ.USER.AUTH service) as live. (Frontmatter `status` field still reads `implemented`; the body Status section is the supersession record.)
- **adr-20260112-nats-websocket-sync** — historical (original NATS sync adoption; surfaced in c3-209 reverse graph at depth 2).
- **adr-20260126-user-notification-ui** — historical/implemented; explains why `natsSync` subscribes to both subjects ("Extend `natsSync.ts` to subscribe to both subjects", from search output). Current entity docs (c3-101) confirm the dual subscription is live.

### Verification

| Check | How |
|-------|-----|
| Key pairing intact | Derive public key from `NATS_ACCOUNT_SEED`, confirm it equals the `resolver_preload` key in `infra/nats.conf` (ref-nats-jwt-auth "Generating NKeys") |
| Credential generation | Load an authenticated page, confirm `loaderData.natsCredentials` contains `jwt` + `nkey`/seed |
| Connection accepted | Browser connects to WSS 8080; NATS monitoring at `http://nats-server:8222/healthz` healthy; no rejection logs (c3-4 "Health Check") |
| Broadcast sync observable | Run a mutation from client A, assert client B receives a delta on `sync.broadcast` and client A's `result.wait()` resolves well under the 2s timeout (timeout-resolution = delivery broken, per ref-sync) |
| Per-user leg observable | Trigger an approval notification, assert a message arrives on `sync.user.{escaped_email}` and the notifications store/bell updates (c3-101, c3-211) |
| Backend leg unaffected | Confirm backend TCP 4222 publish still works (`{prefix}.>`) and JetStream `NOTIFICATIONS` consumer still acks (c3-4, c3-211) |
| Failure-mode probe | Connect with an expired or garbage JWT → expect NATS rejection while HTTP app stays healthy; confirm UI degrades to 2s-timeout waits + stale data, not errors |
| Docs in lockstep | Update ref-nats-jwt-auth, c3-209, c3-4 (and c3-101 if client wiring changes) together; `c3 check` after mutation; cascade-review citers per ref help hint |

### Recommended Approach

1. Treat c3-209 + ref-nats-jwt-auth + `infra/nats.conf` (c3-4) as one atomic change set — the signing seed and resolver public key are a matched pair.
2. Preserve the permission envelope (subscribe-only, both subjects, WEBSOCKET-only) unless the change is explicitly about permissions; if it is, re-verify the c3-211 in-app leg separately from broadcast sync — they fail independently.
3. Open an ADR first (per skill change discipline) if this becomes an actual change; this sweep is advisory only.

## Grounding

| Claim | Source |
|-------|--------|
| Loader generates creds via `credGen.generate(currentUser.email, 3600)`, returns `{jwt, nkey}` via `loaderData.natsCredentials`, 1h expiry | `c3 read ref-nats-jwt-auth --full` ("Credential Flow", "Choice") |
| Permissions: subscribe `{prefix}.broadcast` + `{prefix}.user.{escaped_email}`, empty publish, WEBSOCKET-only; email escaping `@`/`.` → `_`; seed server-side only; ephemeral user keys | `c3 read c3-209 --full` ("Permission Model", "Security", "How It Works") |
| NATS validates via MEMORY resolver + `resolver_preload` in `infra/nats.conf`, operator JWT, no callout; WSS 8080 / TCP 4222; backend APP-account perms `{prefix}.>`, `$JS.API.>`, `_INBOX.>`; healthz on 8222 | `c3 read ref-nats-jwt-auth --full` ("NATS Server Config") + `c3 read c3-4 --full` ("Wiring", "Required Permissions", "Health Check") |
| `natsSync` connects with `jwtAuthenticator`, subscribes `sync.broadcast` + `sync.user.<email>`, feeds stores + `executionTracker`; prefix hardcoded client-side | `c3 read c3-101 --full` ("NATS Sync Wiring", "Atoms") |
| Deltas/acks ride broadcast; `result.wait()` is UX-only with 2s timeout fallback; executionId string contract; prefix lockstep warning; Cited-By list (c3-104/105/205/206/207/212/209/211) | `c3 read ref-sync --full` ("Execution ID Contract", "Subject Prefix Contract", "Cited By") |
| inAppChannel = "NATS publish (real-time) + JetStream (persistence)"; email/Slack channels server-side; JetStream `NOTIFICATIONS` stream durable | `c3 read c3-211 --full` ("Built-in Channels", "notificationPublisher") + `c3 read recipe-realtime-sync --full` (sync ephemeral vs notifications durable) |
| `natsConfig.{wsUrl, accountSeed, subjectPrefix}` tags live in c3-202 `tags.ts` | `c3 read c3-202 --full` ("Configuration Tags") |
| NATS auth is "separate from HTTP auth — NATS has its own identity layer"; creds generated by c3-209 | `c3 read recipe-auth-and-access --full` |
| Direct dependents of ref-nats-jwt-auth = adr-20260113, c3-209, recipe-auth-and-access | `c3 graph ref-nats-jwt-auth --direction reverse --depth 2` (cited-by list on ref node) |
| ADR statuses: jwt-resolver implemented 2026-01-13; auth-callout body "Superseded - 2026-01-13"; pre-auth state was open/unauthenticated WS | `c3 read adr-20260113-nats-jwt-resolver`, `c3 read adr-20260112-nats-auth-callout` |
| `natsSync.ts` dual-subject extension decision | `c3 search "client NATS websocket connection natsSync subscribe"` (adr-20260126-user-notification-ui snippet), confirmed live in c3-101 read |
| No `rule-*` entities apply | No rule entities in either `c3 search` output, both reverse graphs, or c3-209/c3-101 `uses` lists |

## Caveats

- **Doc drift on the browser permission table:** ref-nats-jwt-auth "Permissions Model" lists browser subscribe as only `sync.broadcast`, while c3-209, c3-4, and c3-101 all include `{prefix}.user.{escaped_email}`. The ref's table appears stale relative to the dual-subject reality (adr-20260126 era). If you implement from the ref alone you'd drop the notification leg. Evidence: the four reads above.
- **Inconsistent `_INBOX.>` claim:** c3-4 "Permission Model" grants browser publish/subscribe `_INBOX.>`; c3-209 says publish is "empty allow (no permissions)". The docs disagree; confirm against actual generator code before relying on either.
- **ADR frontmatter vs body:** adr-20260112-nats-auth-callout has frontmatter `status: implemented` but body Status "Superseded - 2026-01-13" — the supersession is recorded only in the body.
- **No codemap in this fixture:** `c3 lookup '**/nats*'`, `'**/_authed*'`, and `'infra/**'` all returned empty file maps. Every file path cited here (`_authed.tsx`, `natsSync.ts`, `tags.ts`, `infra/nats.conf`) comes from doc bodies, not from code-map lookup, and could not be cross-verified against source.
- **Transitive screens/flows not individually read:** c3-104/c3-105/c3-205/c3-206/c3-207/c3-212 are listed from ref-sync's "Cited By" only; their per-entity behavior under auth failure is inferred from ref-sync's delivery contract, not from reading each entity.
- **Mermaid graph is hand-assembled:** `--format mermaid` under `C3X_MODE=agent` returned node lists, not mermaid; the diagram reflects those node/edge outputs but was rendered by hand.
