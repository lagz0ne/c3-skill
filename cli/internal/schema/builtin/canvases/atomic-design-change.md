---
id: atomic-design-change
c3-seal: 55db56c9945a1967ca8e971318337a4944598b7fdb0aa22dcb909d22295be567
type: canvas
description: Track design-system changes from atom through page with cite-backed impact.
---

domain: design
sections:
    - name: Goal
      content_type: text
      required: true
      purpose: Design-system change objective
    - name: Affected Units
      content_type: table
      required: true
      purpose: Atomic design units touched by the change
      columns:
        - name: Unit
          type: text
        - name: Level
          type: enum
          values:
            - atom
            - molecule
            - organism
            - template
            - page
            - N.A - <reason>
        - name: Why affected
          type: text
        - name: Evidence
          type: cite
    - name: Change Record
      content_type: table
      required: true
      purpose: Specific design deltas and verification state
      columns:
        - name: Change
          type: text
        - name: Break risk
          type: text
        - name: Result
          type: check
        - name: Evidence
          type: cite
reject_if:
    - Affected design units lack cite-backed evidence
    - Change Record has no check result
workorder: Read the referenced design-system docs first; use N.A - <reason> only for truly absent units.
