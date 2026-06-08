---
id: user-story
c3-seal: a5c8a03c60d5afa46a75f82a75d10a8d17fa926e6e2e8785fda203c62fd0b2bd
type: canvas
description: User story canvas with role, need, acceptance, and cite-backed derivation.
---

domain: product
sections:
    - name: Story
      content_type: text
      required: true
      purpose: As-a/I-want/so-that statement
    - name: Acceptance
      content_type: table
      required: true
      purpose: Acceptance criteria with check state
      columns:
        - name: Criterion
          type: text
        - name: Result
          type: check
        - name: Evidence
          type: cite
    - name: Trace
      content_type: table
      required: true
      purpose: Requirement and PRD ancestry
      columns:
        - name: Source
          type: edge<prd|requirement>
        - name: Why derived
          type: text
        - name: Evidence
          type: cite
reject_if:
    - Story has no cited source
    - Acceptance cannot be checked
workorder: ""
