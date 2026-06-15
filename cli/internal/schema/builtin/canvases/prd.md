---
id: prd
c3-seal: 4eb2041897b39cabef90751ff914a5f52b1f900eeeed94084c3ab671beaa27d5
type: canvas
description: Product requirements document canvas with cite-backed facts and story traces.
status: [open, accepted, done, superseded]
---

domain: product
sections:
    - name: Goal
      content_type: text
      required: true
      free: true
      purpose: Product outcome
    - name: Requirements
      content_type: table
      required: true
      purpose: Release requirements and source evidence
      columns:
        - name: Requirement
          type: text
        - name: Priority
          type: enum
          values:
            - must
            - should
            - could
            - wont
        - name: Evidence
          type: cite
    - name: Story Traces
      content_type: table
      required: true
      purpose: Stories derived from requirements
      columns:
        - name: Story
          type: edge<requirement|story>
        - name: Status
          type: check
        - name: Evidence
          type: cite
reject_if:
    - Requirement lacks source evidence
    - Story trace is missing
workorder: ""
