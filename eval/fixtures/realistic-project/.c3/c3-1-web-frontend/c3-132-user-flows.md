---
id: c3-132
c3-version: 3
title: User Flows
type: component
category: documentation
parent: c3-1
summary: Exhaustive user flows mapped to IA regions with preconditions and dependencies
---

# User Flows

Documents all user interactions as flows with entry points, steps, preconditions, and postconditions. Each flow references regions from c3-131 Information Architecture.

## Uses

| Category | Component | For |
|----------|-----------|-----|
| Documentation | c3-131 Information Architecture | Region IDs referenced in flows |
| Feature | c3-121 Invoice Screen | INV-* flow implementation |
| Feature | c3-122 Payment Requests Screen | PR-*, APR-* flow implementation |
| Feature | c3-124 Admin Screen | ADM-* flow implementation |
| Backend | c3-221 PR Flows | Server-side flow handlers |

## Conventions

| Rule | Why |
|------|-----|
| Prefix AUTH- | Authentication flows |
| Prefix NAV- | Navigation flows |
| Prefix INV- | Invoice domain flows |
| Prefix PR- | Payment request flows |
| Prefix APR- | Approval flows |
| Prefix PAY- | Payment method flows |
| Prefix ADM- | Admin flows |

## Flow Categories

| Category | Count | Entry Points |
|----------|-------|--------------|
| Authentication | 2 | SCR-LOGIN, GBL-SIDEBAR |
| Navigation | 3 | GBL-SIDEBAR |
| Invoices | 11 | SCR-INV, PNL-INV-DETAIL |
| Payment Requests | 12 | SCR-PR, PNL-PR-DETAIL |
| Approvals | 6 | SCR-APR, PNL-APR-DETAIL |
| Payment Methods | 3 | SCR-PAY, PNL-PAY-DETAIL |
| Admin | 4 | SCR-ADMIN |

## Key Flow Chains

| Chain | Flows | Description |
|-------|-------|-------------|
| Invoice-to-Payment | INV-001→INV-003→INV-005→PR-004→APR-001→PR-006 | Full invoice payment lifecycle |
| Bulk Operations | INV-009→INV-010/INV-011 | Batch invoice status changes |
| Approval Batch | APR-004→APR-005 | Multi-PR approval |

## State Machines

| Entity | States | Transitions |
|--------|--------|-------------|
| Invoice | imported, inprogress, obsolete | INV-005, INV-006, INV-007, INV-008 |
| Payment Request | draft, requested, approved, rejected, complete | PR-003, PR-004, PR-005, PR-006, PR-007, APR-001, APR-002 |

## Testing

| Scenario | Verifies |
|----------|----------|
| Flow completeness | All UI actions have documented flows |
| Precondition enforcement | Flows respect documented preconditions |
| State transitions | Flows produce documented postconditions |
| Region coverage | All regions appear in at least one flow |

## References

- `apps/start/src/screens/` - Flow implementations
- `apps/server/src/flows/` - Backend flow handlers
- [c3-131 Information Architecture](./c3-131-information-architecture.md) - Region definitions
- [c3-133 UI Patterns](./c3-133-ui-patterns.md) - Pattern usage in flows
