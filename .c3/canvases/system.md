---
id: system
c3-seal: c7ee5e1843104fd517b45198c15ddeb080bf79275fc4c38e9f5af1d3f6583244
type: canvas
description: 'System context: top-level objective, containers, and system-wide constraints.'
---

domain: software
sections:
    - name: Goal
      content_type: text
      required: true
      purpose: System-level objective
    - name: Containers
      content_type: table
      required: true
      purpose: Top-level deployment units
      columns:
        - name: ID
          type: entity_id
        - name: Name
          type: text
        - name: Boundary
          type: text
        - name: Status
          type: text
        - name: Responsibilities
          type: text
        - name: Goal Contribution
          type: text
    - name: Abstract Constraints
      content_type: table
      required: true
      purpose: System-wide architectural rules
      columns:
        - name: Constraint
          type: text
        - name: Rationale
          type: text
        - name: Affected Containers
          type: text
reject_if: []
workorder: ""
