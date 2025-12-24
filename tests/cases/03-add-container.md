# Test: Add New Container

## Setup

**Fixture:** `fixtures/03-add-container/`

Pre-existing state (CONFLICT SCENARIO):
- `.c3/README.md` - Context with:
  - c3-1 (API Backend) and c3-2 (Database)
  - Diagram shows API → SendGrid (external)
  - Notes mention c3-1 currently handles basic email
- `.c3/c3-1-api/README.md` - Container with:
  - c3-103 Email Notifier component (EXISTING notification capability)
  - SendGrid in tech stack
  - Notes: "Consider extracting to dedicated notification service"

## Query

```
We're adding a new notification service container.
It will handle email, SMS, and push notifications.
Uses Bull for job queues and integrates with SendGrid, Twilio, and Firebase.
```

## Expect

### Conflict Detection
- [ ] Noticed existing c3-103 Email Notifier in c3-1
- [ ] Noticed SendGrid already in c3-1 tech stack
- [ ] Identified migration/extraction decision needed

---

## Context Level (c3-0) Update Criteria

### PASS: Should Update
| Element | Check |
|---------|-------|
| New container in inventory | c3-3 (next available ID) with purpose |
| WHY it exists | "Centralized notification handling across channels" |
| Relationships updated | How c3-1 now calls c3-3 instead of direct SendGrid |
| Migration noted | c3-103 moving from c3-1 to c3-3, or coexistence explained |

### FAIL: Ignored Conflict
| Element | Failure Reason |
|---------|----------------|
| No mention of existing c3-103 | Critical - duplication without addressing |
| Two SendGrid integrations | Should consolidate or explain why separate |
| Diagram still shows API → SendGrid | Should update to API → Notifications → SendGrid |

---

## Container Level (c3-3) Criteria

### PASS: Should Include
| Element | Check |
|---------|-------|
| Component inventory | Queue processor, Email sender, SMS sender, Push sender |
| WHAT each component does | Role (not implementation) |
| Tech stack table | Bull, SendGrid, Twilio, Firebase with purposes |
| Foundational aspects | Queue processing, retry logic, rate limiting patterns |
| Component relationships | How queue feeds into channel senders |

### PASS: Handled Migration
| Element | Check |
|---------|-------|
| c3-103 fate addressed | Moved to c3-3 OR deprecated OR coexists with explanation |
| ID continuity | If moved, new ID in c3-3 (c3-301?) or kept as c3-103? |

### FAIL: Skipped Foundation
| Element | Failure Reason |
|---------|----------------|
| No queue processor component | Core to notification architecture |
| Listed only business senders | Missing foundational infrastructure |

---

## c3-1 Update Criteria

### PASS: Should Update
| Element | Check |
|---------|-------|
| Tech stack | Remove SendGrid if moved to c3-3 |
| Components | Mark c3-103 as migrated/deprecated OR remove |
| Internal diagram | Update to show c3-3 dependency |

### FAIL: Left c3-1 Stale
| Element | Failure Reason |
|---------|----------------|
| c3-103 still active | Duplication - email in both containers |
| SendGrid in both tech stacks | Unclear ownership |
