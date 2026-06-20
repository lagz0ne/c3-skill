---
id: pm-requirement
c3-seal: ec6b85e1175987a5871437205c63ea0c895b0e00a023cbc1cd23b63c4c94347a
type: canvas
description: Requirement canvas with source facts, acceptance checks, and trace edges.
---

domain: product
sections:
    - name: Need
      content_type: text
      required: true
      purpose: User or business need being captured
    - name: Facts
      content_type: table
      required: true
      purpose: Current product facts that constrain the requirement
      columns:
        - name: Fact
          type: text
        - name: Evidence
          type: cite
    - name: Acceptance
      content_type: table
      required: true
      purpose: Verifiable acceptance checks
      columns:
        - name: Scenario
          type: text
        - name: Result
          type: check
        - name: Trace
          type: edge<fact|prd|story>
reject_if:
    - Facts are uncited
    - Acceptance checks cannot be verified
workorder: ""
