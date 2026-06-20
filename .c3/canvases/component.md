---
id: component
c3-seal: c00049adca2867b7fe539ff217f87fac5494f898c26f13b056be26a0b38a5a6a
type: canvas
description: 'Component: an owned unit of behavior inside a container.'
---

domain: software
sections:
    - name: Goal
      content_type: text
      required: true
      purpose: What this component exists to do
      min_words: 4
    - name: Parent Fit
      content_type: table
      required: true
      purpose: How this component fits top-down into the parent container
      columns:
        - name: Field
          type: text
        - name: Value
          type: text
      min_rows: 4
    - name: Purpose
      content_type: text
      required: true
      purpose: Concrete ownership and non-goals
      min_words: 12
    - name: Governance
      content_type: table
      required: true
      purpose: Refs, rules, ADRs, specs, and precedence governing this component
      columns:
        - name: Reference
          type: reference
          edge: uses
          targets:
            - ref
            - rule
        - name: Type
          type: enum
          values:
            - ref
            - rule
            - adr
            - spec
            - policy
            - example
            - N.A - <reason>
        - name: Governs
          type: text
        - name: Precedence
          type: text
        - name: Notes
          type: text
      min_rows: 1
    - name: Contract
      content_type: table
      required: true
      purpose: Behavior surfaces that downstream code/material must honor
      columns:
        - name: Surface
          type: text
        - name: Direction
          type: enum
          values:
            - IN
            - OUT
            - IN/OUT
            - N.A - <reason>
        - name: Contract
          type: text
        - name: Boundary
          type: text
        - name: Evidence
          type: evidence
      min_rows: 2
    - name: Derived Materials
      content_type: table
      required: true
      purpose: Code, config, tests, docs, prompts, or assets that must derive from this component
      columns:
        - name: Material
          type: text
        - name: Must derive from
          type: text
        - name: Allowed variance
          type: text
        - name: Evidence
          type: evidence
      min_rows: 1
    - name: Foundational Flow
      content_type: table
      required: false
      purpose: Preconditions, inputs, state, and shared dependencies
      columns:
        - name: Aspect
          type: text
        - name: Detail
          type: text
        - name: Reference
          type: reference
      min_rows: 4
    - name: Business Flow
      content_type: table
      required: false
      purpose: Business outcome, primary path, alternates, and failure behavior
      columns:
        - name: Aspect
          type: text
        - name: Detail
          type: text
        - name: Reference
          type: reference
      min_rows: 4
    - name: Change Safety
      content_type: table
      required: false
      purpose: Risks, triggers, detection, and verification required before done
      columns:
        - name: Risk
          type: text
        - name: Trigger
          type: text
        - name: Detection
          type: text
        - name: Required Verification
          type: evidence
      min_rows: 2
reject_if: []
workorder: ""
