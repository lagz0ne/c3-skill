---
id: ref-approval-chain
c3-seal: c857a241401917c3aa3f18c6c06d4e03264a2212e24ff6cf6ed665ae3f7eba1b
title: Approval Chain Pattern
type: ref
goal: Multi-step approval chain for payment requests. Each step has assigned approvers and a mode (`anyof` = first approver advances, `allof` = every approver must sign).
---

# Approval Chain Pattern

## Goal

Multi-step approval chain for payment requests. Each step has assigned approvers and a mode (`anyof` = first approver advances, `allof` = every approver must sign).

## Choice

Represent approval workflows as explicit step graphs persisted in relational tables (`approval`, `approval_steps`, `approval_step_users`, `approval_records`) and drive progression in application services.

## Why

This keeps authorization and progression logic explicit, auditable, and compatible with mixed approval semantics (`anyof`/`allof`) without embedding workflow logic in ad-hoc UI logic.

## Happy Path

```
createPr (with approval config)
  → PR created in draft, approval row + steps + step_users written

requestApprovals
  → PR status: draft → pending, current_step stays at 0

approve (step 0, mode: anyof)
  → approval_record inserted
  → anyof satisfied → current_step: 0 → 1

approve (step 1, mode: allof, user A)
  → approval_record inserted
  → allof NOT satisfied (user B hasn't approved) → current_step stays at 1

approve (step 1, mode: allof, user B)
  → approval_record inserted
  → allof satisfied → current_step: 1 → 2
  → current_step === total steps → PR status: pending → approved

complete
  → PR status: approved → completed, linked invoices marked complete
```

## Wiring

```
flow (pr.ts)           → prService              → approvalQueries
─────────────────────────────────────────────────────────────────
approvePr.flow         → prService.approve       → getApprovalByPrId
                                                 → getApproversForStep
                                                 → newApprovalRecord
                                                 → updateApprovalCurrentStep
                                                 → (prQueries.markPrAsApproved if final)
                       → notificationService.notifyNextApprovers (if stepAdvanced)

requestApprovals.flow  → prService.requestForApprovals → prQueries.markPrAsRequestForApprovals
                       → notificationService.notifyNextApprovers

unapprovePr.flow       → prService.unapprove     → getApproversForStep (previous step)
                                                 → deleteApprovalRecord
                                                 → updateApprovalCurrentStep (revert if needed)

recallPr.flow          → prService.revertFromRequestingForApprovals
rejectPr.flow            → markPrAsDraft + reset current_step to 0 + clear all records
```

## State Machine

```
[*] → draft → pending → pending (approve, step incomplete)
                       → approved (approve, all steps done)
                       → draft (recall/reject — resets step to 0, clears records)
approved → completed
completed → approved (uncomplete)
```

## Data Model

```sql
-- 4 tables, PR is the anchor
approval         (id, pr_id, current_step)           -- 1:1 with PR
approval_steps   (id, approval_id, step_number, name, mode)  -- ordered steps
approval_step_users (step_id, user_email)             -- assigned approvers per step
approval_records (id, approval_id, step_number, user_email, approved_at) -- who approved
```

Steps are 0-indexed. `current_step` points to the step awaiting approval. When `current_step >= total steps`, chain is complete.

## Mode Validation (App-Level)

Modes (`anyof`/`allof`) are **app-level logic in `prService.approve`**, not DB constraints. The DB stores the mode string; the service reads it and decides whether to advance:

```typescript
// prService.approve — core mode check
const existingApprovers = await approvalQueries.getApproversForStep(ctx, {
  approval_id: approval.id,
  step_number: currentStep,
});
const currentApprovers = new Set(existingApprovers.map(a => a.user_email));
const usersToCompare = new Set(currentApprovalStep.users);

// anyof: one approver is enough
const matchedAnyOf = currentApprovalStep.mode === "anyof"
  && usersToCompare.has(approver);

// allof: every assigned user must have a record
currentApprovers.add(approver);
const matchedAllOf = currentApprovalStep.mode === "allof"
  && [...usersToCompare].every(u => currentApprovers.has(u));

const isProgressed = matchedAnyOf || matchedAllOf;
```

## Golden Example: Approve + Step Advance

```typescript
// Record approval, then check if step should advance
await approvalQueries.newApprovalRecord(ctx, {
  approval_id: approval.id,
  step_number: currentStep,
  user_email: approver,
});

if (isProgressed) {
  const willEnd = currentStep === approvalFlow.length - 1;
  if (willEnd) {
    await prQueries.markPrAsApproved({ id: prId });
  }
  await approvalQueries.updateApprovalCurrentStep({
    pr_id: prId,
    current_step: currentStep + 1,
  });
}
```

## Golden Example: Create PR with Approval Chain

```typescript
// 1. Create PR
const pr = await prQueries.newPr({ ...prData, maker: user.email });

// 2. Create approval anchor
const approval = await approvalQueries.newApproval({ pr_id: pr.id });

// 3. Write steps + users
for (let i = 0; i < approvalConfig.length; i++) {
  const step = approvalConfig[i];
  const createdStep = await approvalQueries.newApprovalStep({
    approval_id: approval.id,
    step_number: i,
    name: step.name,
    mode: step.mode,
  });
  for (const userEmail of step.users) {
    await approvalQueries.newApprovalStepUser({
      step_id: createdStep.id,
      user_email: userEmail,
    });
  }
}
```

## Shared Types

```typescript
// @acountee/shared — ApprovalFlow schema
const approvalStepSchema = z.object({
  name: z.string(),
  mode: z.enum(["anyof", "allof"]),
  users: z.array(z.string()),
});
const approvalFlowSchema = z.array(approvalStepSchema);
type ApprovalStep = z.infer<typeof approvalStepSchema>;
type ApprovalFlow = z.infer<typeof approvalFlowSchema>;
```

## Edge Cases

| Scenario | Behavior |
| --- | --- |
| Unapprove | Only works on currentStep - 1. Deletes record. If allof mode or no approvers remain, step reverts |
| Recall/Reject | Resets to draft, current_step to 0, clears all approval records (from step 0+) |
| Update approval config | Only allowed when PR is in draft. Deletes all existing steps, recreates from new config |
| Concurrent approvals | Transaction-scoped via execution context |

## Cited By

- `c3-205-pr-flows.md` — PR flow orchestration
