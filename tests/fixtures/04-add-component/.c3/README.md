---
id: c3-0
c3-version: 3
title: Test API System
summary: A test system with a single API container
---

# Test API System

## Overview

A simple API backend for testing component documentation.

## Containers

| ID | Name | Purpose |
|----|------|---------|
| c3-1 | API Backend | Handles REST requests with caching |

## Container Interactions

```mermaid
flowchart LR
    Client((Client)) --> API[c3-1 API Backend]
    API --> Redis[(Redis)]
```

## External Actors

- API clients (web, mobile)
