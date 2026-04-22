---
id: adr-20260415-migrate-batch-repair-plan
c3-seal: 7dd69bc3452bc4be49faa0c41e9ea957637be9bb8bc844474c8eebc6d4446861
title: migrate-batch-repair-plan
type: adr
goal: Make c3x migrate lead failed legacy repairs with a grouped blocker plan, no partial migration writes, and direct next commands so humans and LLM agents can repair strict component docs without guessing.
status: implemented
date: "2026-04-15"
---

# migrate-batch-repair-plan

## Goal

Make c3x migrate lead failed legacy repairs with a grouped blocker plan, no partial migration writes, and direct next commands so humans and LLM agents can repair strict component docs without guessing.
