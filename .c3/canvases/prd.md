---
id: prd
c3-seal: 4eb2041897b39cabef90751ff914a5f52b1f900eeeed94084c3ab671beaa27d5
type: canvas
status:
    - open
    - accepted
    - done
    - superseded
description: Product requirements document canvas with cite-backed facts and story traces.
---

domain: product
sections:
    - name: Goal
      content_type: text
      required: true
      purpose: Product outcome
      free: true
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
