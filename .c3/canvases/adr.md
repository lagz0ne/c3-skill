---
id: adr
c3-seal: ae29add672f943d9032373d42c3cc1af0f44767c9887e995a3697285c7ce1a9c
type: canvas
status:
    - open
    - accepted
    - done
    - superseded
description: Decision record — lean required core (Goal, Context, Decision, Affected Topology, Verification); optional work-order sections (governance, execution, alternatives, risks) climb in for weightier decisions.
---

domain: software
sections:
    - name: Goal
      content_type: text
      required: true
      purpose: Decision context and objective
      fill: State the exact change objective in one concrete paragraph. Name the system behavior or architecture decision being changed, not just the ticket title.
      failure: If this is vague, the ADR can pass mechanically but nobody can tell what decision it is actually authorizing.
      free: true
    - name: Context
      content_type: text
      required: true
      purpose: Current behavior, user pain, constraints, and affected topology
      fill: Describe the current state, the problem or pressure forcing the change, the constraints, and the part of the topology involved.
      failure: If this is thin, later readers cannot tell whether the ADR solved the real problem or introduced drift against current architecture.
      free: true
    - name: Decision
      content_type: text
      required: true
      purpose: Concrete selected approach and why it is the right fit
      fill: Write the chosen approach and why it wins over the realistic alternatives for this repo, branch, or architecture shape.
      failure: If this is hand-wavy, implementation can branch into multiple interpretations and the ADR stops being a work order.
      free: true
    - name: Affected Topology
      content_type: table
      required: true
      purpose: Components or containers this ADR changes, plus the governance review expected for each
      fill: List every system/container/component touched by the decision, why it is affected, cite the current C3 node proving it, and what governance review must happen there.
      failure: If this is incomplete, c3x cannot derive the refs/rules that must be reviewed or complied with, so ADR coverage drifts silently.
      columns:
        - name: Entity
          type: text
        - name: Type
          type: enum
          values:
            - system
            - container
            - component
            - N.A - <reason>
        - name: Why affected
          type: text
        - name: Evidence
          type: cite
        - name: Governance review
          type: text
    - name: Compliance Refs
      content_type: table
      required: false
      purpose: Existing or to-be-created refs that the affected topology must review or comply with
      fill: 'For each governing ref, name the ref, explain why it applies to this ADR, cite the current ref node proving it, and record the action: comply, review, create-ref, update-ref, or N.A with reason.'
      failure: If this is vague or missing, the model will under-mention governing references and the ADR will miss architecture constraints it was supposed to respect.
      columns:
        - name: Ref
          type: text
        - name: Why required
          type: text
        - name: Evidence
          type: cite
        - name: Action
          type: text
    - name: Compliance Rules
      content_type: table
      required: false
      purpose: Existing or to-be-created rules that the affected topology must review or comply with
      fill: For each governing rule, name the rule, explain why it applies, cite the current rule node proving it, and say whether the work must comply, needs review, or must create/update the rule.
      failure: If this is vague or missing, rule enforcement becomes implicit again and downstream code can violate golden patterns without being called out in the ADR.
      columns:
        - name: Rule
          type: text
        - name: Why required
          type: text
        - name: Evidence
          type: cite
        - name: Action
          type: text
    - name: Work Breakdown
      content_type: table
      required: false
      purpose: Files, docs, commands, or entities to change and how each maps to the decision
      fill: Name the concrete implementation/doc work items and tie each one back to the decision. Prefer files, commands, entities, or scopes over vague task labels.
      failure: If this is generic, another agent cannot recover execution steps from the ADR alone and work will depend on chat history.
      columns:
        - name: Area
          type: text
        - name: Detail
          type: text
        - name: Evidence
          type: text
    - name: Underlay C3 Changes
      content_type: table
      required: false
      purpose: C3 CLI files, validators, commands, hints, help, schemas, templates, or tests changed by this decision
      fill: 'List exact C3 underlay surfaces changed by this ADR: commands, validators, tests, schema rows, hints, templates, docs, and the proof that each was updated.'
      failure: If this is weak, C3-facing changes ship without their enforcing validator/help/test surface and the documented contract drifts from the actual CLI.
      columns:
        - name: Underlay area
          type: text
        - name: Exact C3 change
          type: text
        - name: Verification evidence
          type: text
    - name: Enforcement Surfaces
      content_type: table
      required: false
      purpose: Commands, validators, tests, docs, or runtime paths that enforce the decision
      fill: 'Name every place that will catch drift: commands, runtime checks, tests, docs, guardrails, or validators.'
      failure: If this is missing, the ADR describes intent but gives no proof path, so regressions become opinion-driven instead of mechanically catchable.
      columns:
        - name: Surface
          type: text
        - name: Behavior
          type: text
        - name: Evidence
          type: text
    - name: Alternatives Considered
      content_type: table
      required: false
      purpose: Real options rejected and why
      fill: List the real competing approaches and the repo-specific reason each was rejected.
      failure: If this is fake or generic, the ADR gives no decision pressure and future readers will reopen already-rejected paths.
      columns:
        - name: Alternative
          type: text
        - name: Rejected because
          type: text
    - name: Risks
      content_type: table
      required: false
      purpose: Failure modes, mitigations, and verification
      fill: Name concrete failure modes introduced by the decision, how they are mitigated, and how the mitigation will be verified.
      failure: If this stays soft, the ADR will approve risky work without naming how failure would show up or be contained.
      columns:
        - name: Risk
          type: text
        - name: Mitigation
          type: text
        - name: Verification
          type: text
    - name: Verification
      content_type: table
      required: true
      purpose: Exact commands or evidence required before marking the ADR implemented
      fill: Write exact commands, smoke checks, or artifacts required before calling the ADR implemented. Prefer executable proof over prose promises.
      failure: If this is vague, the work can be marked done without proof and the ADR stops enforcing the project's verify-before-done rule.
      columns:
        - name: Check
          type: text
        - name: Result
          type: text
reject_if:
    - Any required section absent or filled with TBD/TODO/"see above"/"as needed"
    - Compliance rows must say why the ref/rule applies, unless the whole row is N.A - <reason>
    - Affected Topology rows must say why the entity is affected, unless the whole row is N.A - <reason>
    - Verification has no executable command, smoke check, or named artifact
    - Alternatives Considered rows have no repo-specific rejection reason
    - Underlay C3 Changes lacks the exact validators/tests/help that enforce the decision
workorder: |-
    Run c3x schema adr before drafting; do not draft ADR prose first and reconcile later.
    Before the ADR body, make a volatile Discovery Brief from the task goal and targeted c3x reads: owner, governing material, stop condition.
    Treat each 'fill' line as required authoring guidance, not optional commentary.
    Required core: Goal, Context, Decision, Affected Topology, Verification — a small change needs only these.
    The work-order sections (Compliance Refs/Rules, Work Breakdown, Underlay C3 Changes, Enforcement Surfaces, Alternatives, Risks) are optional — include them for weightier decisions; any you DO include must be substantive (thin included sections fail).
