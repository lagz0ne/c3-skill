---
id: adr-20260202-notification-on-step-advance
c3-seal: e8a3c61686b7c60bbc23bf40f8e7f8503e5fbc57bcda988f77ce5c8eb77aa0c8
title: Notify Approvers When Approval Step Advances
type: adr
goal: 'Document and implement the architectural decision: Notify Approvers When Approval Step Advances.'
status: implemented
date: "2026-02-02"
affects:
    - c3-205
    - c3-211
approved-files:
    - apps/start/src/server/flows/pr.ts
    - apps/start/src/server/services/pr.ts
---

# Notify Approvers When Approval Step Advances

## Goal

Document and implement the architectural decision: Notify Approvers When Approval Step Advances.

## Status

**Implemented** - 2026-02-02

## Problem

Users who are in step N+1 of an approval chain never receive notifications when a PR advances from step N to step N+1.

**Current behavior:**

1. `requestApprovals` flow calls `notificationService.notifyNextApprovers` -> Step 0 users notified
2. `approvePr` flow calls `prService.approve` which advances `current_step` but does NOT notify -> Step 1+ users NOT notified
3. `approveAll` flow has the same issue

**Example:** If duke@silentium.io is in step 2 of the approval chain, they never receive a notification when the PR advances from step 1 to step 2.

**Evidence:** `prService.approve` (lines 386-396) advances the step when `isProgressed && !willEnd` but this information is not returned to the calling flow, so the flow cannot know to trigger notifications.

## Decision

### 1. Modify prService.approve to return step advancement info

Add `stepAdvanced: boolean` to the return type to explicitly signal when a step advanced (but PR is not yet fully approved).

```typescript
return {
  success: true as const,
  pr: updatedPr,
  stepAdvanced: isProgressed && !willEnd  // NEW: explicit signal
};
```

### 2. Modify approvePr and approveAll flows to notify on step advance

```typescript
// In approvePr flow
const result = await ctx.exec({ fn: prService.approve, params: [ctx.input.prId, currentUser.email] });

if (result.success && result.stepAdvanced) {
  await ctx.exec({ fn: notificationService.notifyNextApprovers, params: [ctx.input.prId] });
  logger.info({ prId: ctx.input.prId }, "notified next approvers after step advance");
}
```

Similar change in `approveAll` for each PR that has `stepAdvanced: true`.

## Rationale

| Considered | Rejected Because |
| --- | --- |
| Infer from next_approver presence | Fragile - next_approver exists even when step didn't advance (multiple approvers in same step) |
| Notify in prService.approve | Services should not trigger side effects; flows orchestrate per C3 pattern |
| Notify on every approval | Wasteful - only notify when step actually advances |
| Notify when PR fully approved | No action required - PR is done |

**Why explicit `stepAdvanced` flag:**

- The service already computes `isProgressed` and `willEnd`
- Returning this info avoids fragile inference in the flow
- Clear contract: "step advanced, you should notify"

**Concurrency/Idempotency:**

- If two approvals race and both trigger notifications, `notificationPublisher.publish` is idempotent (JetStream dedupes by message ID)
- Notification log tracks delivery attempts for debugging

## Affected Layers

| Layer | Document | Change |
| --- | --- | --- |
| c3-205 | PR Flows | Add notification call in approvePr and approveAll |
| c3-211 | Notification System | No change - already supports this use case |
| c3-2 | API Backend | Service return type change (backward compatible) |

## Step-Advancing Paths Audit

| Path | Advances Step? | Needs Notification? |
| --- | --- | --- |
| approvePr | Yes | Yes - fixing in this ADR |
| approveAll | Yes (calls prService.approve) | Yes - fixing in this ADR |
| requestApprovals | No (sets to step 0) | Already notifies |
| rejectPr / recallPr | Reverts to draft | No - approval chain reset |
| Admin override | Not implemented | N/A |

## Verification

- [ ] Create PR with 3-step approval chain (user A step 0, user B step 1, user C step 2)
- [ ] Request approvals as maker -> user A receives notification
- [ ] Approve as user A (step advances 0->1) -> user B receives notification
- [ ] Approve as user B (step advances 1->2) -> user C receives notification
- [ ] Approve as user C (PR approved) -> no notification sent
- [ ] Test `approveAll` with multiple PRs at different steps -> correct notifications sent
- [ ] Check notification_log for expected entries

## Context

N.A - historical ADR; original context is captured in the git log around the ADR date and in the current code that implements the decision.

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| N.A - historical | Shipped via git commits; the c3 topology and code-map reflect the resulting structure. | c3x list --include-adr and git log around the ADR date |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| N.A - historical | Current .c3 entities, refs, and code-map are the post-change state. | c3x verify and c3x check |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| N.A - historical | Enforcement is implicit in the currently linked components and refs. | c3x graph and cited ref ids on the relevant components |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| N.A - historical | Alternatives were considered at decision time; rationale is preserved in the original commit message or branch discussion. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| N.A - historical | Risks were assessed pre-merge; the decision has since shipped without outstanding incidents tied to this ADR. | git log and project test suite |
