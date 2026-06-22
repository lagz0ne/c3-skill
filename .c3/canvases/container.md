---
id: container
c3-seal: f6e570dd56744c331669f112a581353865e26efadd2958c87071fccdec341d23
type: canvas
description: 'Container: a deployable/process unit and the components it owns.'
---

domain: software
sections:
    - name: Goal
      content_type: text
      required: true
      purpose: What this container exists to do
    - name: Components
      content_type: table
      required: true
      purpose: Parts that compose this container
      columns:
        - name: ID
          type: entity_id
        - name: Name
          type: text
        - name: Category
          type: text
        - name: Status
          type: text
        - name: Goal Contribution
          type: text
    - name: Responsibilities
      content_type: text
      required: true
      purpose: What this container is accountable for
    - name: Complexity Assessment
      content_type: text
      required: false
      purpose: Known complexity and risks
reject_if: []
workorder: ""
