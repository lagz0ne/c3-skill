---
id: prd
c3-seal: bad862bccd901408ddd4cd4cc32c2d6ca8681d647214c4bec774f1e4982525ce
type: canvas
description: Product requirements document canvas with cite-backed facts and story traces.
---

domain: product
sections:
    - name: Goal
      content_type: text
      required: true
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
