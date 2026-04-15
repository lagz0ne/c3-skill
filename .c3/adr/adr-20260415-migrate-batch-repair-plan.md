---
id: adr-20260415-migrate-batch-repair-plan
c3-seal: 160a14d666481f2a5e13af5500cc99ce786dfe349b64cfac7d0c8cbcac0c1c34
title: migrate-batch-repair-plan
type: adr
goal: Make c3x migrate lead failed legacy repairs with a grouped blocker plan, no partial migration writes, and direct next commands so humans and LLM agents can repair strict component docs without guessing.
status: implemented
date: "2026-04-15"
---

# migrate-batch-repair-plan
## Goal

Make c3x migrate lead failed legacy repairs with a grouped blocker plan, no partial migration writes, and direct next commands so humans and LLM agents can repair strict component docs without guessing.
