---
id: ref-zerobased-dev
c3-version: 4
c3-seal: 4ceb0adca922daa1260b1cfca5b55756b9eb6be053bb68dd0ab4afa65e04ab46
title: Zerobased Local Dev Environment
type: ref
goal: All local dev uses zerobased for Docker service routing. No manual port configuration, no setup scripts, no host-mapped ports. Services are auto-discovered via `docker.sock` and routed through Unix sockets (DBs) or `.localhost` subdomains (HTTP).
scope:
    - c3-0
---

# Zerobased Local Dev Environment

## Goal

All local dev uses zerobased for Docker service routing. No manual port configuration, no setup scripts, no host-mapped ports. Services are auto-discovered via `docker.sock` and routed through Unix sockets (DBs) or `.localhost` subdomains (HTTP).

## Choice

**zerobased** as the sole local service router, replacing portless + manual port aliases.

## Why

| Concern | portless (old) | zerobased (new) |
| --- | --- | --- |
| Setup | Manual alias script per project | Zero config — watches docker.sock |
| Port conflicts | Worktrees fight over host ports | No host ports needed |
| DB connections | TCP via mapped port | Unix socket (lower latency, no port clash) |
| HTTP routing | Proxy with explicit aliases | Auto <service>-<port>.<project>.localhost |
| Maintenance | Script must match compose changes | Auto-adapts to container changes |

## How

### 1. Start daemon (once per machine)

```bash
zerobased start
```

### 2. Bring up infrastructure

```bash
docker compose up -d          # in project root
zerobased env acountee        # verify all services discovered
```

### 3. Environment variables

| Variable | Value |
| --- | --- |
| PGURI | postgresql://postgres:pretty-insecure@/postgres?host=/home/lagz0ne/.zerobased/sockets/acountee |
| NATS_SERVER_URL | nats://localhost:<port> — get from zerobased get nats |
| NATS_WS_URL | ws://nats-80.acountee.localhost |
| SMTP_HOST | unchanged (no mailhog in compose) |

### 4. Dev server

```bash
cd apps/start && zerobased run acountee pnpm dev
```

App available at `http://acountee.localhost`.

### 5. Service URLs

| Service | Type | URL / Connection |
| --- | --- | --- |
| App | http | http://acountee.localhost |
| Postgres | socket | postgresql://postgres@/postgres?host=~/.zerobased/sockets/acountee |
| NATS client | port | localhost:<hash> (use zerobased get nats) |
| NATS WebSocket | http | http://nats-80.acountee.localhost |
| NATS monitor | http | http://nats-8222.acountee.localhost |
| OTEL gRPC | socket | unix://~/.zerobased/sockets/acountee/otel-collector-4317.sock |
| OTEL HTTP | http | http://otel-collector-4318.acountee.localhost |

### 6. Verify

- App loads at `http://acountee.localhost`
- Postgres connects via socket (check app logs)
- NATS connects via hashed port
- NATS monitoring at `http://nats-8222.acountee.localhost`

## Not This

| Alternative | Rejected Because |
| --- | --- |
| portless + alias scripts | Manual setup per project, port conflicts across worktrees |
| Host-mapped ports: in docker-compose.yml | Conflicts when multiple worktrees run same compose stack |
| Direct localhost:PORT connections | Fragile, requires remembering port numbers |

## Scope

**Applies to:**

- All local development (main branch and worktrees)
- All Docker Compose services (postgres, nats, otel-collector)
- Reset scripts (`reset-and-dev.sh`, `reset-and-start.sh`)

**Does NOT apply to:**

- CI/CD pipelines (use direct container networking)
- Production / Docker Compose `app` service (uses internal container DNS)
- E2E test environments (may use dedicated compose overrides)

## Override

To override this ref:

1. Document justification in an ADR under "Pattern Overrides"
2. Cite this ref and explain why the override is necessary
3. Specify the scope of the override (which components deviate)

## Cited By
