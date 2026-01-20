---
id: ref-ai-integration
title: AI Integration Patterns
---

# AI Integration Patterns

## Goal

Establish conventions for AI provider integration including model selection, prompt engineering, response handling, and citation formatting for consistent AI-powered features.

## Overview

Conventions for AI provider integration and prompt engineering.

## Provider Usage

### Model Selection

| Use Case | Model | Rationale |
|----------|-------|-----------|
| Chat completion | gpt-4o, claude-3.5-sonnet | Best quality |
| Embeddings | text-embedding-ada-002 | Cost-effective, consistent |
| Fast suggestions | gpt-4o-mini | Speed over depth |

### Rate Limiting

| Tier | Tokens/min | Strategy |
|------|------------|----------|
| Free | 10,000 | Queue + delay |
| Pro | 100,000 | Burst allowed |
| Team | 500,000 | Priority queue |

## Prompt Engineering

### Structure

| Section | Purpose |
|---------|---------|
| System | Role, constraints, format |
| Context | Relevant knowledge |
| Examples | Few-shot demonstrations |
| User | Current request |

### Best Practices

| Practice | Why |
|----------|-----|
| Clear role definition | Consistent persona |
| Explicit format requirements | Parseable output |
| Context before question | Relevance |
| Limit context to relevant | Token efficiency |

## Response Handling

### Streaming

| Event | Content |
|-------|---------|
| chunk | Token fragment |
| citation | Referenced concept ID |
| suggestion | Action recommendation |
| done | End of stream |

### Error Recovery

| Error | Recovery |
|-------|----------|
| Rate limited | Backoff, retry after |
| Timeout | Return partial, indicate |
| Invalid response | Retry with simpler prompt |
| Content filter | Return safe alternative |

## Citation Format

Format: `[c:concept-id]` in response text

| Behavior | Implementation |
|----------|----------------|
| Extract citations | Regex match in response |
| Validate existence | Check concept exists in canvas |
| Render as link | Frontend transforms to clickable |

## Cited By

- c3-113 (AI Chat Panel)
- c3-214 (Chat Handler)
- c3-501 (Provider Abstraction)
- c3-502 (Prompt Templates)
- c3-513 (Suggestion Engine)
