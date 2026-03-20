---
id: c3-2
c3-version: 4
title: Claude Skill
type: container
boundary: app
parent: c3-0
goal: Expose c3 architecture workflows through natural language by routing user intent to the right operation and executing it via c3x
summary: SKILL.md intent router + per-operation reference docs that orchestrate AI-driven c3x workflows inside Claude Code sessions
---

# Claude Skill

## Goal

Expose c3 architecture workflows through natural language by routing user intent to the right operation and executing it via c3x.

## Responsibilities

- Classify natural language intent into one of seven operations
- Load and execute the appropriate operation reference
- Call c3x commands at each step of an operation
- Apply ASSUMPTION_MODE when the user declines to answer clarifying questions

## Complexity Assessment

**Level:** moderate
**Why:** Intent classification must be reliable across varied phrasings; each operation is a multi-step workflow; skill descriptions are hard-constrained to ≤1024 chars.

## Components

| ID | Name | Category | Status | Goal Contribution |
|----|------|----------|--------|-------------------|
| c3-201 | skill-router | foundation | active | Classifies intent and dispatches to the correct operation reference |
| c3-210 | operation-refs | feature | active | Provides step-by-step guidance for each operation (onboard/query/audit/change/ref/rule/sweep) |
