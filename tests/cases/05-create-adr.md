# Test: Create Architecture Decision Record

## Setup

**Fixture:** `fixtures/05-create-adr/`

Pre-existing state (MULTI-CONTAINER + EXISTING ADR):
- `.c3/README.md` - Context with 3 containers:
  - c3-1 API Backend (REST, 40+ endpoints, OpenAPI docs)
  - c3-2 Gateway (routing, auth, versioning)
  - c3-3 Mobile BFF (aggregates c3-1 calls, has pain points)
  - Notes: different clients have different needs
- `.c3/c3-1-api/README.md` - REST API with Joi validation, OpenAPI spec
- `.c3/c3-2-gateway/README.md` - Handles versioning, partner APIs
- `.c3/c3-3-mobile-bff/README.md` - Documents pain points:
  - 5-10 REST calls per screen
  - Over-fetching, under-fetching
  - Complex aggregator workaround
  - Explicit request for GraphQL
- `.c3/adr/adr-20241115-api-versioning.md` - **EXISTING ADR** (implemented)
  - URL path versioning for partners
  - Affects c3-1 and c3-2

## Query

```
We decided to switch from REST to GraphQL for the API.
Document this as an ADR with the rationale and impact.
```

## Expect

### Multi-Container Analysis
- [ ] Identified all 3 containers affected differently
- [ ] Noted existing versioning ADR interaction
- [ ] Recognized different client needs (mobile vs partners)

---

## ADR Content Criteria

### PASS: Required Sections Present
| Section | Check |
|---------|-------|
| Context | WHY this decision was needed (mobile pain points) |
| Decision | WHAT was decided (GraphQL for some/all clients?) |
| Rationale | WHY GraphQL over alternatives (addressed over/under-fetching) |
| Consequences | Positive AND negative impacts |
| Changes Across Layers | What docs need updating (all 3 containers) |
| Verification Checklist | How to confirm implementation |

### PASS: Addressed Complexity
| Element | Check |
|---------|-------|
| Partner APIs | How does GraphQL affect versioning? (existing ADR) |
| Mobile BFF fate | Does c3-3 still need to exist? Simplified? |
| Gateway routing | How does c3-2 route GraphQL vs REST? |
| Phased approach? | All-at-once or gradual migration |

### FAIL: Missing Critical Content
| Element | Failure Reason |
|---------|----------------|
| No alternatives considered | ADR must show options weighed |
| No negative consequences | Every decision has tradeoffs |
| No verification items | Can't confirm decision applied |
| Ignored existing ADR | Versioning decision may conflict/interact |

---

## Layer Impact Assessment Criteria

### PASS: Correct Layer Attribution
| Layer | Should Identify |
|-------|--------------------|
| Context (c3-0) | Protocol change in interactions diagram |
| c3-1 (API) | New GraphQL schema/resolvers, OpenAPI fate |
| c3-2 (Gateway) | GraphQL routing, versioning strategy |
| c3-3 (Mobile BFF) | Simplification or deprecation |

### FAIL: Wrong Layer Attribution
| Element | Failure Reason |
|---------|----------------|
| "Context needs resolver details" | Resolvers are Component level |
| "Container handles schema design" | Schema patterns may be Component level |
| "Ignored c3-3 changes" | Major impact on BFF |

---

## Cross-Reference Criteria

### PASS: Proper References
- [ ] References all affected containers by C3 ID
- [ ] References existing versioning ADR
- [ ] Does NOT duplicate content from other docs
- [ ] Points to docs that need updating

### FAIL: Content Duplication
| Element | Failure Reason |
|---------|----------------|
| Re-explains mobile BFF pain points | Already documented in c3-3 |
| Copies versioning decision | Reference adr-20241115, don't duplicate |

---

## Verification Checklist Criteria

### PASS: Actionable Verification
| Element | Check |
|---------|-------|
| Specific checks | "GraphQL endpoint at /graphql responds" |
| Per-container items | Checks for c3-1, c3-2, c3-3 changes |
| Existing ADR compat | "Partner versioning still works" |
| Testable items | "Mobile app uses single GraphQL query per screen" |

### FAIL: Vague Verification
| Element | Failure Reason |
|---------|----------------|
| "GraphQL works" | Not specific enough |
| "API is updated" | No concrete check |
| "Tests pass" | Which tests? What validates? |

---

## ADR Lifecycle Criteria

### PASS: Correct Status
- [ ] Status is `proposed` (not `accepted` or `implemented`)
- [ ] Audit Record section included
- [ ] Lifecycle table with only Proposed marked

### FAIL: Wrong Status
| Element | Failure Reason |
|---------|----------------|
| Status: accepted | Only human can accept |
| Status: implemented | Requires audit pass first |
| No Audit Record | Required for tracking |
