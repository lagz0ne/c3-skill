---
id: adr-20260414-rollout-sharpening-coverage
c3-seal: 869d7d71a947a8a896304a8daf8fbf05b1f54c298095b45d7f48a35fa835ebd6
title: rollout-sharpening-coverage
type: adr
goal: Expand C3 ownership coverage so lookup and context-gate flow cover most active CLI paths during rollout verification.
status: proposed
date: "2026-04-14"
---

# Rollout Sharpening Coverage

## Goal

Expand C3 ownership coverage so lookup and context-gate flow cover most active CLI paths during rollout verification.

## Context

Current repo has working v9 flow and baseline codemap, but many active command families remain unmapped.

## Decision

Map all files already implied by current components first, then add minimal missing components for major unmapped command families only where ownership is clearly distinct.

## Consequences

- Better lookup coverage for real repo work
- Fewer blind spots during change/audit/query flows
- Architecture grows only where current repo already shows stable ownership boundaries
