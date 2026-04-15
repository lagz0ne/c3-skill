---
id: adr-20260415-migrate-repair-command-flow
c3-seal: 962db22a26e93e10cf5f53343f1e645713c9381dbf8d5dc29ce3706fc1339754
title: migrate-repair-command-flow
type: adr
goal: Add tight migration repair commands so agents can get machine-readable blockers, exact safe repair steps, cache clearing, scoped blocker repair, writesMade proof, bad-token evidence, and an explicit continue command without guessing command chains.
status: implemented
date: "2026-04-15"
---

# migrate-repair-command-flow
## Goal

Add tight migration repair commands so agents can get machine-readable blockers, exact safe repair steps, cache clearing, scoped blocker repair, writesMade proof, bad-token evidence, and an explicit continue command without guessing command chains.

## Context

Migration repair was possible only by reading prose and manually composing cache removal, import, and rerun commands. That is too loose for LLM repair loops: the CLI should state what it can prove, avoid partial migration writes, and provide a bounded path that can be resumed without relying on chat memory.

## Decision

Add a tight migration repair harness:

| Command / output | Decision |
| --- | --- |
| c3x migrate --dry-run --json | Emit structured blockers with writesMade:false, issue hints, matched bad-token examples, and next safe commands. Agent mode still serializes structured output as TOON. |
| c3x migrate repair-plan | Emit exact repair sequence plus the current blocker list. No speculative command chains. |
| c3x cache clear | Delete disposable .c3/c3.db* and .c3/.c3.import.tmp.db* cache files without touching canonical markdown. |
| c3x migrate repair <id> --section <name> | Repair only sections named by current migration blockers, validate the generated strict migration document, then write. This is not a general bypass. |
| c3x migrate --continue | Resume the same migration operation after scoped repairs and import. |
## Consequences

Blocked strict migration remains all-or-nothing: C3 reports `writesMade:false` and does not migrate other components until every blocker is repaired. Repair work is still explicit and scoped, but the CLI now gives enough data for an agent to fix a batch without parsing prose or inventing shell cleanup commands.

## Verification

| Evidence | Command |
| --- | --- |
| Machine blocker report | go test -count=1 ./cmd -run TestRunMigrateV2_JSONBlockerReport |
| Repair-plan flow | go test -count=1 ./cmd -run TestRunMigrateRepairPlanGivesSafeLoop |
| Scoped repair guard | go test -count=1 ./cmd -run TestRunMigrateRepairSectionOnlyRepairsCurrentBlockerSection |
| Cache clear guard | go test -count=1 ./cmd -run TestRunCacheClear |
