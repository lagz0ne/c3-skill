---
id: ref-onboarding-patterns
title: Onboarding and Progressive Disclosure
---

# Onboarding and Progressive Disclosure

## Goal

Define patterns for guided user onboarding and progressive feature disclosure, enabling personalized discovery paths and gradual complexity revelation.

## Overview

Conventions for guided user experiences and gradual feature revelation.

## Onboarding Philosophy

### Principles

| Principle | Implementation |
|-----------|----------------|
| Show, don't tell | Interactive, not modal text |
| Respect agency | Skip always available |
| Personalize | Adapt to stated goals |
| Progressive | Reveal complexity gradually |

### Anti-patterns

| Anti-pattern | Instead |
|--------------|---------|
| Long tutorial video | Interactive steps |
| Feature tour | Contextual hints |
| Forced completion | Optional, skippable |
| One-size-fits-all | Goal-based paths |

## Progressive Disclosure

### Revelation Triggers

| Trigger | What Reveals |
|---------|--------------|
| First concept created | Linking suggestions |
| First link created | AI suggestions panel |
| 5 concepts | Advanced templates |
| First AI question | Full chat features |
| Collaboration invite | Sharing controls |

### Feature Tiers

| Tier | Features | Unlock |
|------|----------|--------|
| 1 - Core | Create, position, basic links | Default |
| 2 - Enhanced | AI suggestions, templates | Engagement |
| 3 - Advanced | Import, export, analytics | Request or time |
| 4 - Power | API access, custom templates | Subscription |

## Template System

### Template Structure

| Field | Description |
|-------|-------------|
| name | Display name |
| description | Use case description |
| goals | Target user goals |
| seedConcepts | Initial concept set |
| progressiveSteps | Phased revelation |
| adaptations | AI customization points |

### AI Adaptation

| Input | Adaptation |
|-------|------------|
| User-stated goal | Concept prioritization |
| Domain keywords | Template selection |
| Existing knowledge | Skip redundant concepts |
| Learning style | Pace adjustment |

## Hint System

### Hint Types

| Type | Trigger | Display |
|------|---------|---------|
| Contextual | Feature proximity | Tooltip |
| Achievement | Milestone reached | Toast |
| Suggestion | Inactivity | Subtle banner |
| Tutorial | Explicit request | Modal/panel |

### Hint State

| State | Description |
|-------|-------------|
| Unseen | Never shown |
| Seen | Shown once |
| Dismissed | User closed |
| Completed | User acted |

## Cited By

- c3-115 (Onboarding Flow)
